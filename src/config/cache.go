package config

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"finicky/version"
)

// CacheData represents the structure for caching bundled configuration files
type CacheData struct {
	ConfigPath  string    `json:"configPath"`
	BundlePath  string    `json:"bundlePath"`
	ModTime     time.Time `json:"modTime"`
	AppVersion  string    `json:"appVersion"` // Store app version with cache data
}

// ConfigCache manages persistent caching of bundled configurations
type ConfigCache struct {
	cachedBundlePath string
	cachedConfigPath string
	cachedModTime    time.Time
	cachePath        string // Path to store persistent cache information
	appVersion       string // The current app version
}

// NewConfigCache creates a new cache manager
func NewConfigCache() *ConfigCache {
	// Get current app version
	appVersion := version.GetCurrentVersion()

	// Get the cache file path with version in the name
	cachePath := getCachePath("", fmt.Sprintf("config_cache_%s.json", getContentHash(appVersion, 8)))

	cache := &ConfigCache{
		cachePath:  cachePath,
		appVersion: appVersion,
	}

	// Load cache from disk if it exists
	cache.loadCache()

	return cache
}

// GetCachedBundle returns the cached bundle if it exists and is valid
// Returns the bundle path and true if cache hit, empty string and false otherwise
func (cc *ConfigCache) GetCachedBundle(configPath string) (string, bool) {
	// Check if we have a cache and that it's for the current version
	if cc.cachedBundlePath != "" && cc.cachedConfigPath == configPath {
		// Check if file has been modified
		fileInfo, err := os.Stat(configPath)
		if err == nil && !fileInfo.ModTime().After(cc.cachedModTime) {
			// Verify bundled file still exists
			if _, err := os.Stat(cc.cachedBundlePath); err == nil {
				slog.Debug("Using cached bundled config", "path", cc.cachedBundlePath, "version", cc.appVersion)
				return cc.cachedBundlePath, true
			} else {
				slog.Debug("Cached bundle file no longer exists, rebundling")
			}
		}
	}
	return "", false
}

// UpdateCache updates the cache with a new bundled file
func (cc *ConfigCache) UpdateCache(configPath, bundlePath string) error {
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	cc.cachedBundlePath = bundlePath
	cc.cachedConfigPath = configPath
	cc.cachedModTime = fileInfo.ModTime()

	// Save cache to disk for persistence between runs
	cc.saveCache()
	return nil
}

// Clear clears the current cache and persists that change to disk
func (cc *ConfigCache) Clear() {
	slog.Debug("Clearing configuration cache")
	cc.cachedBundlePath = ""
	cc.cachedConfigPath = ""
	// Update cache file to reflect cleared cache
	cc.saveCache()
}

// loadCache loads the cache information from disk
func (cc *ConfigCache) loadCache() {
	// If cache file doesn't exist, just return
	if _, err := os.Stat(cc.cachePath); os.IsNotExist(err) {
		return
	}

	// Read cache file
	data, err := os.ReadFile(cc.cachePath)
	if err != nil {
		slog.Debug("Failed to read cache file", "error", err)
		return
	}

	// Parse cache data
	var cacheData CacheData
	if err := json.Unmarshal(data, &cacheData); err != nil {
		slog.Debug("Failed to parse cache file", "error", err)
		return
	}

	// Verify the cache is for the current app version
	if cacheData.AppVersion != cc.appVersion {
		slog.Debug("Cache is for a different app version, ignoring",
			"cacheVersion", cacheData.AppVersion,
			"currentVersion", cc.appVersion)
		return
	}

	// Verify the cache is still valid
	if cacheData.BundlePath != "" && cacheData.ConfigPath != "" {
		// Check if bundle file exists
		if _, err := os.Stat(cacheData.BundlePath); os.IsNotExist(err) {
			slog.Debug("Cached bundle file no longer exists", "path", cacheData.BundlePath)
			return
		}

		// Update cache fields
		cc.cachedConfigPath = cacheData.ConfigPath
		cc.cachedBundlePath = cacheData.BundlePath
		cc.cachedModTime = cacheData.ModTime
		slog.Debug("Loaded config cache",
			"configPath", cacheData.ConfigPath,
			"bundlePath", cacheData.BundlePath,
			"version", cacheData.AppVersion)
	}
}

// saveCache saves the cache information to disk
func (cc *ConfigCache) saveCache() {
	// Skip if no cache data
	if cc.cachedConfigPath == "" || cc.cachedBundlePath == "" {
		// If cache is empty, try to delete the cache file
		if _, err := os.Stat(cc.cachePath); err == nil {
			if err := os.Remove(cc.cachePath); err == nil {
				slog.Debug("Removed empty cache file", "path", cc.cachePath)
			}
		}
		return
	}

	// Create cache data structure
	cacheData := CacheData{
		ConfigPath:  cc.cachedConfigPath,
		BundlePath:  cc.cachedBundlePath,
		ModTime:     cc.cachedModTime,
		AppVersion:  cc.appVersion, // Include app version in the cache data
	}

	// Marshal to JSON
	data, err := json.Marshal(cacheData)
	if err != nil {
		slog.Debug("Failed to marshal cache data", "error", err)
		return
	}

	// Write to file
	err = os.WriteFile(cc.cachePath, data, 0644)
	if err != nil {
		slog.Debug("Failed to write cache file", "error", err)
		return
	}

	slog.Debug("Saved config cache", "path", cc.cachePath, "version", cc.appVersion)
}

