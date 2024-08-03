package config

import (
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	// Act
	reader := strings.NewReader(`files: ["**/*.go"]`)

	// Arrange
	config, err := NewFromReader(reader)
	if err != nil {
		t.Errorf("Error parsing YAML: %s", err)
	}

	// Assert
	if len(config.Files) != 1 || config.Files[0] != "**/*.go" {
		t.Errorf("Expected config to have 1 file, got %d", len(config.Files))
	}
}
