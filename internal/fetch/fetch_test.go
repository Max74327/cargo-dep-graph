package fetch_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cargo-dep-graph/internal/fetch"
)

// writeTemp writes content to a temporary file and returns its path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "Cargo.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

// --- NormalizeGithubURL ---

func TestNormalizeGithubURLPlain(t *testing.T) {
	got := fetch.NormalizeGithubURL("https://github.com/serde-rs/serde")
	want := "https://raw.githubusercontent.com/serde-rs/serde/HEAD/Cargo.toml"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeGithubURLWithGitSuffix(t *testing.T) {
	got := fetch.NormalizeGithubURL("https://github.com/serde-rs/serde.git")
	want := "https://raw.githubusercontent.com/serde-rs/serde/HEAD/Cargo.toml"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeGithubURLWithBranch(t *testing.T) {
	got := fetch.NormalizeGithubURL("https://github.com/serde-rs/serde/tree/master")
	want := "https://raw.githubusercontent.com/serde-rs/serde/master/Cargo.toml"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeGithubURLNonGithub(t *testing.T) {
	in := "https://example.com/Cargo.toml"
	if got := fetch.NormalizeGithubURL(in); got != in {
		t.Errorf("got %q, want %q", got, in)
	}
}

// --- CargoToml: local files ---

func TestCargoTomlFromPlainPath(t *testing.T) {
	path := writeTemp(t, "[dependencies]\nserde = \"1.0\"\n")
	got, err := fetch.CargoToml(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "serde") {
		t.Errorf("content = %q, want it to contain %q", got, "serde")
	}
}

func TestCargoTomlFromFileURL(t *testing.T) {
	path := writeTemp(t, "[package]\nname = \"demo\"\n")
	got, err := fetch.CargoToml("file://" + path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "demo") {
		t.Errorf("content = %q, want it to contain %q", got, "demo")
	}
}

func TestCargoTomlMissingFile(t *testing.T) {
	_, err := fetch.CargoToml("no-such-file.toml")
	if !errors.Is(err, fetch.ErrFileNotFound) {
		t.Fatalf("err = %v, want ErrFileNotFound", err)
	}
}

func TestCargoTomlFileTooBig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.toml")
	big := make([]byte, fetch.MaxTOMLSize+1)
	for i := range big {
		big[i] = 'a'
	}
	if err := os.WriteFile(path, big, 0o600); err != nil {
		t.Fatalf("write big file: %v", err)
	}
	_, err := fetch.CargoToml(path)
	if !errors.Is(err, fetch.ErrFileTooBig) {
		t.Fatalf("err = %v, want ErrFileTooBig", err)
	}
}

// --- CargoToml: HTTP via httptest ---

func TestCargoTomlHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("[dependencies]\ntokio = \"1\"\n"))
	}))
	defer srv.Close()

	got, err := fetch.CargoToml(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "tokio") {
		t.Errorf("content = %q, want it to contain %q", got, "tokio")
	}
}

func TestCargoTomlHTTPStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := fetch.CargoToml(srv.URL)
	if !errors.Is(err, fetch.ErrHTTPStatus) {
		t.Fatalf("err = %v, want ErrHTTPStatus", err)
	}
}

func TestCargoTomlHTTPTooBig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		big := make([]byte, fetch.MaxTOMLSize+1)
		_, _ = w.Write(big)
	}))
	defer srv.Close()

	_, err := fetch.CargoToml(srv.URL)
	if !errors.Is(err, fetch.ErrFileTooBig) {
		t.Fatalf("err = %v, want ErrFileTooBig", err)
	}
}
