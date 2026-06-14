package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseFile parses a schema file by extension
func ParseFile(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".bmsg":
		return ParseBmsg(data)
	case ".yaml", ".yml":
		return ParseYAML(data)
	case ".json":
		return ParseJSON(data)
	case ".toml":
		return ParseTOML(data)
	default:
		// Try .bmsg first, then YAML
		s, err := ParseBmsg(data)
		if err == nil {
			return s, nil
		}
		return ParseYAML(data)
	}
}

// ParseYAML parses YAML content
func ParseYAML(data []byte) (*Schema, error) {
	var schema Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	if err := validate(&schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

// ParseJSON parses JSON content
func ParseJSON(data []byte) (*Schema, error) {
	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	if err := validate(&schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

// ParseTOML parses TOML content
func ParseTOML(data []byte) (*Schema, error) {
	// TOML support: convert to YAML internally
	// Simple approach: parse TOML line by line and build Schema
	// For now, require explicit format flag
	return nil, fmt.Errorf("toml support coming soon, use .bmsg or .yaml")
}