// fileInfo represents a file with its path and modification time
type fileInfo struct {
	path    string
	modTime time.Time
}

// getFilteredFiles returns a list of files in a directory filtered by prefix and suffix,
// sorted by modification time (oldest first)
func getFilteredFiles(dir, prefix, suffix string, excludeFile string) ([]fileInfo, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var fileInfos []fileInfo
	for _, file := range files {
		if strings.HasPrefix(file.Name(), prefix) && strings.HasSuffix(file.Name(), suffix) {
			filePath := filepath.Join(dir, file.Name())
			if filePath == excludeFile {
				continue // Skip the excluded file
			}

			info, err := file.Info()
			if err != nil {
				continue
			}

			fileInfos = append(fileInfos, fileInfo{
				path:    filePath,
				modTime: info.ModTime(),
			})
		}
	}

	// Sort by modification time (oldest first)
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime.Before(fileInfos[j].modTime)
	})

	return fileInfos, nil
}

// CleanupOldFiles removes old cache files to prevent disk space issues
func CleanupOldFiles(subDir string, currentFile string) {
	cacheDir := getFinickyCacheDir()
	if subDir != "" {
		cacheDir = filepath.Join(cacheDir, subDir)
	}

	// Determine the file prefix based on the subdirectory
	prefix := "finicky_babel_"
	if subDir == "" {
		prefix = "finicky_"
	}

	// Get app version for logging
	appVersion := version.GetCurrentVersion()

	// Get filtered and sorted list of files
	fileInfos, err := getFilteredFiles(cacheDir, prefix, ".js", currentFile)
	if err != nil {
		slog.Debug("Failed to read cache directory for cleanup", "error", err)
		return
	}

	slog.Debug("Cleaning up old cache files", "count", len(fileInfos), "version", appVersion)

	// Keep only the latest 5 files (including current)
	if len(fileInfos) <= 4 { // We're excluding current, so 4 others + current = 5
		return
	}

	// Delete all but the 4 newest files (plus the current one makes 5)
	for i := 0; i < len(fileInfos)-4; i++ {
		err := os.Remove(fileInfos[i].path)
		if err != nil {
			slog.Debug("Failed to remove old cache file", "path", fileInfos[i].path, "error", err)
		} else {
			slog.Debug("Removed old cache file", "path", fileInfos[i].path)
		}
	}
}

// getFinickyCacheDir returns the path to the finicky cache directory,
// creating it if it doesn't exist.
func getFinickyCacheDir() string {
	// Get cache directory in user's cache directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		slog.Debug("Could not get user cache directory", "error", err)
		cacheDir = os.TempDir()
	}

	finickyCacheDir := filepath.Join(cacheDir, "finicky")
	err = os.MkdirAll(finickyCacheDir, 0755)
	if err != nil {
		slog.Debug("Could not create finicky cache directory", "error", err)
	}

	return finickyCacheDir
}

// getCachePath returns a path within the finicky cache directory with optional subdirectories
func getCachePath(subDir string, fileName string) string {
	cacheDir := getFinickyCacheDir()

	if subDir != "" {
		cacheDir = filepath.Join(cacheDir, subDir)
		err := os.MkdirAll(cacheDir, 0755)
		if err != nil {
			slog.Debug("Failed to create cache subdirectory", "dir", subDir, "error", err)
			// Fallback to main cache dir
			cacheDir = getFinickyCacheDir()
		}
	}

	return filepath.Join(cacheDir, fileName)
}

// getContentHash generates a hash for the given content
func getContentHash(content string, length int) string {
	if length <= 0 {
		length = 8 // Default length
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	if len(hash) > length {
		hash = hash[:length]
	}
	return hash
}

// GetTransformedPath returns a deterministic path for a transformed file
func GetTransformedPath(content string) string {
	appVersion := version.GetCurrentVersion()

	// Combine content and app version for the hash
	contentWithVersion := content + "|version:" + appVersion
	contentHash := getContentHash(contentWithVersion, 12)
	return getCachePath("transform", fmt.Sprintf("finicky_babel_%s.js", contentHash))
}

// GetBundlePath returns a deterministic path for a bundled file
func GetBundlePath(configPath string) string {
	appVersion := version.GetCurrentVersion()

	// Combine config path and app version for the hash
	pathWithVersion := configPath + "|version:" + appVersion
	configHash := getContentHash(pathWithVersion, 8)
	return getCachePath("", fmt.Sprintf("finicky_bundle_%s.js", configHash))
}