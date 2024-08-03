package config

import (
	"io"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Files     []string `yaml:"files"`
	OutputDir string   `yaml:"outputDirectory"`
	AddonName string   `yaml:"name"`
}

// Parse config from a reader
func NewFromReader(reader io.Reader) (*Config, error) {
	contents, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Unmarshal YAML
	var config Config
	err = yaml.Unmarshal(contents, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
