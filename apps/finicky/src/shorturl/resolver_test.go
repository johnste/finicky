package shorturl

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResolveURLStopsWhenLeavingShortener(t *testing.T) {
	// The target site redirects to a login page, which would lose the link.
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	}))
	defer target.Close()
	// Reach the target through "localhost" so its host differs from the shortener's.
	targetURL := strings.Replace(target.URL, "127.0.0.1", "localhost", 1) + "/user/settings"

	shortener := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, targetURL, http.StatusMovedPermanently)
	}))
	defer shortener.Close()

	saved := shortenerDomains
	shortenerDomains = []string{"127.0.0.1"}
	defer func() { shortenerDomains = saved }()

	got, err := ResolveURL(shortener.URL + "/abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != targetURL {
		t.Errorf("got %q, want %q", got, targetURL)
	}
}

func TestResolveURLIgnoresNonShortener(t *testing.T) {
	saved := shortenerDomains
	shortenerDomains = []string{"short.example"}
	defer func() { shortenerDomains = saved }()

	const u = "https://example.com/page"
	got, err := ResolveURL(u)
	if err != nil || got != u {
		t.Errorf("got %q, %v; want %q, nil", got, err, u)
	}
}
