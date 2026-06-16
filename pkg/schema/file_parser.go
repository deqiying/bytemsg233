package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseFile parses a schema file by extension.
//
// JSON is the default DSL. YAML remains supported, and legacy .bmsg syntax is
// accepted as a compatibility/export target for future tooling.
func ParseFile(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".bmsg":
		s, err := ParseJSON(data)
		if err == nil {
			return s, nil
		}
		s, err = ParseYAML(data)
		if err == nil {
			return s, nil
		}
		return ParseBmsg(data)
	case ".yaml", ".yml":
		return ParseYAML(data)
	case ".json":
		return ParseJSON(data)
	case ".toml":
		return ParseTOML(data)
	default:
		s, err := ParseJSON(data)
		if err == nil {
			return s, nil
		}
		s, err = ParseYAML(data)
		if err == nil {
			return s, nil
		}
		return ParseBmsg(data)
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
	if len(schema.Messages) == 0 {
		native, err := parseNativeJSON(data)
		if err != nil {
			return nil, err
		}
		schema = *native
	}
	if err := validate(&schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

func parseNativeJSON(data []byte) (*Schema, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	s := &Schema{
		Messages: make(map[string]*Message),
		Enums:    make(map[string]*Enum),
	}

	if raw, ok := root["schema"]; ok {
		_ = json.Unmarshal(raw, &s.Version)
	}
	if raw, ok := root["package"]; ok {
		_ = json.Unmarshal(raw, &s.Package)
	}
	if raw, ok := root["enums"]; ok {
		if err := json.Unmarshal(raw, &s.Enums); err != nil {
			return nil, fmt.Errorf("parse json enums: %w", err)
		}
	}

	for name, raw := range root {
		if isReservedNativeJSONKey(name) {
			continue
		}

		msg, err := parseNativeJSONMessage(raw)
		if err != nil {
			return nil, fmt.Errorf("parse json message %s: %w", name, err)
		}
		s.Messages[name] = msg
	}

	return s, nil
}

func parseNativeJSONMessage(raw json.RawMessage) (*Message, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}

	msg := &Message{Fields: make(map[string]*Field)}
	if rawDesc, ok := obj["description"]; ok {
		var desc Description
		if err := json.Unmarshal(rawDesc, &desc); err != nil {
			return nil, fmt.Errorf("description: %w", err)
		}
		msg.Description = &desc
	}

	if rawFields, ok := obj["fields"]; ok {
		if err := json.Unmarshal(rawFields, &msg.Fields); err != nil {
			return nil, fmt.Errorf("fields: %w", err)
		}
		return msg, nil
	}

	for fieldName, rawField := range obj {
		if fieldName == "description" {
			continue
		}
		field, err := parseNativeJSONField(rawField)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", fieldName, err)
		}
		msg.Fields[fieldName] = field
	}

	return msg, nil
}

func parseNativeJSONField(raw json.RawMessage) (*Field, error) {
	var field Field
	if err := json.Unmarshal(raw, &field); err == nil && field.Type != "" {
		return &field, nil
	}

	var fieldType string
	if err := json.Unmarshal(raw, &fieldType); err == nil && fieldType != "" {
		return &Field{Type: fieldType}, nil
	}

	return nil, fmt.Errorf("expected field object with type/tag")
}

func isReservedNativeJSONKey(key string) bool {
	switch key {
	case "schema", "$schema", "package", "namespace", "messages", "enums":
		return true
	default:
		return false
	}
}

// ParseTOML parses TOML content
func ParseTOML(data []byte) (*Schema, error) {
	// TOML support: convert to YAML internally
	// Simple approach: parse TOML line by line and build Schema
	// For now, require explicit format flag
	return nil, fmt.Errorf("toml support coming soon, use .bmsg.json, .json, or .yaml")
}
