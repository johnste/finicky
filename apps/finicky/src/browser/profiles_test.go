package browser

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeTempFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// denyDir makes a directory unreadable for the rest of the test. Root ignores
// the mode bits, so there is nothing to assert there.
func denyDir(t *testing.T, dir string) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("running as root: permission bits are not enforced")
	}
	if err := os.Chmod(dir, 0o000); err != nil {
		t.Fatal(err)
	}
	// Restore access so t.TempDir cleanup can remove the directory.
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
}

const sampleProfilesIni = `[Profile0]
Name=default-release
Path=Profiles/x1y2z3w4.default-release

[Profile1]
Name=Work
Path=Profiles/abcd1234.work
`

func TestReportBrowserDataReadError(t *testing.T) {
	// A denial and an absence must not be reported the same way: one is
	// fixed by a privacy setting, the other means the file is simply gone.
	if got := reportBrowserDataReadError("profiles.ini", "/x", fs.ErrPermission); !got {
		t.Error("expected a permission error to be reported as denied")
	}
	if got := reportBrowserDataReadError("profiles.ini", "/x", fs.ErrNotExist); got {
		t.Error("expected a missing file not to be reported as denied")
	}
}

func TestReadFirefoxProfileNames(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.ini")
	writeTempFile(t, path, sampleProfilesIni)

	names, denied := readFirefoxProfileNames(path)
	if denied {
		t.Error("expected denied=false for a readable file")
	}
	if want := []string{"default-release", "Work"}; !reflect.DeepEqual(names, want) {
		t.Errorf("names: got %v, want %v", names, want)
	}
}

func TestReadFirefoxProfileNames_Missing(t *testing.T) {
	names, denied := readFirefoxProfileNames(filepath.Join(t.TempDir(), "profiles.ini"))
	if len(names) != 0 {
		t.Errorf("expected no names, got %v", names)
	}
	if denied {
		t.Error("a missing profiles.ini must not be reported as a denial")
	}
}

func TestReadFirefoxProfileNames_Denied(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, filepath.Join(dir, "profiles.ini"), sampleProfilesIni)
	denyDir(t, dir)

	names, denied := readFirefoxProfileNames(filepath.Join(dir, "profiles.ini"))
	if len(names) != 0 {
		t.Errorf("expected no names, got %v", names)
	}
	if !denied {
		t.Error("expected denied=true when the directory cannot be read")
	}
}

func TestParseFirefoxProfiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.ini")
	writeTempFile(t, path, sampleProfilesIni)

	if name, ok, denied := parseFirefoxProfiles(path, "Work"); !ok || denied || name != "Work" {
		t.Errorf("match: got (%q, %v, %v)", name, ok, denied)
	}
	// Readable list, name absent: genuinely wrong, and not a denial.
	if name, ok, denied := parseFirefoxProfiles(path, "Nope"); ok || denied || name != "" {
		t.Errorf("miss: got (%q, %v, %v)", name, ok, denied)
	}
}

func TestParseFirefoxProfiles_Denied(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, filepath.Join(dir, "profiles.ini"), sampleProfilesIni)
	denyDir(t, dir)

	_, ok, denied := parseFirefoxProfiles(filepath.Join(dir, "profiles.ini"), "Work")
	if ok {
		t.Error("expected no match when the list cannot be read")
	}
	if !denied {
		t.Error("expected denied=true so the caller can fall back")
	}
}

func TestFirefoxFallbackArgs(t *testing.T) {
	dir := t.TempDir()
	profileDir := filepath.Join(dir, "Profiles", "abcd1234.work")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatal(err)
	}
	aFile := filepath.Join(dir, "a-file")
	writeTempFile(t, aFile, "")

	cases := []struct {
		name    string
		profile string
		denied  bool
		want    []string
		ok      bool
	}{
		// An absolute path needs no access to anything, so it resolves
		// whether or not the list could be read.
		{"existing path", profileDir, false, []string{"--profile", profileDir}, true},
		{"existing path under denial", profileDir, true, []string{"--profile", profileDir}, true},
		// Firefox creates an empty profile for a directory that is not
		// there, which loses a session silently, so refuse it.
		{"missing path", filepath.Join(dir, "Profiles", "typo"), true, nil, false},
		{"path is a file", aFile, true, nil, false},
		// Firefox resolves a name itself, but only worth handing over when
		// we were not allowed to check it.
		{"name under denial", "Work", true, []string{"-P", "Work"}, true},
		{"name without denial", "Work", false, nil, false},
		{"empty", "", true, nil, false},
	}
	for _, c := range cases {
		got, ok := firefoxFallbackArgs(c.profile, c.denied)
		if ok != c.ok || !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got (%v, %v), want (%v, %v)", c.name, got, ok, c.want, c.ok)
		}
	}
}

func TestParseProfiles_Denied(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, filepath.Join(dir, "Local State"), `{"profile":{"info_cache":{"Profile 1":{"name":"Work"}}}}`)

	path := filepath.Join(dir, "Local State")
	if got, ok, denied := parseProfiles(path, "Work"); !ok || denied || got != "Profile 1" {
		t.Errorf("match: got (%q, %v, %v)", got, ok, denied)
	}

	denyDir(t, dir)
	got, ok, denied := parseProfiles(path, "Work")
	if ok || got != "" {
		t.Errorf("expected no match under a denial, got %q", got)
	}
	if !denied {
		t.Error("expected denied=true so the caller can fall back")
	}
}

// A JSON or shape problem is not a denial: it must not send the user off to
// change a privacy setting.
func TestParseProfiles_MalformedIsNotDenial(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Local State")
	writeTempFile(t, path, "not json")

	if _, ok, denied := parseProfiles(path, "Work"); ok || denied {
		t.Errorf("malformed: got (ok=%v, denied=%v)", ok, denied)
	}
}
