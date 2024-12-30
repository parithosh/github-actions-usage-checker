package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GithubToken   string              `yaml:"github_token"`
	Actions       []string            `yaml:"actions"`
	Organizations []string            `yaml:"organizations"`
	Repositories  map[string][]string `yaml:"repositories"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Check for token in env if not in config
	if cfg.GithubToken == "" {
		cfg.GithubToken = os.Getenv("GITHUB_TOKEN")
	}

	// Validate required fields
	if cfg.GithubToken == "" {
		return nil, fmt.Errorf("github_token is required either in config file or GITHUB_TOKEN environment variable")
	}

	if len(cfg.Actions) == 0 {
		return nil, fmt.Errorf("at least one action must be specified")
	}

	if len(cfg.Organizations) == 0 && len(cfg.Repositories) == 0 {
		return nil, fmt.Errorf("at least one organization or repository must be specified")
	}

	return &cfg, nil
}
