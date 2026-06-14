package gocodegen

import (
	"fmt"
	"strings"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/i18n"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string          { return "go" }
func (g *Generator) FileExtension() string { return ".go" }

func (g *Generator) Generate(s *schema.Schema, options *codegen.GenerateOptions) ([]*codegen.GeneratedFile, error) {
	var buf strings.Builder

	packageName := s.Package
	if options.Package != "" {
		packageName = options.Package
	}
	buf.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	for name, enum := range s.Enums {
		g.generateEnum(&buf, name, enum, options.Locale)
		buf.WriteString("\n")
	}

	for name, msg := range s.Messages {
		g.generateMessage(&buf, name, msg, options.Locale)
		buf.WriteString("\n")
	}

	return []*codegen.GeneratedFile{
		{Path: "types" + g.FileExtension(), Content: []byte(buf.String())},
	}, nil
}

func (g *Generator) generateEnum(buf *strings.Builder, name string, enum *schema.Enum, locale string) {
	if enum.Description != nil {
		desc := i18n.GetDescription(enum.Description.Zh, enum.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("// %s\n", desc))
		}
	}

	buf.WriteString(fmt.Sprintf("type %s int32\n\n", name))
	buf.WriteString("const (\n")

	for valueName, value := range enum.Values {
		buf.WriteString(fmt.Sprintf("\t%s%s %s = %d\n", name, valueName, name, value))
	}

	buf.WriteString(")\n")
}

func (g *Generator) generateMessage(buf *strings.Builder, name string, msg *schema.Message, locale string) {
	if msg.Description != nil {
		desc := i18n.GetDescription(msg.Description.Zh, msg.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("// %s\n", desc))
		}
	}

	buf.WriteString(fmt.Sprintf("type %s struct {\n", name))

	for fieldName, field := range msg.Fields {
		if field.Description != nil {
			desc := i18n.GetDescription(field.Description.Zh, field.Description.En)
			if desc != "" {
				buf.WriteString(fmt.Sprintf("\t// %s\n", desc))
			}
		}

		goType := g.mapType(field.Type)
		goName := toGoName(fieldName)
		buf.WriteString(fmt.Sprintf("\t%s %s `bytemsg:\"%d\"`\n", goName, goType, field.Tag))
	}

	buf.WriteString("}\n")
}

func (g *Generator) mapType(schemaType string) string {
	switch schemaType {
	case "bool":
		return "bool"
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "uint32":
		return "uint32"
	case "uint64":
		return "uint64"
	case "float32":
		return "float32"
	case "float64":
		return "float64"
	case "string":
		return "string"
	case "bytes":
		return "[]byte"
	default:
		if strings.HasPrefix(schemaType, "list<") {
			inner := strings.TrimPrefix(schemaType, "list<")
			inner = strings.TrimSuffix(inner, ">")
			return "[]" + g.mapType(inner)
		}
		if strings.HasPrefix(schemaType, "map<") {
			inner := strings.TrimPrefix(schemaType, "map<")
			inner = strings.TrimSuffix(inner, ">")
			parts := strings.SplitN(inner, ",", 2)
			if len(parts) == 2 {
				keyType := g.mapType(strings.TrimSpace(parts[0]))
				valueType := g.mapType(strings.TrimSpace(parts[1]))
				return fmt.Sprintf("map[%s]%s", keyType, valueType)
			}
		}
		return schemaType
	}
}

func toGoName(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func init() {
	codegen.Register(New())
}
