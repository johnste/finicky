//go:build darwin

package util

import (
	"strings"
	"testing"
)

func TestGetSystemInfo(t *testing.T) {
	info := GetSystemInfo()

	if info["localizedName"] == "" {
		t.Error("localizedName is empty")
	}
	if info["name"] == "" {
		t.Error("name is empty")
	}
}

func TestUserHomeDir(t *testing.T) {
	dir, err := UserHomeDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(dir, "/") {
		t.Errorf("expected absolute path, got %q", dir)
	}
}

func TestUserCacheDir(t *testing.T) {
	dir, err := UserCacheDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(dir, "/") {
		t.Errorf("expected absolute path, got %q", dir)
	}
}

func TestShortenPath(t *testing.T) {
	home, err := UserHomeDir()
	if err != nil {
		t.Skip("could not get home dir")
	}

	cases := []struct {
		input string
		want  string
	}{
		{home, "~"},
		{home + "/foo/bar", "~/foo/bar"},
		{"/other/path", "/other/path"},
		{"", ""},
	}
	for _, c := range cases {
		if got := ShortenPath(c.input); got != c.want {
			t.Errorf("ShortenPath(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
