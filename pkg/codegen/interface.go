package codegen

import (
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

// CodeGenerator defines the interface for language-specific code generators
type CodeGenerator interface {
	// Name returns the generator name (e.g., "go", "csharp", "java")
	Name() string

	// FileExtension returns the file extension for generated files
	FileExtension() string

	// Generate generates code from a schema
	Generate(schema *schema.Schema, options *GenerateOptions) ([]*GeneratedFile, error)
}

// GenerateOptions contains options for code generation
type GenerateOptions struct {
	OutputDir string
	Locale    string
	Package   string
}

// GeneratedFile represents a generated file
type GeneratedFile struct {
	Path    string
	Content []byte
}
