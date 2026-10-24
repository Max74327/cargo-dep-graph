package registry

// CrateResponse is the top-level response from /api/v1/crates/{name}.
type CrateResponse struct {
	Crate Crate `json:"crate"`
}

// Crate holds the crate metadata; we only need the latest version.
type Crate struct {
	MaxVersion string `json:"max_version"`
}

// DependenciesResponse is the response from /api/v1/crates/{name}/{version}/dependencies.
type DependenciesResponse struct {
	Dependencies []Dependency `json:"dependencies"`
}

// Dependency is a single dependency entry.
type Dependency struct {
	CrateName string `json:"crate_id"` // name of the dependency
	Kind      string `json:"kind"`     // "normal", "dev", "build"
	Optional  bool   `json:"optional"`
}
