package window

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestServer(t *testing.T) {
	if err := StartAPIServer(); err != nil {
		t.Fatal(err)
	}
	base := strings.TrimSuffix(UIURL(), "/")

	get := func(t *testing.T, path, host, token string) *http.Response {
		t.Helper()
		req, err := http.NewRequest("GET", base+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if host != "" {
			req.Host = host
		}
		if token != "" {
			req.Header.Set("X-Finicky-Token", token)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { res.Body.Close() })
		return res
	}

	t.Run("serves the UI without a token", func(t *testing.T) {
		res := get(t, "/", "", "")
		body, _ := io.ReadAll(res.Body)
		if res.StatusCode != http.StatusOK || !strings.Contains(string(body), `<div id="app">`) {
			t.Fatalf("got %d %q", res.StatusCode, body)
		}
	})

	t.Run("serves UI assets with their content type", func(t *testing.T) {
		res := get(t, "/index.js", "", "")
		if res.StatusCode != http.StatusOK || !strings.Contains(res.Header.Get("Content-Type"), "javascript") {
			t.Fatalf("got %d %q", res.StatusCode, res.Header.Get("Content-Type"))
		}
	})

	t.Run("requires the token for the API", func(t *testing.T) {
		if res := get(t, "/api/browsers", "", ""); res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("no token: got %d", res.StatusCode)
		}
		if res := get(t, "/api/browsers", "", "wrong"); res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("wrong token: got %d", res.StatusCode)
		}
		if res := get(t, "/api/browsers", "", apiToken); res.StatusCode != http.StatusOK {
			t.Fatalf("valid token: got %d", res.StatusCode)
		}
	})

	t.Run("rejects other Host headers", func(t *testing.T) {
		for _, host := range []string{"evil.example", "localhost"} {
			if res := get(t, "/", host, ""); res.StatusCode != http.StatusForbidden {
				t.Errorf("%s: got %d", host, res.StatusCode)
			}
			if res := get(t, "/api/browsers", host, apiToken); res.StatusCode != http.StatusForbidden {
				t.Errorf("%s with token: got %d", host, res.StatusCode)
			}
		}
	})
}
