package csharpgen

import (
	"fmt"
	"strings"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/i18n"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string          { return "csharp" }
func (g *Generator) FileExtension() string { return ".cs" }

func (g *Generator) Generate(s *schema.Schema, options *codegen.GenerateOptions) ([]*codegen.GeneratedFile, error) {
	var buf strings.Builder

	namespace := s.Package
	if options.Package != "" {
		namespace = options.Package
	}
	buf.WriteString(fmt.Sprintf("namespace %s\n{\n", namespace))

	for name, enum := range s.Enums {
		g.generateEnum(&buf, name, enum, options.Locale, 1)
		buf.WriteString("\n")
	}

	for name, msg := range s.Messages {
		g.generateClass(&buf, name, msg, options.Locale, 1)
		buf.WriteString("\n")
	}

	buf.WriteString("}\n")

	return []*codegen.GeneratedFile{
		{Path: "Types" + g.FileExtension(), Content: []byte(buf.String())},
	}, nil
}

func (g *Generator) generateEnum(buf *strings.Builder, name string, enum *schema.Enum, locale string, indent int) {
	indentStr := strings.Repeat("\t", indent)

	if enum.Description != nil {
		desc := i18n.GetDescription(enum.Description.Zh, enum.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("%s/// <summary>\n", indentStr))
			buf.WriteString(fmt.Sprintf("%s/// %s\n", indentStr, desc))
			buf.WriteString(fmt.Sprintf("%s/// </summary>\n", indentStr))
		}
	}

	buf.WriteString(fmt.Sprintf("%spublic enum %s\n", indentStr, name))
	buf.WriteString(fmt.Sprintf("%s{\n", indentStr))

	for valueName, value := range enum.Values {
		buf.WriteString(fmt.Sprintf("%s\t%s = %d,\n", indentStr, valueName, value))
	}

	buf.WriteString(fmt.Sprintf("%s}\n", indentStr))
}

func (g *Generator) generateClass(buf *strings.Builder, name string, msg *schema.Message, locale string, indent int) {
	indentStr := strings.Repeat("\t", indent)

	if msg.Description != nil {
		desc := i18n.GetDescription(msg.Description.Zh, msg.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("%s/// <summary>\n", indentStr))
			buf.WriteString(fmt.Sprintf("%s/// %s\n", indentStr, desc))
			buf.WriteString(fmt.Sprintf("%s/// </summary>\n", indentStr))
		}
	}

	buf.WriteString(fmt.Sprintf("%spublic class %s\n", indentStr, name))
	buf.WriteString(fmt.Sprintf("%s{\n", indentStr))

	for fieldName, field := range msg.Fields {
		if field.Description != nil {
			desc := i18n.GetDescription(field.Description.Zh, field.Description.En)
			if desc != "" {
				buf.WriteString(fmt.Sprintf("%s\t/// <summary>\n", indentStr))
				buf.WriteString(fmt.Sprintf("%s\t/// %s\n", indentStr, desc))
				buf.WriteString(fmt.Sprintf("%s\t/// </summary>\n", indentStr))
			}
		}

		csharpType := g.mapType(field.Type)
		csharpName := toCSharpName(fieldName)
		buf.WriteString(fmt.Sprintf("%s\tpublic %s %s { get; set; }\n", indentStr, csharpType, csharpName))
	}

	buf.WriteString(fmt.Sprintf("%s}\n", indentStr))
}

func (g *Generator) mapType(schemaType string) string {
	switch schemaType {
	case "bool":
		return "bool"
	case "int32":
		return "int"
	case "int64":
		return "long"
	case "uint32":
		return "uint"
	case "uint64":
		return "ulong"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "string":
		return "string"
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
				return fmt.Sprintf("Dictionary<%s, %s>", keyType, valueType)
			}
		}
		return schemaType
	}
}

func toCSharpName(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func init() {
	codegen.Register(New())
}
