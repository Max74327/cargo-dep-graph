package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const MaxConfigSize = 64 * 1024

var (
	ErrFileNotFound    = errors.New("configuration file not found")
	ErrFileTooBig      = errors.New("configuration file too big")
	ErrInvalidYAML     = errors.New("invalid YAML")
	ErrEmptyPackage    = errors.New("package_name must not be empty")
	ErrEmptyRepo       = errors.New("repository_url must not be empty")
	ErrMissingTestPath = errors.New("test_repo_path is required when test_mode=true")
)

type Config struct {
	PackageName     string
	RepositoryURL   string
	TestMode        bool
	TestRepoPath    string
	ASCIITree       bool
	FilterSubstring string
}

// Load читает YAML-файл и валидирует параметры.
func Load(path string) (*Config, error) {
	data, err := readLimited(path)
	if err != nil {
		return nil, err
	}

	raw, err := ParseYAML(string(data))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidYAML, err)
	}

	cfg := &Config{
		PackageName:     raw.String("package_name"),
		RepositoryURL:   raw.String("repository_url"),
		TestMode:        raw.Bool("test_mode"),
		TestRepoPath:    raw.String("test_repo_path"),
		ASCIITree:       raw.Bool("ascii_tree"),
		FilterSubstring: raw.String("filter_substring"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func readLimited(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", ErrFileNotFound, path)
	}
	if err != nil {
		return nil, err
	}
	if info.Size() > MaxConfigSize {
		return nil, fmt.Errorf("%w: %s", ErrFileTooBig, path)
	}
	return os.ReadFile(path)
}

func (c *Config) validate() error {
	if strings.TrimSpace(c.PackageName) == "" {
		return ErrEmptyPackage
	}
	if !c.TestMode && strings.TrimSpace(c.RepositoryURL) == "" {
		return ErrEmptyRepo
	}
	if c.TestMode && strings.TrimSpace(c.TestRepoPath) == "" {
		return ErrMissingTestPath
	}
	return nil
}

// AsMap возвращает параметры в порядке для вывода «ключ = значение».
func (c *Config) AsMap() [][2]string {
	return [][2]string{
		{"package_name", c.PackageName},
		{"repository_url", c.RepositoryURL},
		{"test_mode", fmt.Sprintf("%t", c.TestMode)},
		{"test_repo_path", c.TestRepoPath},
		{"ascii_tree", fmt.Sprintf("%t", c.ASCIITree)},
		{"filter_substring", c.FilterSubstring},
	}
}
