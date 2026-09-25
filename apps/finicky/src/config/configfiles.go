package config

import (
	"finicky/util"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/fsnotify/fsnotify"
	babel "github.com/jvatic/goja-babel"
)

// ConfigFileWatcher handles watching configuration files for changes
type ConfigFileWatcher struct {
	watcher            *fsnotify.Watcher
	customConfigPath   string
	namespace          string
	configChangeNotify chan struct{}

	// Cache manager
	cache *ConfigCache

	// Debounce rapid file-change events (e.g. editors that write twice)
	debounceMu    sync.Mutex
	debounceTimer *time.Timer
}

// NewConfigFileWatcher creates a new file watcher for configuration files
func NewConfigFileWatcher(customConfigPath string, namespace string, configChangeNotify chan struct{}) (*ConfigFileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	cfw := &ConfigFileWatcher{
		watcher:            watcher,
		customConfigPath:   customConfigPath,
		namespace:          namespace,
		configChangeNotify: configChangeNotify,
		cache:              NewConfigCache(),
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

	homeDir, err := util.UserHomeDir()
	if err != nil {
		slog.Error("Failed to get user home directory", "error", err)
		return configPaths
	}

	if cfw.customConfigPath != "" {
		configPaths = append(configPaths, cfw.customConfigPath)
	} else {
		// XDG Base Directory spec: use $XDG_CONFIG_HOME when set to an
		// absolute path, falling back to ~/.config (relative values are ignored)
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome == "" || !filepath.IsAbs(xdgConfigHome) {
			xdgConfigHome = filepath.Join(homeDir, ".config")
		}

		configPaths = append(configPaths,
			"~/.finicky.js",
			"~/.finicky.ts",
			filepath.Join(xdgConfigHome, "finicky.js"),
			filepath.Join(xdgConfigHome, "finicky.ts"),
			filepath.Join(xdgConfigHome, "finicky", "finicky.js"),
			filepath.Join(xdgConfigHome, "finicky", "finicky.ts"),
		)
	}

	for i, path := range configPaths {
		configPaths[i] = os.ExpandEnv(path)
		configPaths[i] = strings.ReplaceAll(configPaths[i], "~", homeDir)
	}

	return configPaths
}

// GetConfigPath returns the path to an existing configuration file
func (cfw *ConfigFileWatcher) GetConfigPath(log bool) (string, error) {
	configPaths := cfw.GetConfigPaths()

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			// Resolve symlinks to get the actual file path
			resolvedPath, err := resolveSymlink(path)
			if err != nil {
				slog.Warn("Failed to resolve symlink, using original path", "original", path, "error", err)
				resolvedPath = path
			}

			if log {
				if resolvedPath != path {
					slog.Info("Using config file", "path", resolvedPath)
				} else {
					slog.Info("Using config file", "path", path)
				}
			}
			return resolvedPath, nil
		}
	}
	if cfw.customConfigPath != "" {
		return "", fmt.Errorf("no config file found at %s", cfw.customConfigPath)
	}
	return "", fmt.Errorf("no config file found in any of these locations: %s", strings.Join(configPaths, ", "))
}

