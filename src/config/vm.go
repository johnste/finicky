package config

import (
	"embed"
	"finicky/logger"
	"finicky/util"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
)

// VM represents the JavaScript VM configuration
type VM struct {
	runtime *goja.Runtime
	namespace string
}
// New creates and configures a new VM instance
func New(embeddedFiles embed.FS) (*VM, error) {
	vm := &VM{
		runtime: goja.New(),
		namespace: "finickyConfig",
	}

	defer vm.setupLogging()

	bundlePath, err := vm.prepareConfig()

	if err != nil {
		return nil, fmt.Errorf("failed to bundle config: %v", err)
	}

	err = vm.setup(embeddedFiles, bundlePath)
	if  err != nil {
		return nil, err
	}

	return vm, err
}

func (vm *VM) setupLogging() {
	// Setup logging based on config
	logRequests, err := vm.runtime.RunString("finickyConfigAPI.getOption('logRequests', finalConfig)")
	if err != nil {
		fmt.Printf("Warning: Failed to get logRequests option, defaulting to true to help with debugging: %v", err)
		logRequests = vm.runtime.ToValue(true)
	}
	log.Printf("Logging requests to file: %v", logRequests)
	if logRequests.ToBoolean() {
		log.Printf("You are logging urls to the file system, this may include sensitive information. You can disable this by setting logRequests to false in your config.")
	}

	if err := logger.SetupFile(logRequests.ToBoolean()); err != nil {
		log.Printf("Warning: Failed to setup file logging: %v", err)
	}

}


func (vm *VM) setup(embeddedFiles embed.FS, bundlePath string) error {
	apiContent, err := embeddedFiles.ReadFile("build/finickyConfigAPI.js")
	if err != nil {
		return fmt.Errorf("failed to read bundled file: %v", err)
	}

	var content []byte
	if bundlePath != "" {
		content, err = os.ReadFile(bundlePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
	}

	vm.runtime.Set("self", vm.runtime.GlobalObject())
	vm.runtime.Set("console", GetConsoleMap("js"))

	log.Println("Evaluating API script...")
	if _, err = vm.runtime.RunString(string(apiContent)); err != nil {
		return fmt.Errorf("failed to run api script: %v", err)
	}
	log.Println("Done evaluating API script")

	userAPI := vm.runtime.Get("finickyConfigAPI").ToObject(vm.runtime).Get("utilities").ToObject(vm.runtime)
	finicky := make(map[string]interface{})
	for _, key := range userAPI.Keys() {
		finicky[key] = userAPI.Get(key)
	}

	// These will be injected from main.go
	vm.runtime.Set("finicky", finicky)

	if content != nil {
		if _, err = vm.runtime.RunString(string(content)); err != nil {
			return fmt.Errorf("failed to run config script: %v", err)
		}
	} else {
		vm.runtime.Set(vm.namespace, map[string]interface{}{})
	}

	finalConfig, err := vm.runtime.RunString(fmt.Sprintf("%s.default", vm.namespace))
	if err != nil {
		return fmt.Errorf("failed to get merged config: %v", err)
	}
	vm.runtime.Set("finalConfig", finalConfig)

	validConfig, err := vm.runtime.RunString("finickyConfigAPI.validateConfig(finalConfig)")
	if err != nil {
		return fmt.Errorf("failed to get valid config: %v", err)
	}
	if !validConfig.ToBoolean() {
		return fmt.Errorf("configuration is invalid: %s", validConfig.String())
	}

	// Set system-specific functions
	vm.SetModifierKeysFunc(util.GetModifierKeys)
	vm.SetSystemInfoFunc(util.GetSystemInfo)

	return nil
}

func (vm *VM) prepareConfig() (string, error) {
	configPath, err := vm.getConfigPath()
	if err != nil {
		return "", fmt.Errorf("failed to get config %v", err)
	}

	if configPath == "" {
		log.Printf("Found config path: <None>")
	} else {
		log.Printf("Found config path: %s", configPath)
	}

	if configPath != "" {
		bundlePath := os.TempDir() + "/finicky_output.js"

		result := api.Build(api.BuildOptions{
			EntryPoints: []string{configPath},
			Outfile:     bundlePath,
			Bundle:      true,
			Write:       true,
			LogLevel:    api.LogLevelError,
			Platform:    api.PlatformNeutral,
			Target:      api.ES2015,
			Format:      api.FormatIIFE,
			GlobalName:  vm.namespace,
		})

		if len(result.Errors) > 0 {
			var errorTexts []string
			for _, err := range result.Errors {
				errorTexts = append(errorTexts, err.Text)
			}
			return "", fmt.Errorf("build errors: %s", strings.Join(errorTexts, ", "))
		}
		return bundlePath, nil
	}

	return "", nil
}

func (vm *VM) getConfigPath() (string, error) {
	var configPaths []string
	if os.Getenv("DEBUG") == "true" {
		configPaths = append(configPaths, "./test/example.js")
	} else {
		configPaths = append(configPaths,
			"$HOME/.finicky.js",
			"$HOME/.config/.finicky.js",
		)
	}

	for _, path := range configPaths {
		expandedPath := os.ExpandEnv(path)
		if _, err := os.Stat(expandedPath); err == nil {
			return expandedPath, nil
		}
	}

	return "", fmt.Errorf("no config file found")
}

// Runtime returns the underlying goja.Runtime
func (vm *VM) Runtime() *goja.Runtime {
	return vm.runtime
}

// SetModifierKeysFunc sets the getModifierKeys function in the VM
func (vm *VM) SetModifierKeysFunc(fn func() map[string]bool) {
	finicky := vm.runtime.Get("finicky").ToObject(vm.runtime)
	finicky.Set("getModifierKeys", fn)
}

// SetSystemInfoFunc sets the getSystemInfo function in the VM
func (vm *VM) SetSystemInfoFunc(fn func() map[string]string) {
	finicky := vm.runtime.Get("finicky").ToObject(vm.runtime)
	finicky.Set("getSystemInfo", fn)
}