package codegen

import (
	"sort"
	"strings"
	"unicode"

	"github.com/neko233-com/bytemsg233/pkg/schema"
)

type EnumValue struct {
	Name  string
	Value int
}

func SortedEnumNames(s *schema.Schema) []string {
	if s == nil || len(s.Enums) == 0 {
		return nil
	}

	names := make([]string, 0, len(s.Enums))
	for name := range s.Enums {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func SortedMessageNames(s *schema.Schema) []string {
	if s == nil || len(s.Messages) == 0 {
		return nil
	}

	names := make([]string, 0, len(s.Messages))
	for name := range s.Messages {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func SortedFieldNames(msg *schema.Message) []string {
	if msg == nil || len(msg.Fields) == 0 {
		return nil
	}

	names := make([]string, 0, len(msg.Fields))
	for name := range msg.Fields {
		names = append(names, name)
	}

	sort.Slice(names, func(i, j int) bool {
		left := msg.Fields[names[i]]
		right := msg.Fields[names[j]]
		if left.Tag == right.Tag {
			return names[i] < names[j]
		}
		return left.Tag < right.Tag
	})

	return names
}

func SortedEnumValues(enum *schema.Enum) []EnumValue {
	if enum == nil || len(enum.Values) == 0 {
		return nil
	}

	values := make([]EnumValue, 0, len(enum.Values))
	for name, value := range enum.Values {
		values = append(values, EnumValue{Name: name, Value: value})
	}

	sort.Slice(values, func(i, j int) bool {
		if values[i].Value == values[j].Value {
			return values[i].Name < values[j].Name
		}
		return values[i].Value < values[j].Value
	})

	return values
}

func DefaultEnumValue(enum *schema.Enum) (EnumValue, bool) {
	values := SortedEnumValues(enum)
	if len(values) == 0 {
		return EnumValue{}, false
	}
	return values[0], true
}

func ToPascalCase(name string) string {
	if name == "" {
		return ""
	}

	parts := splitIdentifier(name)
	for i, part := range parts {
		runes := []rune(strings.ToLower(part))
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, "")
}

func ToCamelCase(name string) string {
	pascal := ToPascalCase(name)
	if pascal == "" {
		return ""
	}

	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func splitIdentifier(name string) []string {
	var parts []string
	var current []rune

	flush := func() {
		if len(current) == 0 {
			return
		}
		parts = append(parts, string(current))
		current = nil
	}

	for _, r := range name {
		if r == '_' || r == '-' || unicode.IsSpace(r) {
			flush()
			continue
		}
		current = append(current, r)
	}
	flush()

	if len(parts) == 0 {
		return []string{name}
	}

	return parts
}
