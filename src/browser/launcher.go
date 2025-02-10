package browser

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
const char* resolveBundleId(const char* appName);
const char* resolveAppPath(const char* appName);
*/
import "C"
import (
	"log"
	"os/exec"
	"unsafe"
)

type BrowserConfig struct {
	Name             string   `json:"name"`
	AppType          string   `json:"appType"`
	OpenInBackground bool     `json:"openInBackground"`
	Profile          string   `json:"profile"`
	Args             []string `json:"args"`
	URL              string   `json:"url"`
}


func LaunchBrowser(config BrowserConfig) error {
	log.Printf("Starting browser %s for URL %s", config.Name, config.URL)

	appName := config.Name	
	useBundleId := true	
	cName := C.CString(config.Name)
	defer C.free(unsafe.Pointer(cName))

	if config.AppType == "none" {
		log.Printf("AppType is 'none', not launching any browser")
		return nil		
	}
		
	if config.AppType == "bundleId" {
		log.Printf("Using bundle id for %s", config.Name)
	} else if config.AppType == "name" {
		if cBundleId := C.resolveBundleId(cName); cBundleId != nil {
			defer C.free(unsafe.Pointer(cBundleId))
			bundleId := C.GoString(cBundleId)
			log.Printf("Resolved bundle id for %s: %s", config.Name, bundleId)
			appName = bundleId
		} 
	} else if config.AppType == "appPath" {
		if cPath := C.resolveAppPath(cName); cPath != nil {
			defer C.free(unsafe.Pointer(cPath))
			path := C.GoString(cPath)
			log.Printf("Using app path for %s: %s", config.Name, path)
			appName = path
			useBundleId = false
		} else {
			log.Printf("Could not resolve app %s, using name as-is", config.Name)
		}
	}

	if useBundleId {
		detectBrowserProfile(appName, config.Profile)
	}

	// args := append([]string{}, config.Args...)
	// if config.Profile != "" {
	// 	args = append(args, fmt.Sprintf("--profile=%s", config.Profile))
	// }

	openArgs := []string{"-a", appName}
	if useBundleId {
		openArgs = []string{"-b", appName}
	}

	if config.OpenInBackground {
		openArgs = append(openArgs, "-g")
	}

	if len(config.Args) == 0 {		
		openArgs = append(openArgs, config.URL)
	} else {		
		openArgs = append(openArgs, config.Args...)
	}

	cmd := exec.Command("open", openArgs...)

	log.Printf("Executing command: %v", cmd)
	return cmd.Start()
}



func detectBrowserProfile(appName string, profile string) bool {
	// Check if this is a known Chromium browser from browsers.json
	browsersJson := []struct {
		ConfigDirRelative string `json:"config_dir_relative"`
		ID               string `json:"id"`
		Type            string `json:"type"` 
	}{}

	// Try to find matching browser by bundle ID
	found := false
	for _, browser := range browsersJson {
		if browser.ID == appName {
			found = true
			break
		}
	}

	if found && profile != "" {
		log.Printf("Browser %s found in browsers.json, try to use profile %s", appName, profile)
	}

	if !found {
		log.Printf("Browser %s not found in browsers.json", appName)
	}
	
	return found
}