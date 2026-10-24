package graph

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadTestRepo reads a test repository file where each line is "A B"
// (package A depends on package B). Lines starting with '#' and empty
// lines are ignored.
//
// Colons are treated as separators ("A: B" and "A:B" parse as "A B"),
// and package names are normalized to uppercase.
//
// It returns a dependency map suitable for Build:
//
//	depsFn := func(name string) []string { return deps[name] }
//	g := Build(root, opts, depsFn)
func LoadTestRepo(path string) (map[string][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open test repo: %w", err)
	}
	defer func() { _ = f.Close() }()

	deps := make(map[string][]string)
	scanner := bufio.NewScanner(f)
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.ReplaceAll(line, ":", " ")
		line = strings.ToUpper(line)

		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf(
				"test repo: line %d: expected 'A B', got %q", lineNo, line,
			)
		}

		from := fields[0]
		to := fields[1]

		deps[from] = append(deps[from], to)
		if _, ok := deps[to]; !ok {
			deps[to] = nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read test repo: %w", err)
	}
	return deps, nil
}
