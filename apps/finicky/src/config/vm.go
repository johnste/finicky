package config

import (
	"embed"
	"finicky/util"
	"fmt"
	"log/slog"
	"os"

	"github.com/dop251/goja"
)

type VM struct {
	runtime *goja.Runtime
	namespace string
}

func New(embeddedFiles embed.FS, namespace string, bundlePath string) (*VM, error) {
	vm := &VM{
		runtime:          goja.New(),
		namespace:        namespace,
	}

	err := vm.setup(embeddedFiles, bundlePath)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func (vm *VM) setup(embeddedFiles embed.FS, bundlePath string) error {
	apiContent, err := embeddedFiles.ReadFile("assets/finickyConfigAPI.js")
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

	vm.runtime.Set("finicky", finicky)

	if content != nil {
		if _, err = vm.runtime.RunString(string(content)); err != nil {
			return fmt.Errorf("error while running config script: %v", err)
		}
	} else {
		vm.runtime.Set(vm.namespace, map[string]interface{}{})
	}

	vm.runtime.Set("namespace", vm.namespace)
	finalConfig, err := vm.runtime.RunString("finickyConfigAPI.getConfiguration(namespace)")
	if err != nil {
		return fmt.Errorf("failed to get merged config: %v", err)
	}

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
	vm.SetPowerInfoFunc(util.GetPowerInfo)

	return nil
}

func (vm *VM) ShouldLogToFile(hasError bool) bool {

	logRequests := vm.runtime.ToValue(hasError)

	if !hasError {
		var err error
		logRequests, err = vm.runtime.RunString("finickyConfigAPI.getOption('logRequests', finalConfig)")
		if err != nil {
			slog.Warn("Failed to get logRequests option", "error", err)
			logRequests = vm.runtime.ToValue(true)
		}
	}

	return logRequests.ToBoolean()
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

// SetPowerInfoFunc sets the getPowerInfo function in the VM
func (vm *VM) SetPowerInfoFunc(fn func() map[string]interface{}) {
	finicky := vm.runtime.Get("finicky").ToObject(vm.runtime)
	finicky.Set("getPowerInfo", fn)
}
