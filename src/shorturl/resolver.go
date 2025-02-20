package shorturl

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Common URL shortener domains
var shortenerDomains = []string{
	"bit.ly",
	"buff.ly", // buffer.com
	"goo.gl",  // Deprecated, will be shut down after 2025-08-25
	// https://developers.googleblog.com/en/google-url-shortener-links-will-no-longer-be-available/
	"is.gd",    // active as of 2025-01-23
	"ow.ly",    // hootsuite.com
	"spoti.fi", // spotify.com, duh
	"t.co",     // twitter.com
	"wu8.in",
	"dub.sh",
	"d.to",
	"tiny.cc",
	"tinyurl.com",
	"urlshortener.teams.microsoft.com",
	"msteams.link",
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

	log.Printf("URL host looks like a short URL: %s", parsedURL.Host)

	// Create a client with a timeout
	client := &http.Client{
		Timeout: 2 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			log.Printf("Redirected to: %s", req.URL.String())
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
		log.Printf("Got a successful response: %s", resp.Request.URL.String())
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
		log.Printf("Got a successful response: %s", resp.Request.URL.String())
		return resp.Request.URL.String(), nil
	}

	// If both HEAD and GET failed, return original URL
	return originalURL, fmt.Errorf("failed to resolve URL: status code %d", resp.StatusCode)
}
