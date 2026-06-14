package schema

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Parse parses YAML content into a Schema
func Parse(data []byte) (*Schema, error) {
	var schema Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	if err := validate(&schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

// validate performs basic validation on the schema
func validate(s *Schema) error {
	if s.Version == "" {
		return fmt.Errorf("schema version is required")
	}

	if s.Messages == nil {
		s.Messages = make(map[string]*Message)
	}

	if s.Enums == nil {
		s.Enums = make(map[string]*Enum)
	}

	for name, msg := range s.Messages {
		if msg.Fields == nil {
			msg.Fields = make(map[string]*Field)
		}
		for fieldName, field := range msg.Fields {
			if field.Tag <= 0 {
				return fmt.Errorf("message %s, field %s: tag must be positive", name, fieldName)
			}
		}
	}

	return nil
}
