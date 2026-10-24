// Package config implements a minimal YAML parser for flat configs.
package config

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseError describes a YAML parsing failure.
type ParseError struct {
	Line int
	Msg  string
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	return fmt.Sprintf("yaml: line %d: %s", e.Line, e.Msg)
}

// Node is a flat YAML mapping from keys to raw string values.
type Node struct {
	values map[string]string
}

// String returns the value for key, or "" if absent.
func (n *Node) String(key string) string {
	if n == nil {
		return ""
	}
	return n.values[key]
}

// Bool returns the value for key as a bool, or false if absent or invalid.
func (n *Node) Bool(key string) bool {
	v := strings.ToLower(strings.TrimSpace(n.String(key)))
	return v == "true"
}

// Int returns the value for key as an int, or def if absent or invalid.
func (n *Node) Int(key string, def int) int {
	if n == nil {
		return def
	}
	raw, ok := n.values[key]
	if !ok {
		return def
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return def
	}
	return v
}

// ParseYAML parses a flat YAML document into a Node.
// Only top-level "key: value" lines and comments are supported.
func ParseYAML(input string) (*Node, error) {
	n := &Node{values: make(map[string]string)}
	lines := strings.Split(input, "\n")

	for i, raw := range lines {
		line := stripComment(raw)
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		colon := strings.Index(line, ":")
		if colon < 0 {
			return nil, &ParseError{Line: i + 1, Msg: "expected 'key: value'"}
		}

		key := strings.TrimSpace(line[:colon])
		if key == "" {
			return nil, &ParseError{Line: i + 1, Msg: "empty key"}
		}

		value, err := parseScalar(strings.TrimSpace(line[colon+1:]))
		if err != nil {
			return nil, &ParseError{Line: i + 1, Msg: err.Error()}
		}
		n.values[key] = value
	}
	return n, nil
}

// stripComment removes a trailing '#' comment that is not inside quotes.
func stripComment(line string) string {
	inSingle, inDouble := false, false
	for i, r := range line {
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				return line[:i]
			}
		}
	}
	return line
}

// parseScalar normalizes a YAML scalar: unquotes strings and rejects
// unsupported constructs such as lists and nested maps.
func parseScalar(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	if s[0] == '[' || s[0] == '{' {
		return "", fmt.Errorf("lists and maps are not supported")
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		unq, err := strconv.Unquote(s)
		if err != nil {
			return "", fmt.Errorf("invalid double-quoted string: %v", err)
		}
		return unq, nil
	}
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return strings.ReplaceAll(s[1:len(s)-1], "''", "'"), nil
	}
	return s, nil
}
