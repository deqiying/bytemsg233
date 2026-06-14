package schema

// Schema represents a complete .bmsg schema file
type Schema struct {
	Version  string            `yaml:"schema"`
	Package  string            `yaml:"package"`
	Messages map[string]*Message `yaml:"messages"`
	Enums    map[string]*Enum    `yaml:"enums"`
}

// Message represents a message type definition
type Message struct {
	Fields      map[string]*Field `yaml:"fields"`
	Description *Description      `yaml:"description,omitempty"`
}

// Field represents a field in a message
type Field struct {
	Type        string       `yaml:"type"`
	Description *Description `yaml:"description,omitempty"`
	Tag         int          `yaml:"tag"`
}

// Enum represents an enumeration type
type Enum struct {
	Values      map[string]int `yaml:"values"`
	Description *Description   `yaml:"description,omitempty"`
}

// Description holds i18n descriptions
type Description struct {
	Zh string `yaml:"zh,omitempty"`
	En string `yaml:"en,omitempty"`
}
