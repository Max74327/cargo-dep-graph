// Package registry provides a client for the crates.io API.
package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// UserAgent is sent with every request; crates.io rejects requests without it.
const UserAgent = "cargo-dep-graph/0.2"

// MaxResponseSize is the upper bound for an API response (1 MiB).
const MaxResponseSize = 10 << 20

// Client is a crates.io API client with an in-memory cache.
type Client struct {
	httpClient  *http.Client
	baseURL     string
	cache       map[string][]string
	lastRequest time.Time
}

// NewClient creates a new crates.io client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		baseURL:    "https://crates.io",
		cache:      make(map[string][]string),
	}
}

// NewClientWithBaseURL creates a client pointing at a custom base URL.
// Intended for tests.
func NewClientWithBaseURL(baseURL string) *Client {
	c := NewClient()
	c.baseURL = baseURL
	return c
}

// FetchDependencies returns the names of normal runtime dependencies of name.
// The result is cached in memory for the lifetime of the client.
func (c *Client) FetchDependencies(name string) ([]string, error) {

	if since := time.Since(c.lastRequest); since < 100*time.Millisecond {
		time.Sleep(100*time.Millisecond - since)
	}
	c.lastRequest = time.Now()

	if deps, ok := c.cache[name]; ok {
		return deps, nil
	}

	version, err := c.latestVersion(name)
	if err != nil {
		return nil, fmt.Errorf("registry: latest version of %s: %w", name, err)
	}

	deps, err := c.dependencies(name, version)
	if err != nil {
		return nil, fmt.Errorf("registry: dependencies of %s@%s: %w", name, version, err)
	}

	c.cache[name] = deps
	return deps, nil
}

// latestVersion returns the latest version of the crate.
func (c *Client) latestVersion(name string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/crates/%s", c.baseURL, name)
	var resp CrateResponse
	if err := c.getJSON(url, &resp); err != nil {
		return "", err
	}
	if resp.Crate.MaxVersion == "" {
		return "", fmt.Errorf("empty max_version for %s", name)
	}
	return resp.Crate.MaxVersion, nil
}

// dependencies returns normal (runtime) dependencies of name@version.
func (c *Client) dependencies(name, version string) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/crates/%s/%s/dependencies", c.baseURL, name, version)
	var resp DependenciesResponse
	if err := c.getJSON(url, &resp); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(resp.Dependencies))
	for _, d := range resp.Dependencies {
		if d.Kind == "normal" {
			result = append(result, d.CrateName)
		}
	}
	return result, nil
}

// getJSON performs a GET request and decodes the JSON body into dst.
func (c *Client) getJSON(url string, dst any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseSize+1))
	if err != nil {
		return err
	}
	if len(body) > MaxResponseSize {
		return fmt.Errorf("response larger than %d bytes", MaxResponseSize)
	}

	return json.Unmarshal(body, dst)
}
