package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestGetConfigPathsDefaultsToDotConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	cfw := &ConfigFileWatcher{}
	paths := cfw.GetConfigPaths()

	if len(paths) != 6 {
		t.Fatalf("expected 6 config paths, got %d: %v", len(paths), paths)
	}

	for _, p := range paths {
		if strings.Contains(p, "~") {
			t.Errorf("path %q was not expanded", p)
		}
	}

	if !strings.HasSuffix(paths[2], filepath.Join(".config", "finicky.js")) {
		t.Errorf("expected fallback to ~/.config, got %q", paths[2])
	}
	if !strings.HasSuffix(paths[4], filepath.Join(".config", "finicky", "finicky.js")) {
		t.Errorf("expected fallback to ~/.config/finicky, got %q", paths[4])
	}
}

func TestGetConfigPathsHonorsXDGConfigHome(t *testing.T) {
	xdgHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdgHome)

	cfw := &ConfigFileWatcher{}
	paths := cfw.GetConfigPaths()

	expected := []string{
		filepath.Join(xdgHome, "finicky.js"),
		filepath.Join(xdgHome, "finicky.ts"),
		filepath.Join(xdgHome, "finicky", "finicky.js"),
		filepath.Join(xdgHome, "finicky", "finicky.ts"),
	}

	for i, want := range expected {
		got := paths[i+2]
		if got != want {
			t.Errorf("paths[%d] = %q, want %q", i+2, got, want)
		}
	}
}

func TestGetConfigPathsIgnoresRelativeXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "relative/config")

	cfw := &ConfigFileWatcher{}
	paths := cfw.GetConfigPaths()

	for _, p := range paths {
		if strings.Contains(p, "relative") {
			t.Errorf("relative XDG_CONFIG_HOME should be ignored, got %q", p)
		}
	}
	if !strings.HasSuffix(paths[2], filepath.Join(".config", "finicky.js")) {
		t.Errorf("expected fallback to ~/.config, got %q", paths[2])
	}
}

func TestNearestExistingDir(t *testing.T) {
	base := t.TempDir()

	if got := nearestExistingDir(base); got != base {
		t.Errorf("nearestExistingDir(%q) = %q, want the dir itself", base, got)
	}

	missing := filepath.Join(base, "finicky", "nested")
	if got := nearestExistingDir(missing); got != base {
		t.Errorf("nearestExistingDir(%q) = %q, want nearest ancestor %q", missing, got, base)
	}
}

func TestIsAncestorOfAny(t *testing.T) {
	folders := map[string]bool{
		"/home/user/.config/finicky": true,
	}

	cases := []struct {
		path string
		want bool
	}{
		{"/home/user/.config/finicky", true},
		{"/home/user/.config", true},
		{"/home/user", true},
		{"/home/user/.config/finick", false},
		{"/home/other", false},
	}

	for _, c := range cases {
		if got := isAncestorOfAny(c.path, folders); got != c.want {
			t.Errorf("isAncestorOfAny(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestGetConfigPathsCustomPathWins(t *testing.T) {
	cfw := &ConfigFileWatcher{customConfigPath: "/tmp/custom-finicky.js"}
	paths := cfw.GetConfigPaths()

	if len(paths) != 1 || paths[0] != "/tmp/custom-finicky.js" {
		t.Errorf("expected only the custom path, got %v", paths)
	}
}

func TestWatcherPicksUpConfigCreatedInMissingDirs(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "a", "b", "finicky.js")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	notify := make(chan struct{}, 1)
	cfw := &ConfigFileWatcher{
		watcher:            watcher,
		customConfigPath:   configPath,
		configChangeNotify: notify,
		cache:              &ConfigCache{cachePath: filepath.Join(root, "cache.json")},
	}
	defer cfw.TearDown()
	go cfw.StartWatching()

	waitForNotify := func(what string) {
		t.Helper()
		select {
		case <-notify:
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for notification after %s", what)
		}
	}

	// Give the watcher time to watch the nearest existing ancestor (root)
	time.Sleep(100 * time.Millisecond)

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte("export default {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	waitForNotify("creating config in missing directories")

	if err := os.WriteFile(configPath, []byte("export default { defaultBrowser: \"Safari\" }\n"), 0644); err != nil {
		t.Fatal(err)
	}
	waitForNotify("editing config")
}
