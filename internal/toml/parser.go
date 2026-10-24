// Package toml implements a minimal TOML parser for Cargo.toml manifests.
package toml

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseError describes a TOML parsing failure.
type ParseError struct {
	Line int
	Msg  string
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	return fmt.Sprintf("toml: line %d: %s", e.Line, e.Msg)
}

// Document is the result of parsing a TOML file.
type Document struct {
	tables map[string]map[string]Value
}

// GetTable returns the named section and whether it exists.
func (d *Document) GetTable(name string) (map[string]Value, bool) {
	t, ok := d.tables[name]
	return t, ok
}

// Value is a parsed TOML value.
type Value struct {
	Str   string
	Arr   []Value
	Table map[string]Value
	IsArr bool
	IsTbl bool
}

// Parse parses a TOML document. Only sections and key = value pairs are supported.
func Parse(input string) (*Document, error) {
	doc := &Document{tables: map[string]map[string]Value{}}
	current := map[string]Value{}
	doc.tables[""] = current

	for i, raw := range strings.Split(input, "\n") {
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return nil, &ParseError{Line: i + 1, Msg: "unterminated section header"}
			}
			name := strings.TrimSpace(line[1 : len(line)-1])
			if _, ok := doc.tables[name]; !ok {
				doc.tables[name] = map[string]Value{}
			}
			current = doc.tables[name]
			continue
		}

		key, valText, ok := strings.Cut(line, "=")
		if !ok {
			return nil, &ParseError{Line: i + 1, Msg: "expected key = value"}
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, &ParseError{Line: i + 1, Msg: "empty key"}
		}
		val, err := parseValue(strings.TrimSpace(valText))
		if err != nil {
			return nil, &ParseError{Line: i + 1, Msg: err.Error()}
		}
		current[key] = val
	}
	return doc, nil
}

// stripComment removes a '#' comment that is not inside quotes.
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

// parseValue parses a single TOML value: string, array, inline table, bool or number.
func parseValue(s string) (Value, error) {
	if s == "" {
		return Value{Str: ""}, nil
	}
	if s[0] == '"' || s[0] == '\'' {
		return parseString(s)
	}
	if s[0] == '[' {
		return parseArray(s)
	}
	if s[0] == '{' {
		return parseInlineTable(s)
	}
	return Value{Str: s}, nil
}

func parseString(s string) (Value, error) {
	quote := s[0]
	if len(s) < 2 || s[len(s)-1] != quote {
		return Value{}, fmt.Errorf("unterminated string")
	}
	body := s[1 : len(s)-1]
	if quote == '"' {
		unq, err := strconv.Unquote(s)
		if err != nil {
			return Value{}, fmt.Errorf("invalid string: %v", err)
		}
		return Value{Str: unq}, nil
	}
	return Value{Str: strings.ReplaceAll(body, "''", "'")}, nil
}

func parseArray(s string) (Value, error) {
	if len(s) < 2 || s[len(s)-1] != ']' {
		return Value{}, fmt.Errorf("unterminated array")
	}
	body := strings.TrimSpace(s[1 : len(s)-1])
	if body == "" {
		return Value{IsArr: true}, nil
	}
	parts := splitTopLevel(body, ',')
	arr := make([]Value, 0, len(parts))
	for _, p := range parts {
		v, err := parseValue(strings.TrimSpace(p))
		if err != nil {
			return Value{}, err
		}
		arr = append(arr, v)
	}
	return Value{Arr: arr, IsArr: true}, nil
}

func parseInlineTable(s string) (Value, error) {
	if len(s) < 2 || s[len(s)-1] != '}' {
		return Value{}, fmt.Errorf("unterminated inline table")
	}
	body := strings.TrimSpace(s[1 : len(s)-1])
	tbl := map[string]Value{}
	if body == "" {
		return Value{Table: tbl, IsTbl: true}, nil
	}
	parts := splitTopLevel(body, ',')
	for _, p := range parts {
		k, vText, ok := strings.Cut(p, "=")
		if !ok {
			return Value{}, fmt.Errorf("expected key = value in inline table")
		}
		key := strings.TrimSpace(k)
		if key == "" {
			return Value{}, fmt.Errorf("empty key in inline table")
		}
		v, err := parseValue(strings.TrimSpace(vText))
		if err != nil {
			return Value{}, err
		}
		tbl[key] = v
	}
	return Value{Table: tbl, IsTbl: true}, nil
}

// splitTopLevel splits s on sep, but only at the top level
// (not inside quotes, brackets, or braces).
func splitTopLevel(s string, sep byte) []string {
	var parts []string
	depth := 0
	inSingle, inDouble := false, false
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\'' && !inDouble:
			inSingle = !inSingle
		case c == '"' && !inSingle:
			inDouble = !inDouble
		case inSingle || inDouble:
			// skip
		case c == '[' || c == '{':
			depth++
		case c == ']' || c == '}':
			depth--
		case c == sep && depth == 0:
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}
