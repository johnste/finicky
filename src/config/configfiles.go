package config

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/fsnotify/fsnotify"
	babel "github.com/jvatic/goja-babel"
)

// ConfigFileWatcher handles watching configuration files for changes
type ConfigFileWatcher struct {
	watcher           *fsnotify.Watcher
	customConfigPath  string
	namespace         string
	configChangeNotify chan struct{}
}

// NewConfigFileWatcher creates a new file watcher for configuration files
func NewConfigFileWatcher(customConfigPath string, namespace string, configChangeNotify chan struct{}) (*ConfigFileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %v", err)
	}

	cfw := &ConfigFileWatcher{
		watcher:           watcher,
		customConfigPath:  customConfigPath,
		namespace:         namespace,
		configChangeNotify: configChangeNotify,
	}

	go cfw.StartWatching()

	return cfw, nil
}

// TearDown closes the file watcher
func (cfw *ConfigFileWatcher) TearDown() {
	if cfw.watcher != nil {
		cfw.watcher.Close()
	}
}

// GetConfigPaths returns a list of potential configuration file paths
func (cfw *ConfigFileWatcher) GetConfigPaths() []string {
	var configPaths []string

	if cfw.customConfigPath != "" {
		configPaths = append(configPaths, cfw.customConfigPath)
	} else {
		configPaths = append(configPaths,
			"$HOME/.finicky.js",
			"$HOME/.config/finicky.js",
			"$HOME/.config/finicky/finicky.js",
		)
	}

	for i, path := range configPaths {
		configPaths[i] = os.ExpandEnv(path)
	}

	return configPaths
}

// GetConfigPath returns the path to an existing configuration file
func (cfw *ConfigFileWatcher) GetConfigPath(log bool) (string, error) {
	configPaths := cfw.GetConfigPaths()

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			if log {
				slog.Info("Using config file", "path", path)
			}
			return path, nil
		}
	}
	if cfw.customConfigPath != "" {
		return "", fmt.Errorf("no config file found at %s", cfw.customConfigPath)
	}
	return "", fmt.Errorf("no config file found in any of these locations: %s", strings.Join(configPaths, ", "))
}

func (cfw *ConfigFileWatcher) BundleConfig() (string, error) {
	configPath, err := cfw.GetConfigPath(true)

	if (configPath == "" || err != nil) {
		return "", err
	}

	err = cfw.babelTransform(configPath)
	if err != nil {
		return "", err
	}

	slog.Debug("Bundling config")
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
		GlobalName:  cfw.namespace,
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

func (cfw *ConfigFileWatcher) babelTransform(configPath string) (error) {

	startTime := time.Now()
	slog.Debug("Transforming config with babel")

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}
	configString := string(configBytes)

	babel.Init(1) // Setup 4 transformers (can be any number > 0)
	res, err := babel.Transform(strings.NewReader(configString), map[string]interface{}{
		"plugins": []string{
			"transform-named-capturing-groups-regex",
		},
	})
	if err != nil {
		slog.Error("Babel error", "error", err)
	}

	resBytes, err := io.ReadAll(res)
	if err != nil {
		return err
	}
	resString := string(resBytes)

	tempFile, err := os.CreateTemp("", "finicky_babel_*.js")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(resString); err != nil {
		return fmt.Errorf("error writing to temp file: %w", err)
	}

	configPath = tempFile.Name()
	slog.Debug("Saved babel output", "path", configPath)

	slog.Debug("Babel transform complete", "duration", fmt.Sprintf("%.2fms", float64(time.Since(startTime).Microseconds())/1000))
	return nil
}

func (cfw *ConfigFileWatcher) StartWatching() (error) {
	for {
		configPath, err := cfw.GetConfigPath(false)

		if err != nil {
			// Watch any potential config paths
			configPaths := cfw.GetConfigPaths()

			// Use a map to track unique folders
			uniqueFolders := make(map[string]bool)
			for _, path := range configPaths {
				expandedPath := os.ExpandEnv(path)
				folder := filepath.Dir(expandedPath)
				uniqueFolders[folder] = true
			}

			// Convert to slice for logging
			var watchPaths []string
			for folder := range uniqueFolders {
				watchPaths = append(watchPaths, folder)
			}

			slog.Debug("Watching for config files", "paths", watchPaths)

			// Add each unique folder to the watcher
			for folder := range uniqueFolders {
				if err := cfw.watcher.Add(folder); err != nil {
					slog.Debug("Error watching folder", "folder", folder, "error", err)
				}
			}

			detectedCreation := false
			for !detectedCreation {
				select {
				case event, ok := <-cfw.watcher.Events:
					if !ok {
						return fmt.Errorf("watcher closed")
					}


					if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
						// Check if the event path matches any of our config paths
						eventName := event.Name
						isConfigFile := false
						for _, path := range configPaths {
							expandedPath := os.ExpandEnv(path)
							if eventName == expandedPath {
								isConfigFile = true
								break
							}
						}

						if !isConfigFile {
							break
						}

						detectedCreation = true

						if event.Has(fsnotify.Create) {
							slog.Debug("Configuration file created", "path", event.Name)
						}
						if event.Has(fsnotify.Write) {
							slog.Debug("Configuration file changed", "path", event.Name)
						}

						// Add a small delay to avoid rapid reloading
						time.Sleep(500 * time.Millisecond)
						cfw.configChangeNotify <- struct{}{}

						for folder := range uniqueFolders {
							if err := cfw.watcher.Remove(folder); err != nil {
								slog.Debug("Error removing watch on folder", "folder", folder, "error", err)
							}
						}
					}

				case err, ok := <-cfw.watcher.Errors:
					if !ok {
						return fmt.Errorf("watcher closed")
					}
					slog.Debug("error:", "error", err)
				}
			}



		} else {
			slog.Debug("Watching config file", "path", configPath)

			cfw.watcher.Add(configPath)

			select {
			case event, ok := <-cfw.watcher.Events:
				if !ok {
					return fmt.Errorf("watcher closed")
				}
				if event.Has(fsnotify.Write) {
					slog.Debug("Configuration file changed")
					// Add a small delay to avoid rapid reloading
					time.Sleep(500 * time.Millisecond)
					cfw.configChangeNotify <- struct{}{}
				}

				if event.Has(fsnotify.Remove) {
					slog.Debug("Configuration file removed", "path", event.Name)
					cfw.configChangeNotify <- struct{}{}
					return fmt.Errorf("configuration file removed")
				}
			case err, ok := <-cfw.watcher.Errors:
				if !ok {
					return fmt.Errorf("watcher closed")
				}
				slog.Debug("error:", "error", err)
			}

		}

	}
	return nil
}
