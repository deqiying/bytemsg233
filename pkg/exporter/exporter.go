package exporter

import (
	"fmt"
	"strings"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

func Markdown(s *schema.Schema) []byte {
	var buf strings.Builder

	buf.WriteString("# ByteMsg Protocol\n\n")
	buf.WriteString(fmt.Sprintf("- Schema: `%s`\n", s.Version))
	buf.WriteString(fmt.Sprintf("- Package: `%s`\n\n", s.Package))

	if len(s.Enums) > 0 {
		buf.WriteString("## Enums\n\n")
		for _, enumName := range codegen.SortedEnumNames(s) {
			enum := s.Enums[enumName]
			buf.WriteString(fmt.Sprintf("### %s\n\n", enumName))
			if enum.Description != nil {
				buf.WriteString(descriptionLine(enum.Description))
				buf.WriteString("\n")
			}
			buf.WriteString("| Name | Value |\n|---|---:|\n")
			for _, value := range codegen.SortedEnumValues(enum) {
				buf.WriteString(fmt.Sprintf("| `%s` | %d |\n", value.Name, value.Value))
			}
			buf.WriteString("\n")
		}
	}

	if len(s.Messages) > 0 {
		buf.WriteString("## Messages\n\n")
		for _, msgName := range codegen.SortedMessageNames(s) {
			msg := s.Messages[msgName]
			buf.WriteString(fmt.Sprintf("### %s\n\n", msgName))
			if msg.Description != nil {
				buf.WriteString(descriptionLine(msg.Description))
				buf.WriteString("\n")
			}
			buf.WriteString("| Tag | Field | Type | Description |\n|---:|---|---|---|\n")
			for _, fieldName := range codegen.SortedFieldNames(msg) {
				field := msg.Fields[fieldName]
				desc := ""
				if field.Description != nil {
					desc = inlineDescription(field.Description)
				}
				buf.WriteString(fmt.Sprintf("| %d | `%s` | `%s` | %s |\n", field.Tag, fieldName, field.Type, desc))
			}
			buf.WriteString("\n")
		}
	}

	buf.WriteString("## CLI\n\n")
	buf.WriteString("```bash\n")
	buf.WriteString("bytemsg233 compile protocol.bmsg.json -l go,csharp,typescript,rust,java -o ./gen\n")
	buf.WriteString("bytemsg233 install-lib csharp --to ./Assets/Plugins/ByteMsg233\n")
	buf.WriteString("```\n")

	return []byte(buf.String())
}

func Bmsg(s *schema.Schema) []byte {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("schema: %s\n", s.Version))
	buf.WriteString(fmt.Sprintf("package: %s\n\n", s.Package))

	for _, enumName := range codegen.SortedEnumNames(s) {
		enum := s.Enums[enumName]
		buf.WriteString(fmt.Sprintf("enum %s {\n", enumName))
		for _, value := range codegen.SortedEnumValues(enum) {
			buf.WriteString(fmt.Sprintf("    %s = %d\n", value.Name, value.Value))
		}
		buf.WriteString("}\n\n")
	}

	for _, msgName := range codegen.SortedMessageNames(s) {
		msg := s.Messages[msgName]
		buf.WriteString(fmt.Sprintf("message %s {\n", msgName))
		for _, fieldName := range codegen.SortedFieldNames(msg) {
			field := msg.Fields[fieldName]
			buf.WriteString(fmt.Sprintf("    %s %s = %d", field.Type, fieldName, field.Tag))
			if field.Description != nil {
				buf.WriteString(fmt.Sprintf(" // %q | %q", field.Description.Zh, field.Description.En))
			}
			buf.WriteString("\n")
		}
		buf.WriteString("}\n\n")
	}

	return []byte(buf.String())
}

func descriptionLine(desc *schema.Description) string {
	return fmt.Sprintf("> zh: %s\n>\n> en: %s\n", desc.Zh, desc.En)
}

func inlineDescription(desc *schema.Description) string {
	if desc.Zh != "" && desc.En != "" {
		return fmt.Sprintf("%s / %s", desc.Zh, desc.En)
	}
	if desc.Zh != "" {
		return desc.Zh
	}
	return desc.En
}
