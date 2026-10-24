// Package fetch retrieves Cargo.toml from a repository URL or a local file.
package fetch

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// MaxTOMLSize is the upper bound for the manifest size (1 MiB).
const MaxTOMLSize = 1 << 20

// Fetch errors.
var (
	ErrFileNotFound = errors.New("manifest file not found")
	ErrFileTooBig   = errors.New("manifest file too big")
	ErrHTTPStatus   = errors.New("unexpected HTTP status")
)

// httpClient is used for all HTTP requests; the timeout prevents hanging.
var httpClient = &http.Client{Timeout: 15 * time.Second}

// FetchCargoToml returns the text of Cargo.toml from a URL or local path.
func FetchCargoToml(source string) (string, error) {
	switch {
	case strings.HasPrefix(source, "http://"), strings.HasPrefix(source, "https://"):
		return fetchHTTP(source)
	case strings.HasPrefix(source, "file://"):
		return readFile(strings.TrimPrefix(source, "file://"))
	default:
		return readFile(source)
	}
}

func readFile(path string) (string, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("%w: %s", ErrFileNotFound, path)
	}
	if err != nil {
		return "", err
	}
	if info.Size() > MaxTOMLSize {
		return "", fmt.Errorf("%w: %s", ErrFileTooBig, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func fetchHTTP(url string) (string, error) {
	rawURL := NormalizeGithubURL(url)
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "cargo-dep-graph/0.2")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %d for %s", ErrHTTPStatus, resp.StatusCode, rawURL)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxTOMLSize+1))
	if err != nil {
		return "", err
	}
	if len(body) > MaxTOMLSize {
		return "", fmt.Errorf("%w: response larger than %d bytes", ErrFileTooBig, MaxTOMLSize)
	}
	return string(body), nil
}

// NormalizeGithubURL converts github.com/u/r[/tree/b] to raw.githubusercontent.com/u/r/[b/]Cargo.toml.
func NormalizeGithubURL(url string) string {
	const prefix = "https://github.com/"
	if !strings.HasPrefix(url, prefix) {
		return url
	}
	rest := strings.TrimSuffix(strings.TrimPrefix(url, prefix), ".git")
	if repo, branch, found := strings.Cut(rest, "/tree/"); found {
		return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/Cargo.toml", repo, branch)
	}
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/HEAD/Cargo.toml", rest)
}
