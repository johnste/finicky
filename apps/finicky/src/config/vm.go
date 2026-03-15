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
	runtime   *goja.Runtime
	namespace string
}

// ConfigState represents the current state of the configuration
type ConfigState struct {
	Handlers       int16  `json:"handlers"`
	Rewrites       int16  `json:"rewrites"`
	DefaultBrowser string `json:"defaultBrowser"`
}

func New(embeddedFiles embed.FS, namespace string, bundlePath string) (*VM, error) {
	var content []byte
	if bundlePath != "" {
		var err error
		content, err = os.ReadFile(bundlePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %v", err)
		}
	}
	return newFromContent(embeddedFiles, namespace, content)
}

// NewFromScript creates a VM from an inline JavaScript config string.
func NewFromScript(embeddedFiles embed.FS, namespace string, script string) (*VM, error) {
	return newFromContent(embeddedFiles, namespace, []byte(script))
}

func newFromContent(embeddedFiles embed.FS, namespace string, content []byte) (*VM, error) {
	vm := &VM{
		runtime:   goja.New(),
		namespace: namespace,
	}
	if err := vm.setup(embeddedFiles, content); err != nil {
		return nil, err
	}
	return vm, nil
}

func (vm *VM) setup(embeddedFiles embed.FS, content []byte) error {
	apiContent, err := embeddedFiles.ReadFile("assets/finickyConfigAPI.js")
	if err != nil {
		return fmt.Errorf("failed to read bundled file: %v", err)
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

	// Set system-specific functions
	finicky["getModifierKeys"] = util.GetModifierKeys
	finicky["getSystemInfo"] = util.GetSystemInfo
	finicky["getPowerInfo"] = util.GetPowerInfo
	finicky["isAppRunning"] = util.IsAppRunning

	vm.runtime.Set("finicky", finicky)

	if len(content) > 0 {
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

	return nil
}

func (vm *VM) GetConfigState() *ConfigState {
	state, err := vm.runtime.RunString("finickyConfigAPI.getConfigState(finalConfig)")
	if err != nil {
		slog.Error("Failed to get config state", "error", err)
		return nil
	}

	// Convert the JavaScript object to a Go struct
	stateObj := state.ToObject(vm.runtime)

	// Extract values from the JavaScript object
	handlers := stateObj.Get("handlers").ToInteger()
	rewrites := stateObj.Get("rewrites").ToInteger()
	defaultBrowser := stateObj.Get("defaultBrowser").String()

	return &ConfigState{
		Handlers:       int16(handlers),
		Rewrites:       int16(rewrites),
		DefaultBrowser: defaultBrowser,
	}
}

// Runtime returns the underlying goja.Runtime
func (vm *VM) Runtime() *goja.Runtime {
	return vm.runtime
}