func (cfw *ConfigFileWatcher) BundleConfig() (string, string, error) {
	configPath, err := cfw.GetConfigPath(true)

	if configPath == "" || err != nil {
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

func (cfw *ConfigFileWatcher) StartWatching() error {
	for {
		configPath, err := cfw.GetConfigPath(false)

		if err != nil {
			// Watch any potential config paths
			configPaths := cfw.GetConfigPaths()

			// Use a map to track unique folders
			uniqueFolders := make(map[string]bool)
			for _, path := range configPaths {
				folder := filepath.Dir(path)
				uniqueFolders[folder] = true
			}

			// Folders may not exist yet (e.g. $XDG_CONFIG_HOME/finicky), so
			// watch the nearest existing ancestor of each and refresh watches
			// as intermediate directories get created. fsnotify watches are
			// not recursive. The watcher's live watch list is the source of
			// truth: fsnotify drops a watch itself when a directory is
			// deleted, so a directory that is removed and recreated gets
			// watched again here.
			addWatches := func() {
				watched := make(map[string]bool)
				for _, path := range cfw.watcher.WatchList() {
					watched[path] = true
				}
				for folder := range uniqueFolders {
					ancestor := nearestExistingDir(folder)
					if ancestor == "" || watched[ancestor] {
						continue
					}
					if err := cfw.watcher.Add(ancestor); err != nil {
						slog.Debug("Error watching folder", "folder", ancestor, "error", err)
					}
				}
			}
			addWatches()

			removeAllWatches := func() {
				for _, path := range cfw.watcher.WatchList() {
					if err := cfw.watcher.Remove(path); err != nil {
						slog.Debug("Error removing watch on folder", "folder", path, "error", err)
					}
				}
			}

			slog.Debug("Watching for config files", "paths", cfw.watcher.WatchList())

			detectedCreation := false
			for !detectedCreation {
				select {
				case event, ok := <-cfw.watcher.Events:
					if !ok {
						return fmt.Errorf("watcher closed")
					}

					// A removed or renamed directory loses its fsnotify watch,
					// and its ancestors may not be watched; fall back to the
					// nearest existing ancestor so a recreation is detected.
					if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
						addWatches()
					}

					if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
						// Check if the event path matches any of our config paths
						eventName := event.Name
						isConfigFile := false
						for _, path := range configPaths {
							if eventName == path {
								isConfigFile = true
								break
							}
						}

						if !isConfigFile && event.Has(fsnotify.Create) && isAncestorOfAny(eventName, uniqueFolders) {
							// A directory on the way to a config folder was
							// created; start watching it so a config file
							// created inside it is detected.
							addWatches()

							// The directory may have been moved into place
							// with a config file already inside it, in which
							// case no separate file event will follow.
							if foundPath, err := cfw.GetConfigPath(false); err == nil {
								event = fsnotify.Event{Name: foundPath, Op: fsnotify.Create}
								isConfigFile = true
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

						removeAllWatches()
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
	// Ignore CHMOD-only events (permission changes) as they don't affect config content
	// Note: Some editors may send CHMOD along with WRITE, so we only ignore pure CHMOD
	if event.Op == fsnotify.Chmod {
		return nil
	}

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
		select {
		case cfw.configChangeNotify <- struct{}{}:
		default:
		}
		return fmt.Errorf("configuration file removed")
	}

	// Debounce: reset the timer so only the last event in a burst fires.
	cfw.debounceMu.Lock()
	if cfw.debounceTimer != nil {
		cfw.debounceTimer.Stop()
	}
	notify := cfw.configChangeNotify
	cfw.debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
		select {
		case notify <- struct{}{}:
		default: // drop if a notification is already pending
		}
	})
	cfw.debounceMu.Unlock()
	return nil
}

// nearestExistingDir walks up from path until it finds a directory that
// exists, returning "" if none does
func nearestExistingDir(path string) string {
	for {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}
		parent := filepath.Dir(path)
		if parent == path {
			return ""
		}
		path = parent
	}
}

// isAncestorOfAny reports whether path equals, or is an ancestor of, any of
// the given folders
func isAncestorOfAny(path string, folders map[string]bool) bool {
	for folder := range folders {
		if folder == path || strings.HasPrefix(folder, path+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

// resolveSymlink resolves a symlink to its target file path
// If the path is not a symlink, it returns the original path
func resolveSymlink(path string) (string, error) {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return path, err
	}

	// Check if it's a symlink
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		resolvedPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return path, fmt.Errorf("failed to resolve symlink %s: %w", path, err)
		}
		slog.Debug("Resolved symlink", "original", path, "resolved", resolvedPath)
		return resolvedPath, nil
	}

	return path, nil
}
