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

	prevLocale := i18n.GetLocale()
	if options != nil && options.Locale != "" {
		i18n.SetLocale(options.Locale)
		defer i18n.SetLocale(prevLocale)
	}

	namespace := s.Package
	if options != nil && options.Package != "" {
		namespace = options.Package
	}

	buf.WriteString("using System;\n")
	buf.WriteString("using System.Collections.Concurrent;\n")
	buf.WriteString("using System.Collections.Generic;\n\n")
	buf.WriteString(fmt.Sprintf("namespace %s\n{\n", namespace))

	for _, name := range codegen.SortedEnumNames(s) {
		g.generateEnum(&buf, name, s.Enums[name], 1)
		buf.WriteString("\n")
	}

	for _, name := range codegen.SortedMessageNames(s) {
		g.generateClass(&buf, s, name, s.Messages[name], 1)
		buf.WriteString("\n")
	}

	buf.WriteString("}\n")

	return []*codegen.GeneratedFile{
		{Path: "Types" + g.FileExtension(), Content: []byte(buf.String())},
	}, nil
}

func (g *Generator) generateEnum(buf *strings.Builder, name string, enum *schema.Enum, indent int) {
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
	values := codegen.SortedEnumValues(enum)
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("%s\t%s = %d,\n", indentStr, codegen.ToPascalCase(value.Name), value.Value))
	}
	buf.WriteString(fmt.Sprintf("%s}\n\n", indentStr))

	buf.WriteString(fmt.Sprintf("%spublic static class %sExtensions\n", indentStr, name))
	buf.WriteString(fmt.Sprintf("%s{\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\tpublic static %s FromValue(int raw)\n", indentStr, name))
	buf.WriteString(fmt.Sprintf("%s\t{\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\treturn raw switch\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\t{\n", indentStr))
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("%s\t\t\t%d => %s.%s,\n", indentStr, value.Value, name, codegen.ToPascalCase(value.Name)))
	}
	buf.WriteString(fmt.Sprintf("%s\t\t\t_ => throw new ArgumentOutOfRangeException(nameof(raw), raw, \"Unknown enum value\")\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\t};\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t}\n\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\tpublic static bool IsDefinedValue(this %s value)\n", indentStr, name))
	buf.WriteString(fmt.Sprintf("%s\t{\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\treturn value switch\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\t{\n", indentStr))
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("%s\t\t\t%s.%s => true,\n", indentStr, name, codegen.ToPascalCase(value.Name)))
	}
	buf.WriteString(fmt.Sprintf("%s\t\t\t_ => false,\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\t};\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t}\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s}\n", indentStr))
}

func (g *Generator) generateClass(buf *strings.Builder, s *schema.Schema, name string, msg *schema.Message, indent int) {
	indentStr := strings.Repeat("\t", indent)

	if msg.Description != nil {
		desc := i18n.GetDescription(msg.Description.Zh, msg.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("%s/// <summary>\n", indentStr))
			buf.WriteString(fmt.Sprintf("%s/// %s\n", indentStr, desc))
			buf.WriteString(fmt.Sprintf("%s/// </summary>\n", indentStr))
		}
	}

	buf.WriteString(fmt.Sprintf("%spublic sealed class %s\n", indentStr, name))
	buf.WriteString(fmt.Sprintf("%s{\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\tprivate static readonly ConcurrentBag<%s> Pool = new();\n\n", indentStr, name))

	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		if field.Description != nil {
			desc := i18n.GetDescription(field.Description.Zh, field.Description.En)
			if desc != "" {
				buf.WriteString(fmt.Sprintf("%s\t/// <summary>\n", indentStr))
				buf.WriteString(fmt.Sprintf("%s\t/// %s\n", indentStr, desc))
				buf.WriteString(fmt.Sprintf("%s\t/// </summary>\n", indentStr))
			}
		}

		csharpType := g.mapType(field.Type)
		csharpName := codegen.ToPascalCase(fieldName)
		buf.WriteString(fmt.Sprintf("%s\tpublic %s %s { get; set; } = %s;\n", indentStr, csharpType, csharpName, g.defaultValueExpr(s, field.Type)))
	}

	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("%s\tpublic static %s Rent()\n", indentStr, name))
	buf.WriteString(fmt.Sprintf("%s\t{\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\tif (Pool.TryTake(out var value))\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\t{\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\t\treturn value;\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\t}\n\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\treturn new %s();\n", indentStr, name))
	buf.WriteString(fmt.Sprintf("%s\t}\n\n", indentStr))

	buf.WriteString(fmt.Sprintf("%s\tpublic static void Return(%s? value)\n", indentStr, name))
	buf.WriteString(fmt.Sprintf("%s\t{\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\tif (value is null)\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\t{\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\t\treturn;\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\t}\n\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\tvalue.Reset();\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\tPool.Add(value);\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t}\n\n", indentStr))

	buf.WriteString(fmt.Sprintf("%s\tpublic void Release()\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t{\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t\tReturn(this);\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t}\n\n", indentStr))

	buf.WriteString(fmt.Sprintf("%s\tpublic void Reset()\n", indentStr))
	buf.WriteString(fmt.Sprintf("%s\t{\n", indentStr))
	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		csharpName := codegen.ToPascalCase(fieldName)
		buf.WriteString(fmt.Sprintf("%s\t\t%s = %s;\n", indentStr, csharpName, g.defaultValueExpr(s, field.Type)))
	}
	buf.WriteString(fmt.Sprintf("%s\t}\n", indentStr))
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

func (g *Generator) defaultValueExpr(s *schema.Schema, schemaType string) string {
	switch schemaType {
	case "bool":
		return "false"
	case "int32", "int64", "uint32", "uint64":
		return "0"
	case "float32":
		return "0f"
	case "float64":
		return "0d"
	case "string":
		return "string.Empty"
	case "bytes":
		return "Array.Empty<byte>()"
	default:
		if strings.HasPrefix(schemaType, "list<") {
			return "new()"
		}
		if strings.HasPrefix(schemaType, "map<") {
			return "new()"
		}
		if enum, ok := s.Enums[schemaType]; ok {
			if value, exists := codegen.DefaultEnumValue(enum); exists {
				return fmt.Sprintf("%s.%s", schemaType, codegen.ToPascalCase(value.Name))
			}
		}
		return fmt.Sprintf("default(%s)!", g.mapType(schemaType))
	}
}

func init() {
	codegen.Register(New())
}
