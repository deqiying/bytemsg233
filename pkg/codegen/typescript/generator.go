package tsgen

import (
	"fmt"
	"strings"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/i18n"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string          { return "typescript" }
func (g *Generator) FileExtension() string { return ".ts" }

func (g *Generator) Generate(s *schema.Schema, options *codegen.GenerateOptions) ([]*codegen.GeneratedFile, error) {
	var buf strings.Builder

	for name, enum := range s.Enums {
		g.generateEnum(&buf, name, enum, options.Locale)
		buf.WriteString("\n")
	}

	for name, msg := range s.Messages {
		g.generateInterface(&buf, name, msg, options.Locale)
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
			buf.WriteString(fmt.Sprintf("/** %s */\n", desc))
		}
	}
	buf.WriteString(fmt.Sprintf("export enum %s {\n", name))
	for valueName, value := range enum.Values {
		buf.WriteString(fmt.Sprintf("\t%s = %d,\n", valueName, value))
	}
	buf.WriteString("}\n")
}

func (g *Generator) generateInterface(buf *strings.Builder, name string, msg *schema.Message, locale string) {
	if msg.Description != nil {
		desc := i18n.GetDescription(msg.Description.Zh, msg.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("/** %s */\n", desc))
		}
	}
	buf.WriteString(fmt.Sprintf("export interface %s {\n", name))
	for fieldName, field := range msg.Fields {
		if field.Description != nil {
			desc := i18n.GetDescription(field.Description.Zh, field.Description.En)
			if desc != "" {
				buf.WriteString(fmt.Sprintf("\t/** %s */\n", desc))
			}
		}
		tsType := g.mapType(field.Type)
		buf.WriteString(fmt.Sprintf("\t%s: %s;\n", fieldName, tsType))
	}
	buf.WriteString("}\n")
}

func (g *Generator) mapType(schemaType string) string {
	switch schemaType {
	case "bool":
		return "boolean"
	case "int32", "int64", "uint32", "uint64", "float32", "float64":
		return "number"
	case "string":
		return "string"
	case "bytes":
		return "Uint8Array"
	default:
		if strings.HasPrefix(schemaType, "list<") {
			inner := strings.TrimPrefix(schemaType, "list<")
			inner = strings.TrimSuffix(inner, ">")
			return fmt.Sprintf("%s[]", g.mapType(inner))
		}
		if strings.HasPrefix(schemaType, "map<") {
			inner := strings.TrimPrefix(schemaType, "map<")
			inner = strings.TrimSuffix(inner, ">")
			parts := strings.SplitN(inner, ",", 2)
			if len(parts) == 2 {
				keyType := g.mapType(strings.TrimSpace(parts[0]))
				valueType := g.mapType(strings.TrimSpace(parts[1]))
				return fmt.Sprintf("Record<%s, %s>", keyType, valueType)
			}
		}
		return schemaType
	}
}

func init() {
	codegen.Register(New())
}
