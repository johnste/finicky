package browser

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Firefox keeps profiles in two places.
//
// Legacy profiles are the [ProfileN] sections of profiles.ini. Each has a Name
// and is launched with "-P <name>".
//
// Profiles created with the newer profile manager (Firefox 138+, the ones
// about:profiles lists) live in "Profile Groups/<StoreID>.sqlite", table
// Profiles(name, path). The group's original profile keeps its profiles.ini
// entry and carries the store ID in its StoreID key; the other members of the
// group exist only in the store. Firefox itself launches those with
// "--profile <absolute path>", which is what we emit for them.

type firefoxProfile struct {
	Name string
	Dir  string // absolute profile directory
	// Legacy is true for profiles.ini entries, which are launched by name.
	Legacy bool
	// StoreID is the profile group store referenced by a profiles.ini entry.
	StoreID string
}

const (
	firefoxProfilesIni      = "profiles.ini"
	firefoxProfileGroupsDir = "Profile Groups"
	firefoxGroupStoreQuery  = "SELECT name, path FROM Profiles ORDER BY id;"
	sqliteReadTimeout       = 1 * time.Second
)

var firefoxProfileSection = regexp.MustCompile(`^\[Profile[0-9]+\]$`)

// readFirefoxProfiles returns profile-group profiles first, then profiles.ini
// profiles. Group profiles come first because those are the names the user
// sees in about:profiles. The bool is false when a group store could not be
// read, so callers can say so instead of reporting a profile as missing.
func readFirefoxProfiles(configDir string) ([]firefoxProfile, bool) {
	legacy := readFirefoxIniProfiles(filepath.Join(configDir, firefoxProfilesIni))
	group, readable := readFirefoxGroupProfiles(configDir, firefoxStoreIDs(legacy))
	return append(group, legacy...), readable
}

// firefoxProfileNames returns the distinct profile names for display. A
// directory listed under two names (the group's original profile has both a
// profiles.ini name and a store name) is listed once, under the first name
// seen, which is the group name given the order readFirefoxProfiles uses.
func firefoxProfileNames(profiles []firefoxProfile) []string {
	names := []string{}
	seenName := map[string]bool{}
	seenDir := map[string]bool{}
	for _, p := range profiles {
		if p.Name == "" || seenName[p.Name] {
			continue
		}
		if p.Dir != "" {
			if seenDir[p.Dir] {
				continue
			}
			seenDir[p.Dir] = true
		}
		seenName[p.Name] = true
		names = append(names, p.Name)
	}
	return names
}

// resolveFirefoxProfileArgs maps a configured profile to Firefox command line
// arguments. An exact profiles.ini name wins, so existing configurations keep
// launching what they always did; then an exact group profile name; then the
// profile directory, as a full path or its base name.
func resolveFirefoxProfileArgs(configDir string, profile string) ([]string, bool) {
	profiles, storesReadable := readFirefoxProfiles(configDir)

	var legacy, group *firefoxProfile
	for i := range profiles {
		p := &profiles[i]
		if p.Name != profile {
			continue
		}
		if p.Legacy {
			if legacy == nil {
				legacy = p
			}
		} else if group == nil {
			group = p
		}
	}
	if legacy != nil && group != nil && legacy.Dir != group.Dir {
		slog.Warn("Firefox profile name exists in both profiles.ini and a profile group store with different directories, using profiles.ini", "name", profile, "profiles.ini", legacy.Dir, "profile group store", group.Dir)
	}
	if legacy != nil {
		return []string{"-P", legacy.Name}, true
	}
	if group != nil {
		slog.Info("Found Firefox profile in profile group store", "name", group.Name, "path", group.Dir)
		return []string{"--profile", group.Dir}, true
	}

	for _, p := range profiles {
		if p.Dir == "" {
			continue
		}
		if profile == p.Dir || profile == filepath.Base(p.Dir) {
			slog.Warn("Found Firefox profile using profile path", "path", p.Dir, "name", p.Name, "suggestion", "Please use the profile name instead")
			return []string{"--profile", p.Dir}, true
		}
	}

	attrs := []any{"Expected profile", profile, "Available profiles", strings.Join(firefoxProfileNames(profiles), ", ")}
	if !storesReadable {
		attrs = append(attrs, "note", "a profile group store could not be read, so profiles from about:profiles may be missing from this list")
	}
	slog.Warn("Could not find profile in Firefox profiles.", attrs...)
	return nil, false
}

