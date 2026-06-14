package javagen

import (
	"fmt"
	"strings"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/i18n"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string          { return "java" }
func (g *Generator) FileExtension() string { return ".java" }

func (g *Generator) Generate(s *schema.Schema, options *codegen.GenerateOptions) ([]*codegen.GeneratedFile, error) {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("package %s;\n\n", s.Package))
	buf.WriteString("import java.util.List;\n")
	buf.WriteString("import java.util.Map;\n\n")

	for name, enum := range s.Enums {
		g.generateEnum(&buf, name, enum, options.Locale)
		buf.WriteString("\n")
	}

	for name, msg := range s.Messages {
		g.generateClass(&buf, name, msg, options.Locale)
		buf.WriteString("\n")
	}

	return []*codegen.GeneratedFile{
		{Path: "Types" + g.FileExtension(), Content: []byte(buf.String())},
	}, nil
}

func (g *Generator) generateEnum(buf *strings.Builder, name string, enum *schema.Enum, locale string) {
	if enum.Description != nil {
		desc := i18n.GetDescription(enum.Description.Zh, enum.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("/** %s */\n", desc))
		}
	}
	buf.WriteString(fmt.Sprintf("public enum %s {\n", name))
	for valueName, value := range enum.Values {
		buf.WriteString(fmt.Sprintf("\t%s(%d),\n", valueName, value))
	}
	buf.WriteString("}\n")
}

func (g *Generator) generateClass(buf *strings.Builder, name string, msg *schema.Message, locale string) {
	if msg.Description != nil {
		desc := i18n.GetDescription(msg.Description.Zh, msg.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("/** %s */\n", desc))
		}
	}
	buf.WriteString(fmt.Sprintf("public class %s {\n", name))
	for fieldName, field := range msg.Fields {
		javaType := g.mapType(field.Type)
		buf.WriteString(fmt.Sprintf("\tprivate %s %s;\n", javaType, fieldName))
	}
	buf.WriteString("}\n")
}

func (g *Generator) mapType(schemaType string) string {
	switch schemaType {
	case "bool":
		return "boolean"
	case "int32":
		return "int"
	case "int64":
		return "long"
	case "uint32":
		return "int"
	case "uint64":
		return "long"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "string":
		return "String"
	case "bytes":
		return "byte[]"
	default:
		if strings.HasPrefix(schemaType, "list<") {
			inner := strings.TrimPrefix(schemaType, "list<")
			inner = strings.TrimSuffix(inner, ">")
			return fmt.Sprintf("List<%s>", g.mapType(inner))
		}
		if strings.HasPrefix(schemaType, "map<") {
			inner := strings.TrimPrefix(schemaType, "map<")
			inner = strings.TrimSuffix(inner, ">")
			parts := strings.SplitN(inner, ",", 2)
			if len(parts) == 2 {
				keyType := g.mapType(strings.TrimSpace(parts[0]))
				valueType := g.mapType(strings.TrimSpace(parts[1]))
				return fmt.Sprintf("Map<%s, %s>", keyType, valueType)
			}
		}
		return schemaType
	}
}

func init() {
	codegen.Register(New())
}
