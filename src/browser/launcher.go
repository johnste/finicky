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
	_ "embed"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"unsafe"
)

//go:embed browsers.json
var browsersJsonData []byte

type BrowserConfig struct {
	Name             string   `json:"name"`
	AppType          string   `json:"appType"`
	OpenInBackground bool     `json:"openInBackground"`
	Profile          string   `json:"profile"`
	Args             []string `json:"args"`
	URL              string   `json:"url"`
}

type browserInfo struct {
	ConfigDirRelative string `json:"config_dir_relative"`
	ID               string `json:"id"`
	Type             string `json:"type"`
}

func LaunchBrowser(config BrowserConfig) error {
	appName := config.Name
	useBundleId := true
	cName := C.CString(config.Name)
	defer C.free(unsafe.Pointer(cName))

	if config.AppType == "none" {
		log.Printf("AppType is 'none', not launching any browser")
		return nil
	}

	log.Printf("Starting browser %s for URL %s", config.Name, config.URL)

	if config.AppType == "bundleId" {
		log.Printf("Using bundle id for %s", config.Name)
	} else if config.AppType == "appName" {
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

	openArgs := []string{"-a", appName}
	if useBundleId {
		openArgs = []string{"-b", appName}
	}

	if config.OpenInBackground {
		openArgs = append(openArgs, "-g")
	}

	if len(config.Args) == 0 {
		if useBundleId {
			profileArgument, ok := resolveBrowserProfileArgument(appName, config.Profile)
			if ok {
				// FIXME: This is a hack to get the profile argument to work – this won't work for Firefox
				openArgs = append(openArgs, "-n")
				openArgs = append(openArgs, "--args")
				openArgs = append(openArgs, profileArgument)
			}
		}

		openArgs = append(openArgs, config.URL)
	} else {
		openArgs = append(openArgs, config.Args...)
	}

	cmd := exec.Command("open", openArgs...)

	log.Printf("Executing command: %v", cmd)
	return cmd.Start()
}

func resolveBrowserProfileArgument(bundleId string, profile string) (string, bool) {
	var browsersJson []browserInfo
	if err := json.Unmarshal(browsersJsonData, &browsersJson); err != nil {
		log.Printf("Error parsing browsers.json: %v", err)
		return "", false
	}

	// Try to find matching browser by bundle ID
	var matchedBrowser *browserInfo
	for _, browser := range browsersJson {
		if browser.ID == bundleId {
			matchedBrowser = &browser
			break
		}
	}

	if matchedBrowser == nil {
		log.Printf("Browser %s not found in browsers.json", bundleId)
		return "", false
	}

	log.Printf("Browser %s found in browsers.json (type: %s)", bundleId, matchedBrowser.Type)

	if profile != ""  {
		switch matchedBrowser.Type {
		case "Chromium":
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Printf("Error getting home directory: %v", err)
				return "", false
			}

			localStatePath := filepath.Join(homeDir, "Library/Application Support", matchedBrowser.ConfigDirRelative, "Local State")
			profilePath, ok := parseProfiles(localStatePath, profile)
			if ok {
				return "--profile-directory=" + profilePath, true
			}
		default:
			log.Printf("Browser %s is not a Chromium browser, skipping profile detection", bundleId)
		}
	}

	return "", true
}

func parseProfiles(localStatePath string, profile string) (string, bool) {
	data, err := os.ReadFile(localStatePath)
	if err != nil {
		log.Printf("Error reading Local State file at %s: %v", localStatePath, err)
		return "", false
	}

	var localState map[string]interface{}
	if err := json.Unmarshal(data, &localState); err != nil {
		log.Printf("Error parsing Local State JSON: %v", err)
		return "", false
	}

	profiles, ok := localState["profile"].(map[string]interface{})
	if !ok {
		log.Printf("Could not find profile section in Local State")
		return "", false
	}

	infoCache, ok := profiles["info_cache"].(map[string]interface{})
	if !ok {
		log.Printf("Could not find info_cache in profile section")
		return "", false
	}



	// Look for the specified profile
	for profilePath, info := range infoCache {
		profileInfo, ok := info.(map[string]interface{})
		if !ok {
			continue
		}

		name, ok := profileInfo["name"].(string)
		if !ok {
			continue
		}

		if name == profile {
			log.Printf("Found profile '%s' at path '%s'", name, profilePath)
			return profilePath, true
		}
	}

	// If we didn't find the profile, try to find it by profile folder name

	log.Printf("Could not find profile '%s' in browser profiles, trying to find by profile path...", profile)
	for profilePath, info := range infoCache {
		if profilePath == profile {
			// Try to get the profile name of the profile we want the user to use instead
			if profileInfo, ok := info.(map[string]interface{}); ok {
				if name, ok := profileInfo["name"].(string); ok {
					log.Printf("Found profile at path '%s' using profile path. Please use the profile name instead: %s", profilePath, name)
				}
			}
			return profilePath, true
		}
	}

	log.Printf("Could not find profile '%s' in browser profiles:", profile)
	log.Println("Available profiles:")
	for _, info := range infoCache {
		profileInfo, ok := info.(map[string]interface{})
		if !ok {
			continue
		}

		name, ok := profileInfo["name"].(string)
		if !ok {
			continue
		}

		log.Printf("- %s", name)
	}

	return "", false
}
