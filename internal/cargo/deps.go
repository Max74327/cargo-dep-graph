// Package cargo extracts direct dependencies from a parsed Cargo.toml.
package cargo

import (
	"errors"
	"fmt"
	"sort"

	"cargo-dep-graph/internal/toml"
)

// Dependency is a direct dependency of a Cargo package.
type Dependency struct {
	Name string
}

// Extraction errors.
var (
	ErrNoDependencies = errors.New("section [dependencies] not found")
	ErrInvalidDeps    = errors.New("[dependencies] is not a table")
)

// ExtractDirectDependencies returns direct dependencies from [dependencies].
func ExtractDirectDependencies(doc *toml.Document) ([]Dependency, error) {
	tbl, ok := doc.GetTable("dependencies")
	if !ok {
		return nil, ErrNoDependencies
	}
	if tbl == nil {
		return nil, fmt.Errorf("%w: [dependencies] is nil", ErrInvalidDeps)
	}

	deps := make([]Dependency, 0, len(tbl))
	for name := range tbl {
		deps = append(deps, Dependency{Name: name})
	}
	sort.Slice(deps, func(i, j int) bool { return deps[i].Name < deps[j].Name })
	return deps, nil
}
