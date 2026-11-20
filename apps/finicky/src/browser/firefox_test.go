package browser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// groupStoreSchema mirrors the Profiles table Firefox creates in
// "Profile Groups/<StoreID>.sqlite".
const groupStoreSchema = `CREATE TABLE IF NOT EXISTS "Profiles" (
  id      INTEGER NOT NULL,
  path    TEXT NOT NULL UNIQUE,
  name    TEXT NOT NULL,
  avatar  TEXT NOT NULL,
  themeId TEXT NOT NULL,
  themeFg TEXT NOT NULL,
  themeBg TEXT NOT NULL,
  PRIMARY KEY(id)
);`

type groupRow struct {
	name string
	path string
}

func sqlQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// mkProfileDir creates the profile directory for a relative store/ini path
// and returns its absolute path.
func mkProfileDir(t *testing.T, configDir string, rel string) string {
	t.Helper()
	dir := filepath.Join(configDir, rel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// writeGroupStore creates a profile group store with the given rows. Tests
// that need it skip when sqlite3 is unavailable.
func writeGroupStore(t *testing.T, configDir string, storeID string, rows []groupRow) {
	t.Helper()
	sqlite, ok := sqliteBinary()
	if !ok {
		t.Skip("sqlite3 not available")
	}
	storePath := filepath.Join(configDir, firefoxProfileGroupsDir, storeID+".sqlite")
	if err := os.MkdirAll(filepath.Dir(storePath), 0o755); err != nil {
		t.Fatal(err)
	}
	stmts := "PRAGMA journal_mode=WAL;" + groupStoreSchema
	for _, r := range rows {
		stmts += fmt.Sprintf("INSERT INTO Profiles (path, name, avatar, themeId, themeFg, themeBg) VALUES (%s, %s, '', '', '', '');", sqlQuote(r.path), sqlQuote(r.name))
	}
	out, err := exec.Command(sqlite, "-batch", storePath, stmts).CombinedOutput()
	if err != nil {
		t.Fatalf("creating group store: %v: %s", err, out)
	}
}

const sampleProfilesIni = `[General]
StartWithLastProfile=1
Version=2

[Profile0]
Name=default-release
IsRelative=1
Path=Profiles/x1y2z3w4.default-release
StoreID=1a2b3c4d
ShowSelector=1

[Install0123456789ABCDEF]
Default=Profiles/x1y2z3w4.default-release
Locked=1

[Profile1]
Name=default
IsRelative=1
Path=Profiles/q5w6e7r8.default
Default=1

[Profile2]
Name=external
IsRelative=0
Path=/Volumes/Other/firefox-profile

[ProfileNotASection]
Name=ignored
Path=Profiles/ignored
`

func TestReadFirefoxIniProfiles(t *testing.T) {
	configDir := t.TempDir()
	iniPath := filepath.Join(configDir, firefoxProfilesIni)
	writeFile(t, iniPath, sampleProfilesIni)

	got := readFirefoxIniProfiles(iniPath)
	want := []firefoxProfile{
		{Name: "default-release", Dir: filepath.Join(configDir, "Profiles/x1y2z3w4.default-release"), Legacy: true, StoreID: "1a2b3c4d"},
		{Name: "default", Dir: filepath.Join(configDir, "Profiles/q5w6e7r8.default"), Legacy: true},
		{Name: "external", Dir: "/Volumes/Other/firefox-profile", Legacy: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("readFirefoxIniProfiles:\n got %+v\nwant %+v", got, want)
	}
	if ids := firefoxStoreIDs(got); !reflect.DeepEqual(ids, []string{"1a2b3c4d"}) {
		t.Errorf("firefoxStoreIDs: got %v", ids)
	}
}

func TestReadFirefoxIniProfiles_Missing(t *testing.T) {
	if got := readFirefoxIniProfiles(filepath.Join(t.TempDir(), "profiles.ini")); len(got) != 0 {
		t.Errorf("expected no profiles for missing file, got %+v", got)
	}
}

func TestReadFirefoxGroupProfiles_ReferencedStoresOnly(t *testing.T) {
	configDir := t.TempDir()
	personal := mkProfileDir(t, configDir, "Profiles/x1y2z3w4.default-release")
	work := mkProfileDir(t, configDir, "Profiles/abcd1234.Profile 1")
	mkProfileDir(t, configDir, "Profiles/stale.other")
	writeGroupStore(t, configDir, "1a2b3c4d", []groupRow{
		{name: "Personal", path: "Profiles/x1y2z3w4.default-release"},
		{name: "Work", path: "Profiles/abcd1234.Profile 1"},
	})
	// An unreferenced store with rows must be ignored when a referenced one exists.
	writeGroupStore(t, configDir, "0000ffff", []groupRow{
		{name: "Other", path: "Profiles/stale.other"},
	})

	got, readable := readFirefoxGroupProfiles(configDir, []string{"1a2b3c4d"})
	want := []firefoxProfile{
		{Name: "Personal", Dir: personal},
		{Name: "Work", Dir: work},
	}
	if !readable || !reflect.DeepEqual(got, want) {
		t.Errorf("readFirefoxGroupProfiles:\n got %+v (readable=%v)\nwant %+v", got, readable, want)
	}
}

func TestReadFirefoxGroupProfiles_FallbackToAllStores(t *testing.T) {
	configDir := t.TempDir()
	other := t.TempDir()
	absDir := filepath.Join(other, "abs-profile")
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		t.Fatal(err)
	}
	personal := mkProfileDir(t, configDir, "Profiles/x1y2z3w4.default-release")
	work := mkProfileDir(t, configDir, "Profiles/abcd1234.Profile 1")
	// Stores are read in file name order: "1111" before "2222".
	writeGroupStore(t, configDir, "2222", []groupRow{{name: "Absolute", path: absDir}})
	writeGroupStore(t, configDir, "1111", []groupRow{
		{name: "Personal", path: "Profiles/x1y2z3w4.default-release"},
		{name: "Work", path: "Profiles/abcd1234.Profile 1"},
		{name: "Deleted", path: "Profiles/gone.Profile 2"}, // directory missing: dropped
	})
	writeGroupStore(t, configDir, "3333", nil) // empty store

	// Nothing references a store, so every store in the directory is read.
	got, readable := readFirefoxGroupProfiles(configDir, nil)
	want := []firefoxProfile{
		{Name: "Personal", Dir: personal},
		{Name: "Work", Dir: work},
		{Name: "Absolute", Dir: absDir},
	}
	if !readable || !reflect.DeepEqual(got, want) {
		t.Errorf("readFirefoxGroupProfiles(nil):\n got %+v (readable=%v)\nwant %+v", got, readable, want)
	}
}

func TestReadFirefoxGroupProfiles_MissingReferencedStore(t *testing.T) {
	configDir := t.TempDir()
	work := mkProfileDir(t, configDir, "Profiles/abcd1234.Profile 1")
	writeGroupStore(t, configDir, "1111", []groupRow{{name: "Work", path: "Profiles/abcd1234.Profile 1"}})

	// The referenced store is missing, so nothing is read: the unreferenced
	// "1111" store is stale and must not be used in its place.
	got, readable := readFirefoxGroupProfiles(configDir, []string{"deadbeef"})
	if readable || len(got) != 0 {
		t.Errorf("readFirefoxGroupProfiles([deadbeef]):\n got %+v (readable=%v)\nwant no profiles (readable=false)", got, readable)
	}

	// Referencing the store that exists still reads it.
	got, readable = readFirefoxGroupProfiles(configDir, []string{"1111"})
	want := []firefoxProfile{{Name: "Work", Dir: work}}
	if !readable || !reflect.DeepEqual(got, want) {
		t.Errorf("readFirefoxGroupProfiles([1111]):\n got %+v (readable=%v)\nwant %+v", got, readable, want)
	}
}

func TestReadFirefoxGroupProfiles_UnreadableStore(t *testing.T) {
	configDir := t.TempDir()
	work := mkProfileDir(t, configDir, "Profiles/abcd1234.Profile 1")
	writeGroupStore(t, configDir, "1111", []groupRow{{name: "Work", path: "Profiles/abcd1234.Profile 1"}})
	writeFile(t, filepath.Join(configDir, firefoxProfileGroupsDir, "2222.sqlite"), "not a database")

	got, readable := readFirefoxGroupProfiles(configDir, []string{"1111", "2222"})
	want := []firefoxProfile{{Name: "Work", Dir: work}}
	if readable {
		t.Error("expected readable=false when a store cannot be read")
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("readFirefoxGroupProfiles:\n got %+v\nwant %+v", got, want)
	}
}

func TestReadFirefoxGroupProfiles_NoStoreDir(t *testing.T) {
	got, readable := readFirefoxGroupProfiles(t.TempDir(), nil)
	if !readable || len(got) != 0 {
		t.Errorf("expected no profiles without a Profile Groups dir, got %+v (readable=%v)", got, readable)
	}
}

func TestFirefoxProfileNames(t *testing.T) {
	profiles := []firefoxProfile{
		{Name: "Personal", Dir: "/a"},
		{Name: "Work", Dir: "/b"},
		{Name: "", Dir: "/c"},
		{Name: "Shared", Dir: "/d"},
		{Name: "default-release", Dir: "/a", Legacy: true}, // same directory as Personal: hidden
		{Name: "Shared", Dir: "/e", Legacy: true},          // same name: hidden
		{Name: "default", Dir: "/f", Legacy: true},
		{Name: "nodir", Legacy: true},
	}
	got := firefoxProfileNames(profiles)
	want := []string{"Personal", "Work", "Shared", "default", "nodir"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("firefoxProfileNames: got %v, want %v", got, want)
	}
}

// fixtureConfigDir builds a Firefox config dir with both a legacy profiles.ini
// and a referenced profile group store. "Shared" exists in both with different
// directories to exercise precedence; "Personal" and "default-release" are the
// same directory under two names, as Firefox does for the group's first profile.
func fixtureConfigDir(t *testing.T) string {
	t.Helper()
	configDir := t.TempDir()
	for _, rel := range []string{
		"Profiles/x1y2z3w4.default-release",
		"Profiles/abcd1234.Profile 1",
		"Profiles/group.shared",
		"Profiles/legacy.shared",
	} {
		mkProfileDir(t, configDir, rel)
	}
	writeFile(t, filepath.Join(configDir, firefoxProfilesIni), `[Profile0]
Name=default-release
IsRelative=1
Path=Profiles/x1y2z3w4.default-release
StoreID=1a2b3c4d

[Profile1]
Name=Shared
IsRelative=1
Path=Profiles/legacy.shared
`)
	writeGroupStore(t, configDir, "1a2b3c4d", []groupRow{
		{name: "Personal", path: "Profiles/x1y2z3w4.default-release"},
		{name: "Work", path: "Profiles/abcd1234.Profile 1"},
		{name: "Shared", path: "Profiles/group.shared"},
	})
	return configDir
}

func TestResolveFirefoxProfileArgs(t *testing.T) {
	configDir := fixtureConfigDir(t)
	workDir := filepath.Join(configDir, "Profiles/abcd1234.Profile 1")

	cases := []struct {
		profile string
		want    []string
		ok      bool
	}{
		{"Personal", []string{"--profile", filepath.Join(configDir, "Profiles/x1y2z3w4.default-release")}, true},
		{"Work", []string{"--profile", workDir}, true},
		{"default-release", []string{"-P", "default-release"}, true},
		// profiles.ini wins over a group profile with the same name.
		{"Shared", []string{"-P", "Shared"}, true},
		// Directory base name and full path are accepted as fallbacks.
		{"abcd1234.Profile 1", []string{"--profile", workDir}, true},
		{workDir, []string{"--profile", workDir}, true},
		{"legacy.shared", []string{"--profile", filepath.Join(configDir, "Profiles/legacy.shared")}, true},
		{"Missing", nil, false},
		{"", nil, false},
	}
	for _, c := range cases {
		got, ok := resolveFirefoxProfileArgs(configDir, c.profile)
		if ok != c.ok || !reflect.DeepEqual(got, c.want) {
			t.Errorf("resolveFirefoxProfileArgs(%q): got (%v, %v), want (%v, %v)", c.profile, got, ok, c.want, c.ok)
		}
	}
}

func TestResolveFirefoxProfileArgs_LegacyOnly(t *testing.T) {
	configDir := t.TempDir()
	writeFile(t, filepath.Join(configDir, firefoxProfilesIni), "[Profile0]\nName=default-release\nPath=Profiles/abc.default-release\n")

	got, ok := resolveFirefoxProfileArgs(configDir, "default-release")
	if !ok || !reflect.DeepEqual(got, []string{"-P", "default-release"}) {
		t.Errorf("legacy only: got (%v, %v)", got, ok)
	}
	if _, ok := resolveFirefoxProfileArgs(configDir, "Personal"); ok {
		t.Error("expected no match for a group profile name when no store exists")
	}
}

func TestReadFirefoxProfiles_Order(t *testing.T) {
	configDir := fixtureConfigDir(t)
	profiles, readable := readFirefoxProfiles(configDir)
	if !readable {
		t.Fatal("expected stores to be readable")
	}
	got := firefoxProfileNames(profiles)
	// Group names first; default-release is hidden behind Personal (same
	// directory) and the legacy Shared behind the group Shared (same name).
	want := []string{"Personal", "Work", "Shared"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("profile name order: got %v, want %v", got, want)
	}
}
