package shorturl

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestResolveURL_RequestLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect":
			http.Redirect(w, r, "/target", http.StatusFound)
		case "/get":
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusOK)
		case "/fail", "/loop":
			http.Redirect(w, r, "/loop", http.StatusFound)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(server.Close)
	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	previousDomains := shortenerDomains
	shortenerDomains = []string{parsed.Host}
	t.Cleanup(func() { shortenerDomains = previousDomains })
	previousLogger := slog.Default()
	var output bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	for _, path := range []string{"/redirect", "/get", "/fail"} {
		var firstURL, firstError string
		for _, enabled := range []bool{false, true} {
			output.Reset()
			result, err := ResolveURL(server.URL+path, enabled)
			errorText := ""
			if err != nil {
				errorText = err.Error()
			}
			if !enabled {
				firstURL, firstError = result, errorText
			} else if result != firstURL || errorText != firstError {
				t.Fatalf("logging changed resolution: %q %q vs %q %q", firstURL, firstError, result, errorText)
			}
			logs := output.String()
			for _, marker := range []string{"URL host looks like a short URL", "Redirected to", "Got a successful response", "Falling back to last known URL from redirects"} {
				want := enabled
				if marker == "Redirected to" {
					want = enabled && path != "/get"
				}
				if marker == "Got a successful response" {
					want = enabled && path != "/fail"
				}
				if marker == "Falling back to last known URL from redirects" {
					want = enabled && path == "/fail"
				}
				if strings.Contains(logs, marker) != want {
					t.Errorf("%s %v: %s presence != %v: %s", path, enabled, marker, want, logs)
				}
			}
			if path == "/fail" {
				if err == nil || result != server.URL+"/loop" || !strings.Contains(logs, "level=DEBUG msg=\"Failed to make HEAD request\"") {
					t.Fatalf("failure/fallback diagnostic changed: %q %v %s", result, err, logs)
				}
			} else {
				want := server.URL + path
				if path == "/redirect" {
					want = server.URL + "/target"
				}
				if err != nil || result != want {
					t.Fatalf("result = %q %v, want %q", result, err, want)
				}
			}
		}
	}
}
