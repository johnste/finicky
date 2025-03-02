package config

import (
	"embed"
	"finicky/logger"
	"finicky/util"
	"fmt"
	"log/slog"
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
func New(embeddedFiles embed.FS, customConfigPath string) (*VM, error) {
	vm := &VM{
		runtime: goja.New(),
		namespace: "finickyConfig",
	}

	bundlePath, err := vm.prepareConfig(customConfigPath)
	defer vm.setupLogging(err != nil)
	if err != nil {
		return nil, err
	}

	err = vm.setup(embeddedFiles, bundlePath)
	if err != nil {
		return nil, err
	}

	return vm, err
}

func (vm *VM) setupLogging(hasError bool) {

	logRequests := vm.runtime.ToValue(hasError)

	if !hasError {
		var err error
		logRequests, err = vm.runtime.RunString("finickyConfigAPI.getOption('logRequests', finalConfig)")
		if err != nil {
			slog.Warn("Failed to get logRequests option", "error", err)
			logRequests = vm.runtime.ToValue(true)
		}
	}

	if logRequests.ToBoolean() {
		slog.Warn("Logging requests to disk. Logs may include sensitive information. Disable this by setting logRequests: false.")
	}

	if err := logger.SetupFile(logRequests.ToBoolean()); err != nil {
		slog.Warn("Failed to setup file logging", "error", err)
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
	vm.runtime.Set("console", GetConsoleMap())

	slog.Debug("Evaluating API script...")
	if _, err = vm.runtime.RunString(string(apiContent)); err != nil {
		return fmt.Errorf("failed to run api script: %v", err)
	}
	slog.Debug("Done evaluating API script")

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

	vm.runtime.Set("namespace", vm.namespace)
	finalConfig, err := vm.runtime.RunString("finickyConfigAPI.getConfiguration(namespace)")
	if err != nil {
		return fmt.Errorf("failed to get merged config: %v", err)
	}
	slog.Debug("Final config", "config", finalConfig)
	vm.runtime.Set("finalConfig", finalConfig)

	validConfig, err := vm.runtime.RunString("finickyConfigAPI.validateConfig(finalConfig)")
	if err != nil {
		return fmt.Errorf("failed to validate config: %v", err)
	}
	if !validConfig.ToBoolean() {
		return fmt.Errorf("configuration is invalid")
	}

	// Set system-specific functions
	vm.SetModifierKeysFunc(util.GetModifierKeys)
	vm.SetSystemInfoFunc(util.GetSystemInfo)

	return nil
}

func (vm *VM) prepareConfig(customConfigPath string) (string, error) {


	configPath, err := vm.getConfigPath(customConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to get config %v", err)
	}

	if configPath == "" {
		slog.Info("Found config: <None>")
	} else {
		slog.Info("Found config", "path", configPath)
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

func (vm *VM) getConfigPath(customConfigPath string) (string, error) {
	var configPaths []string

	if customConfigPath != "" {
		configPaths = append(configPaths, customConfigPath)
	} else {
		configPaths = append(configPaths,
			"$HOME/.finicky.js",
			"$HOME/.config/finicky.js",
			"$HOME/.config/finicky/finicky.js",
		)
	}

	for _, path := range configPaths {
		expandedPath := os.ExpandEnv(path)
		if _, err := os.Stat(expandedPath); err == nil {
			return expandedPath, nil
		}
	}

	return "", fmt.Errorf("no config file found, please create one at %s", strings.Join(configPaths, ", "))
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