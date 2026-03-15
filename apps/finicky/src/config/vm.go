package config

import (
	"finicky/util"
	"fmt"
	"log/slog"
	"os"

	"github.com/dop251/goja"
)

type VM struct {
	runtime    *goja.Runtime
	namespace  string
	isJSConfig bool
}

// ConfigOptions holds the values of all runtime config options.
type ConfigOptions struct {
	KeepRunning     bool
	HideIcon        bool
	LogRequests     bool
	CheckForUpdates bool
}

// ConfigState represents the current state of the configuration
type ConfigState struct {
	Handlers       int16  `json:"handlers"`
	Rewrites       int16  `json:"rewrites"`
	DefaultBrowser string `json:"defaultBrowser"`
}

// New creates a VM from a JS config file on disk. The resulting VM is marked
// as a JS-config VM (IsJSConfig() == true).
// apiContent is the pre-read bytes of finickyConfigAPI.js.
func New(apiContent []byte, namespace string, bundlePath string) (*VM, error) {
	var content []byte
	if bundlePath != "" {
		var err error
		content, err = os.ReadFile(bundlePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %v", err)
		}
	}
	vm, err := newFromContent(apiContent, namespace, content)
	if vm != nil {
		vm.isJSConfig = true
	}
	return vm, err
}

// NewFromScript creates a VM from an inline JavaScript config string.
// apiContent is the pre-read bytes of finickyConfigAPI.js.
func NewFromScript(apiContent []byte, namespace string, script string) (*VM, error) {
	return newFromContent(apiContent, namespace, []byte(script))
}

func newFromContent(apiContent []byte, namespace string, content []byte) (*VM, error) {
	vm := &VM{
		runtime:   goja.New(),
		namespace: namespace,
	}
	if err := vm.setup(apiContent, content); err != nil {
		return nil, err
	}
	return vm, nil
}

func (vm *VM) setup(apiContent []byte, content []byte) error {

	vm.runtime.Set("self", vm.runtime.GlobalObject())
	vm.runtime.Set("console", GetConsoleMap())

	slog.Debug("Evaluating API script...")
	if _, err := vm.runtime.RunString(string(apiContent)); err != nil {
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
		if _, err := vm.runtime.RunString(string(content)); err != nil {
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

// IsJSConfig reports whether this VM was built from a JS config file.
func (vm *VM) IsJSConfig() bool {
	return vm.isJSConfig
}

// SetIsJSConfig overrides the JS-config flag. Intended for use in tests.
func (vm *VM) SetIsJSConfig(v bool) {
	vm.isJSConfig = v
}

// GetAllConfigOptions reads all runtime config options in a single JS call.
// Safe to call on a nil VM — returns defaults in that case.
func (vm *VM) GetAllConfigOptions() ConfigOptions {
	defaults := ConfigOptions{
		KeepRunning:     true,
		HideIcon:        false,
		LogRequests:     false,
		CheckForUpdates: true,
	}
	if vm == nil || vm.runtime == nil {
		return defaults
	}
	script := `({
		keepRunning:     finickyConfigAPI.getOption('keepRunning',     finalConfig, true),
		hideIcon:        finickyConfigAPI.getOption('hideIcon',        finalConfig, false),
		logRequests:     finickyConfigAPI.getOption('logRequests',     finalConfig, false),
		checkForUpdates: finickyConfigAPI.getOption('checkForUpdates', finalConfig, true)
	})`
	val, err := vm.runtime.RunString(script)
	if err != nil {
		slog.Error("Failed to get config options", "error", err)
		return defaults
	}
	obj := val.ToObject(vm.runtime)
	return ConfigOptions{
		KeepRunning:     obj.Get("keepRunning").ToBoolean(),
		HideIcon:        obj.Get("hideIcon").ToBoolean(),
		LogRequests:     obj.Get("logRequests").ToBoolean(),
		CheckForUpdates: obj.Get("checkForUpdates").ToBoolean(),
	}
}

// Runtime returns the underlying goja.Runtime
func (vm *VM) Runtime() *goja.Runtime {
	return vm.runtime
}
