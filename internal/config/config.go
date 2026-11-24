// Package config loads the static repo allowlist used to filter PRs.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the list of repos ("owner/name") to show PRs for.
type Config struct {
	Repos []string `yaml:"repos"`
}

// DefaultPath returns the default config file location:
// $XDG_CONFIG_HOME/github-pr-summarizer/repos.yaml (or the OS equivalent).
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve default config dir: %w", err)
	}
	return filepath.Join(dir, "github-pr-summarizer", "repos.yaml"), nil
}

// Load reads and validates the config file at path.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config file %q: %w", path, err)
	}

	if len(cfg.Repos) == 0 {
		return Config{}, fmt.Errorf("config file %q: repos list is empty", path)
	}

	for _, r := range cfg.Repos {
		if !validRepo(r) {
			return Config{}, fmt.Errorf("config file %q: invalid repo entry %q, expected owner/name", path, r)
		}
	}

	return cfg, nil
}

// FromRepo returns a config that holds only the given "owner/name" repo.
func FromRepo(repo string) (Config, error) {
	if !validRepo(repo) {
		return Config{}, fmt.Errorf("invalid repo %q, expected owner/name", repo)
	}
	return Config{Repos: []string{repo}}, nil
}

func validRepo(r string) bool {
	parts := strings.Split(r, "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

// Allowlist returns the configured repos as an "owner/name" set.
func (c Config) Allowlist() map[string]struct{} {
	set := make(map[string]struct{}, len(c.Repos))
	for _, r := range c.Repos {
		set[r] = struct{}{}
	}
	return set
}
