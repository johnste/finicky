package shorturl

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//go:embed shortener_domains.json
var shortenerDomainsFS embed.FS

// Common URL shortener domains
var shortenerDomains []string

func init() {
	// Load shortener domains from embedded JSON file
	data, err := shortenerDomainsFS.ReadFile("shortener_domains.json")
	if err != nil {
		slog.Error("Failed to read shortener domains file", "error", err)
		// Fallback to empty list if file cannot be read
		shortenerDomains = []string{}
		return
	}

	if err := json.Unmarshal(data, &shortenerDomains); err != nil {
		slog.Error("Failed to parse shortener domains JSON", "error", err)
		// Fallback to empty list if JSON is invalid
		shortenerDomains = []string{}
	}
}

// ResolveURL resolves a potentially shortened URL to its final destination by following HTTP redirects, so
// the matcher can match the final URL instead of the short URL.
func ResolveURL(originalURL string) (string, error) {
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return originalURL, fmt.Errorf("failed to parse URL: %v", err)
	}

	// Check if the domain is a known URL shortener
	isShortURL := false
	for _, domain := range shortenerDomains {
		if strings.HasSuffix(parsedURL.Host, domain) {
			isShortURL = true
			break
		}
	}

	if !isShortURL {
		return originalURL, nil
	}

	slog.Debug("URL host looks like a short URL", "host", parsedURL.Host)

	// Create a client with a timeout
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			slog.Debug("Redirected to", "url", req.URL.String())
			// Allow up to 3 redirects
			if len(via) >= 3 {
				return fmt.Errorf("stopped after 3 redirects")
			}
			return nil
		},
	}

	// Make a HEAD request first to follow redirects without downloading content
	req, err := http.NewRequest("HEAD", originalURL, nil)
	if err != nil {
		return originalURL, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "Finicky/4.0")

	resp, err := client.Do(req)
	if err != nil {
		return originalURL, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// If we got a successful response, return the final URL
	if resp.StatusCode == http.StatusOK {
		slog.Debug("Got a successful response", "url", resp.Request.URL.String())
		return resp.Request.URL.String(), nil
	}

	// If HEAD request failed, try GET as fallback
	req, err = http.NewRequest("GET", originalURL, nil)
	if err != nil {
		return originalURL, fmt.Errorf("failed to create GET request: %v", err)
	}
	req.Header.Set("User-Agent", "Finicky/4.0")

	resp, err = client.Do(req)
	if err != nil {
		return originalURL, fmt.Errorf("failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		slog.Debug("Got a successful response", "url", resp.Request.URL.String())
		return resp.Request.URL.String(), nil
	}

	// If both HEAD and GET failed, return original URL
	return originalURL, fmt.Errorf("failed to resolve URL: status code %d", resp.StatusCode)
}
