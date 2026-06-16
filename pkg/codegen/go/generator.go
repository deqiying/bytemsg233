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

	prevLocale := i18n.GetLocale()
	if options != nil && options.Locale != "" {
		i18n.SetLocale(options.Locale)
		defer i18n.SetLocale(prevLocale)
	}

	packageName := s.Package
	if options != nil && options.Package != "" {
		packageName = options.Package
	}

	buf.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	if len(s.Messages) > 0 {
		buf.WriteString("import \"sync\"\n\n")
	}

	for _, name := range codegen.SortedEnumNames(s) {
		g.generateEnum(&buf, name, s.Enums[name])
		buf.WriteString("\n")
	}

	for _, name := range codegen.SortedMessageNames(s) {
		g.generateMessage(&buf, s, name, s.Messages[name])
		buf.WriteString("\n")
	}

	return []*codegen.GeneratedFile{
		{Path: "types" + g.FileExtension(), Content: []byte(buf.String())},
	}, nil
}

func (g *Generator) generateEnum(buf *strings.Builder, name string, enum *schema.Enum) {
	if enum.Description != nil {
		desc := i18n.GetDescription(enum.Description.Zh, enum.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("// %s\n", desc))
		}
	}

	buf.WriteString(fmt.Sprintf("type %s int32\n\n", name))
	buf.WriteString("const (\n")
	values := codegen.SortedEnumValues(enum)
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("\t%s%s %s = %d\n", name, codegen.ToPascalCase(value.Name), name, value.Value))
	}
	buf.WriteString(")\n\n")

	buf.WriteString(fmt.Sprintf("func (x %s) String() string {\n", name))
	buf.WriteString("\tswitch x {\n")
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("\tcase %s%s:\n", name, codegen.ToPascalCase(value.Name)))
		buf.WriteString(fmt.Sprintf("\t\treturn %q\n", value.Name))
	}
	buf.WriteString("\tdefault:\n")
	buf.WriteString(fmt.Sprintf("\t\treturn fmt.Sprintf(\"%s(%%d)\", int32(x))\n", name))
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("func Parse%s(value int32) (%s, bool) {\n", name, name))
	buf.WriteString("\tswitch value {\n")
	for _, value := range values {
		buf.WriteString(fmt.Sprintf("\tcase %d:\n", value.Value))
		buf.WriteString(fmt.Sprintf("\t\treturn %s%s, true\n", name, codegen.ToPascalCase(value.Name)))
	}
	buf.WriteString("\tdefault:\n")
	buf.WriteString(fmt.Sprintf("\t\treturn %s(0), false\n", name))
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("func (x %s) IsValid() bool {\n", name))
	buf.WriteString("\t_, ok := Parse" + name + "(int32(x))\n")
	buf.WriteString("\treturn ok\n")
	buf.WriteString("}\n")
}

func (g *Generator) generateMessage(buf *strings.Builder, s *schema.Schema, name string, msg *schema.Message) {
	if msg.Description != nil {
		desc := i18n.GetDescription(msg.Description.Zh, msg.Description.En)
		if desc != "" {
			buf.WriteString(fmt.Sprintf("// %s\n", desc))
		}
	}

	buf.WriteString(fmt.Sprintf("type %s struct {\n", name))
	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		if field.Description != nil {
			desc := i18n.GetDescription(field.Description.Zh, field.Description.En)
			if desc != "" {
				buf.WriteString(fmt.Sprintf("\t// %s\n", desc))
			}
		}

		goType := g.mapType(field.Type)
		goName := codegen.ToPascalCase(fieldName)
		buf.WriteString(fmt.Sprintf("\t%s %s `bytemsg:\"%d\"`\n", goName, goType, field.Tag))
	}
	buf.WriteString("}\n\n")

	poolName := codegen.ToCamelCase(name) + "Pool"
	buf.WriteString(fmt.Sprintf("var %s = sync.Pool{\n", poolName))
	buf.WriteString("\tNew: func() any {\n")
	buf.WriteString(fmt.Sprintf("\t\treturn &%s{}\n", name))
	buf.WriteString("\t},\n")
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("// Acquire%s gets a reusable %s from the pool.\n", name, name))
	buf.WriteString(fmt.Sprintf("func Acquire%s() *%s {\n", name, name))
	buf.WriteString(fmt.Sprintf("\treturn %s.Get().(*%s)\n", poolName, name))
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("// Release%s resets a %s and returns it to the pool.\n", name, name))
	buf.WriteString(fmt.Sprintf("func Release%s(value *%s) {\n", name, name))
	buf.WriteString("\tif value == nil {\n")
	buf.WriteString("\t\treturn\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tvalue.Reset()\n")
	buf.WriteString(fmt.Sprintf("\t%s.Put(value)\n", poolName))
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("// Reset clears %s before it is reused.\n", name))
	buf.WriteString(fmt.Sprintf("func (x *%s) Reset() {\n", name))
	buf.WriteString("\tif x == nil {\n")
	buf.WriteString("\t\treturn\n")
	buf.WriteString("\t}\n")
	for _, fieldName := range codegen.SortedFieldNames(msg) {
		field := msg.Fields[fieldName]
		goName := codegen.ToPascalCase(fieldName)
		buf.WriteString(fmt.Sprintf("\tx.%s = %s\n", goName, g.zeroValueExpr(s, field.Type)))
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

func (g *Generator) zeroValueExpr(s *schema.Schema, schemaType string) string {
	switch schemaType {
	case "bool":
		return "false"
	case "int32", "int64", "uint32", "uint64", "float32", "float64":
		return "0"
	case "string":
		return "\"\""
	case "bytes":
		return "nil"
	default:
		if strings.HasPrefix(schemaType, "list<") || strings.HasPrefix(schemaType, "map<") {
			return "nil"
		}
		if enum, ok := s.Enums[schemaType]; ok {
			if value, exists := codegen.DefaultEnumValue(enum); exists {
				return fmt.Sprintf("%s%s", schemaType, codegen.ToPascalCase(value.Name))
			}
		}
		return fmt.Sprintf("*new(%s)", schemaType)
	}
}

func init() {
	codegen.Register(New())
}
