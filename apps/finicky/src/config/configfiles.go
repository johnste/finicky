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

	// Cache manager
	cache *ConfigCache
}

// NewConfigFileWatcher creates a new file watcher for configuration files
func NewConfigFileWatcher(customConfigPath string, namespace string, configChangeNotify chan struct{}) (*ConfigFileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	cfw := &ConfigFileWatcher{
		watcher:          watcher,
		customConfigPath: customConfigPath,
		namespace:        namespace,
		configChangeNotify: configChangeNotify,
		cache:            NewConfigCache(),
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
			"$HOME/.finicky.ts",
			"$HOME/.config/finicky.js",
			"$HOME/.config/finicky.ts",
			"$HOME/.config/finicky/finicky.js",
			"$HOME/.config/finicky/finicky.ts",
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

func (cfw *ConfigFileWatcher) BundleConfig() (string, string, error) {
	configPath, err := cfw.GetConfigPath(true)

	if (configPath == "" || err != nil) {
		return "", "", err
	}

	// Check if we can use cached bundle
	if bundlePath, cacheHit := cfw.cache.GetCachedBundle(configPath); cacheHit {
		return bundlePath, configPath, nil
	}

	// Apply babel transformation
	transformedPath, err := cfw.babelTransform(configPath)
	if err != nil {
		return "", configPath, err
	}

	slog.Debug("Bundling config")

	// Use a deterministic filename to help with caching
	bundlePath := GetBundlePath(transformedPath)

	result := api.Build(api.BuildOptions{
		EntryPoints: []string{transformedPath},
		Outfile:     bundlePath,
		Bundle:      true,
		Write:       true,
		LogLevel:    api.LogLevelError,
		Platform:    api.PlatformNeutral,
		Target:      api.ES2015,
		Format:      api.FormatIIFE,
		GlobalName:  cfw.namespace,
		Loader: map[string]api.Loader{
			".ts.symlink": api.LoaderTS,
			".js.symlink": api.LoaderJS,
		},
	})

	if len(result.Errors) > 0 {
		var errorTexts []string
		for _, err := range result.Errors {
			errorTexts = append(errorTexts, err.Text)
		}
		return "", configPath, fmt.Errorf("build errors: %s", strings.Join(errorTexts, ", "))
	}

	// Update cache
	originalConfigPath, err := cfw.GetConfigPath(false)
	if err == nil {
		cfw.cache.UpdateCache(originalConfigPath, bundlePath)
	}

	return bundlePath, configPath, nil
}

func (cfw *ConfigFileWatcher) babelTransform(configPath string) (string, error) {
	startTime := time.Now()
	slog.Debug("Transforming config with babel")

	// Check if we need to transform (only if it's a .js or .mjs file)
	ext := filepath.Ext(configPath)
	if ext != ".js" && ext != ".mjs" {
		slog.Debug("Skipping babel transform for non-JS file", "path", configPath)
		return configPath, nil
	}

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("error reading config file: %w", err)
	}
	configString := string(configBytes)

	babel.Init(1) // Setup transformers (can be any number > 0)
	res, err := babel.Transform(strings.NewReader(configString), map[string]interface{}{
		"plugins": []string{
			"transform-named-capturing-groups-regex",
		},
	})

	if err != nil {
		return "", err
	}

	resBytes, err := io.ReadAll(res)
	if err != nil {
		return "", err
	}
	resString := string(resBytes)

	// Get a deterministic path for the transformed file
	transformedPath := GetTransformedPath(configString)

	// Check if transformed file already exists
	if _, err := os.Stat(transformedPath); err == nil {
		slog.Debug("Using existing transformed file", "path", transformedPath)
		return transformedPath, nil
	}

	// Write to the persistent location
	err = os.WriteFile(transformedPath, []byte(resString), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing to transform file: %w", err)
	}

	slog.Debug("Saved babel output", "path", transformedPath)
	slog.Debug("Babel transform complete", "duration", fmt.Sprintf("%.2fms", float64(time.Since(startTime).Microseconds())/1000))

	// Clean up old transformed files
	CleanupOldFiles("transform", transformedPath)

	return transformedPath, nil
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

						err := cfw.handleConfigFileEvent(event)
						if err != nil {
							return err
						}

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
				err := cfw.handleConfigFileEvent(event)
				if err != nil {
					return err
				}
			case err, ok := <-cfw.watcher.Errors:
				if !ok {
					return fmt.Errorf("watcher closed")
				}
				slog.Debug("error:", "error", err)
			}
		}
	}
	// Unreachable - infinite loop above. Added for completeness only.
	// return nil
}

// handleConfigFileEvent processes configuration file events and takes appropriate actions
// Returns an error if the configuration file was removed
func (cfw *ConfigFileWatcher) handleConfigFileEvent(event fsnotify.Event) error {
	if event.Has(fsnotify.Create) {
		slog.Debug("Configuration file created", "path", event.Name)
	}

	if event.Has(fsnotify.Write) {
		slog.Debug("Configuration file changed", "path", event.Name)
		// Clear the cache when config changes
		cfw.cache.Clear()
	}

	if event.Has(fsnotify.Remove) {
		slog.Debug("Configuration file removed", "path", event.Name)
		// Clear the cache when config is removed
		cfw.cache.Clear()
		cfw.configChangeNotify <- struct{}{}
		return fmt.Errorf("configuration file removed")
	}

	// Add a small delay to avoid rapid reloading
	time.Sleep(500 * time.Millisecond)
	cfw.configChangeNotify <- struct{}{}
	return nil
}