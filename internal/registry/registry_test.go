package registry_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"cargo-dep-graph/internal/registry"
)

// newTestClient creates a Client pointing at a test server.
func newTestClient(t *testing.T, handler http.Handler) *registry.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return registry.NewClientWithBaseURL(srv.URL)
}

// --- Basic fetch ---

func TestFetchDependencies(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/crates/serde", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"crate":{"max_version":"1.0.219"}}`))
	})
	mux.HandleFunc("/api/v1/crates/serde/1.0.219/dependencies", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"dependencies":[
				{"crate_id":"serde_derive","kind":"normal"},
				{"crate_id":"criterion","kind":"dev"},
				{"crate_id":"cc","kind":"build"}
			]
		}`))
	})

	c := newTestClient(t, mux)
	deps, err := c.FetchDependencies("serde")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 1 || deps[0] != "serde_derive" {
		t.Errorf("deps = %v, want [serde_derive]", deps)
	}
}

// --- Cache behaviour ---

func TestFetchDependenciesCache(t *testing.T) {
	requests := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/crates/serde", func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"crate":{"max_version":"1.0.0"}}`))
	})
	mux.HandleFunc("/api/v1/crates/serde/1.0.0/dependencies", func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"dependencies":[]}`))
	})

	c := newTestClient(t, mux)
	if _, err := c.FetchDependencies("serde"); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if _, err := c.FetchDependencies("serde"); err != nil {
		t.Fatalf("second call: %v", err)
	}

	// Two calls, but only one HTTP round-trip pair (version + deps).
	if requests != 2 {
		t.Errorf("HTTP requests = %d, want 2 (cache must skip second fetch)", requests)
	}
}

// --- Error handling ---

func TestFetchDependenciesNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/crates/ghost", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	c := newTestClient(t, mux)
	_, err := c.FetchDependencies("ghost")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestFetchDependenciesEmptyVersion(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/crates/empty", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"crate":{"max_version":""}}`))
	})

	c := newTestClient(t, mux)
	_, err := c.FetchDependencies("empty")
	if err == nil {
		t.Fatal("expected error for empty version")
	}
}

func TestFetchDependenciesNoNormal(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/crates/x", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"crate":{"max_version":"0.1.0"}}`))
	})
	mux.HandleFunc("/api/v1/crates/x/0.1.0/dependencies", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dependencies":[{"crate_id":"y","kind":"dev"}]}`))
	})

	c := newTestClient(t, mux)
	deps, err := c.FetchDependencies("x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 0 {
		t.Errorf("deps = %v, want empty (only dev filtered)", deps)
	}
}