// readFirefoxIniProfiles parses the [ProfileN] sections of profiles.ini.
// Relative paths are resolved against the directory containing profiles.ini.
func readFirefoxIniProfiles(profilesIniPath string) []firefoxProfile {
	data, err := os.ReadFile(profilesIniPath)
	if err != nil {
		slog.Info("Error reading profiles.ini", "path", profilesIniPath, "error", err)
		return nil
	}
	baseDir := filepath.Dir(profilesIniPath)

	var profiles []firefoxProfile
	var current *firefoxProfile
	relative := true
	flush := func() {
		if current == nil {
			return
		}
		if current.Dir != "" && relative {
			current.Dir = filepath.Join(baseDir, current.Dir)
		}
		if current.Name != "" {
			profiles = append(profiles, *current)
		}
		current = nil
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") {
			flush()
			if firefoxProfileSection.MatchString(line) {
				current = &firefoxProfile{Legacy: true}
				relative = true // IsRelative defaults to 1
			}
			continue
		}
		if current == nil {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch key {
		case "Name":
			current.Name = value
		case "Path":
			current.Dir = value
		case "IsRelative":
			relative = value != "0"
		case "StoreID":
			current.StoreID = value
		}
	}
	flush()
	return profiles
}

// firefoxStoreIDs returns the distinct store IDs referenced by profiles.ini
// entries, in order of appearance.
func firefoxStoreIDs(profiles []firefoxProfile) []string {
	var ids []string
	seen := map[string]bool{}
	for _, p := range profiles {
		if p.StoreID == "" || seen[p.StoreID] {
			continue
		}
		seen[p.StoreID] = true
		ids = append(ids, p.StoreID)
	}
	return ids
}

// readFirefoxGroupProfiles reads the stores referenced from profiles.ini.
// Firefox writes the store ID to the group's profiles.ini entry when it
// populates a store, so unreferenced stores are stale or empty. Only when
// nothing references a store is every store in the directory read as a
// fallback; a referenced store that is missing is reported instead, so a
// stale store is never read in its place.
// The bool is false when at least one referenced store could not be read.
func readFirefoxGroupProfiles(configDir string, storeIDs []string) ([]firefoxProfile, bool) {
	groupsDir := filepath.Join(configDir, firefoxProfileGroupsDir)

	readable := true

	var storePaths []string
	for _, id := range storeIDs {
		storePath := filepath.Join(groupsDir, id+".sqlite")
		if _, err := os.Stat(storePath); err != nil {
			slog.Info("Firefox profile group store referenced by profiles.ini not found", "path", storePath)
			readable = false
			continue
		}
		storePaths = append(storePaths, storePath)
	}
	if len(storeIDs) == 0 {
		storePaths, _ = filepath.Glob(filepath.Join(groupsDir, "*.sqlite"))
		sort.Strings(storePaths)
	}

	var profiles []firefoxProfile
	for _, storePath := range storePaths {
		rows, ok := readFirefoxGroupStore(configDir, storePath)
		if !ok {
			readable = false
		}
		profiles = append(profiles, rows...)
	}
	return profiles, readable
}

// readFirefoxGroupStore reads the Profiles table of one group store with the
// sqlite3 command line tool that ships with macOS. The store is opened read
// only; Firefox may have it open at the same time, which is fine in WAL mode.
// Relative paths in the store are resolved against configDir, the same base
// profiles.ini uses. Rows whose directory no longer exists are dropped, since
// launching them would silently create an empty profile.
func readFirefoxGroupStore(configDir string, storePath string) ([]firefoxProfile, bool) {
	sqlite, ok := sqliteBinary()
	if !ok {
		slog.Warn("sqlite3 not found, cannot read Firefox profile group store", "path", storePath)
		return nil, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), sqliteReadTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, sqlite, "-batch", "-readonly", "-json", "-cmd", ".timeout 500", storePath, firefoxGroupStoreQuery)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		slog.Warn("Error reading Firefox profile group store", "path", storePath, "sqlite3", sqlite, "error", err, "stderr", strings.TrimSpace(stderr.String()))
		return nil, false
	}

	out := bytes.TrimSpace(stdout.Bytes())
	if len(out) == 0 {
		return nil, true
	}

	var rows []struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal(out, &rows); err != nil {
		slog.Warn("Error parsing Firefox profile group store output", "path", storePath, "sqlite3", sqlite, "error", err)
		return nil, false
	}

	profiles := make([]firefoxProfile, 0, len(rows))
	for _, row := range rows {
		if row.Name == "" || row.Path == "" {
			continue
		}
		dir := row.Path
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(configDir, dir)
		}
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			slog.Warn("Firefox profile directory listed in profile group store does not exist, skipping", "name", row.Name, "path", dir)
			continue
		}
		profiles = append(profiles, firefoxProfile{Name: row.Name, Dir: dir})
	}
	return profiles, true
}

// sqliteBinary prefers the sqlite3 bundled with macOS and falls back to PATH.
func sqliteBinary() (string, bool) {
	const system = "/usr/bin/sqlite3"
	if _, err := os.Stat(system); err == nil {
		return system, true
	}
	if path, err := exec.LookPath("sqlite3"); err == nil {
		return path, true
	}
	return "", false
}
