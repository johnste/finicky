package config

import (
	"embed"
	"encoding/json"
	"finicky/logger"
	"fmt"
	"log"
	"os"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
)

// VM represents the JavaScript VM configuration
type VM struct {
	runtime *goja.Runtime
	namespace string
}

// GetConsoleMap returns a console object for the VM
func GetConsoleMap(prefix string) map[string]interface{} {
	return map[string]interface{}{
		"log": func(msg string) {
			log.Printf("[%s] %s", prefix, msg)
		},
	}
}

// New creates and configures a new VM instance
func New(embeddedFiles embed.FS) (*VM, error) {
	vm := &VM{
		runtime: goja.New(),
		namespace: "finickyConfig",
	}

	bundlePath, simpleConfigPath, err := vm.prepareConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to bundle config: %v", err)
	}

	if err := vm.setup(embeddedFiles, bundlePath, simpleConfigPath); err != nil {
		return nil, err
	}

	return vm, nil
}

func (vm *VM) setup(embeddedFiles embed.FS, bundlePath, simpleConfigPath string) error {
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

	var simpleContent []byte
	if simpleConfigPath != "" {
		simpleContent, err = os.ReadFile(simpleConfigPath)
		var simpleConfig map[string]interface{}
		if err := json.Unmarshal(simpleContent, &simpleConfig); err != nil {
			return fmt.Errorf("failed to unmarshal simple config: %v", err)
		}
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

	mergedConfig, err := vm.runtime.RunString(fmt.Sprintf("finickyConfigAPI.mergeConfig(%s.default, %s)", vm.namespace, simpleContent))
	if err != nil {
		return fmt.Errorf("failed to get merged config: %v", err)
	}
	vm.runtime.Set("mergedConfig", mergedConfig)

	validConfig, err := vm.runtime.RunString("finickyConfigAPI.validateConfig(mergedConfig)")
	if err != nil {
		return fmt.Errorf("failed to get valid config: %v", err)
	}
	if !validConfig.ToBoolean() {
		return fmt.Errorf("configuration is invalid: %s", validConfig.String())
	}

	// Setup logging based on config
	logRequests, err := vm.runtime.RunString("finickyConfigAPI.getOption('logRequests', mergedConfig)")
	if err != nil {
		return fmt.Errorf("failed to get logRequests option: %v", err)
	}
	log.Printf("logRequests: %v", logRequests)

	if err := logger.SetupFile(logRequests.ToBoolean()); err != nil {
		log.Printf("Warning: Failed to setup file logging: %v", err)
	}

	return nil
}

func (vm *VM) prepareConfig() (string, string, error) {
	simpleConfigPath, errSimple := vm.getSimpleConfigPath()
	configPath, err := vm.getConfigPath()
	if err != nil && errSimple != nil {
		return "", "", fmt.Errorf("failed to get config or simple config: %v", err)
	}

	if configPath == "" {
		log.Printf("Found config path: <None>")
	} else {
		log.Printf("Found config path: %s", configPath)
	}
	if simpleConfigPath == "" {
		log.Printf("Found simple config path: <None>")
	} else {
		log.Printf("Found simple config path: %s", simpleConfigPath)
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
			return "", "", fmt.Errorf("build errors: %v", result.Errors)
		}
		return bundlePath, simpleConfigPath, nil
	}

	return "", simpleConfigPath, nil
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

func (vm *VM) getSimpleConfigPath() (string, error) {
	var configPaths []string
	if os.Getenv("DEBUG") == "true" {
		configPaths = append(configPaths, "./test/example.json")
	} else {
		configPaths = append(configPaths,
			"$HOME/.finicky.json",
			"$HOME/.config/.finicky.json",
		)
	}

	for _, path := range configPaths {
		expandedPath := os.ExpandEnv(path)
		if _, err := os.Stat(expandedPath); err == nil {
			return expandedPath, nil
		}
	}

	return "", fmt.Errorf("no simple config file found")
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