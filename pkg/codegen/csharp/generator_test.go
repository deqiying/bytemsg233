package csharpgen

import (
	"strings"
	"testing"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

func TestCSharpGenerator(t *testing.T) {
	gen := New()

	if gen.Name() != "csharp" {
		t.Errorf("Expected name 'csharp', got '%s'", gen.Name())
	}

	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "Example.User",
		Messages: map[string]*schema.Message{
			"UserProfile": {
				Description: &schema.Description{En: "User profile"},
				Fields: map[string]*schema.Field{
					"id":   {Type: "uint32", Tag: 1, Description: &schema.Description{En: "User ID"}},
					"name": {Type: "string", Tag: 2, Description: &schema.Description{En: "Display name"}},
				},
			},
		},
		Enums: map[string]*schema.Enum{
			"UserType": {
				Values: map[string]int{
					"Admin": 0,
					"User":  1,
				},
			},
		},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := string(files[0].Content)

	if !strings.Contains(content, "namespace Example.User") {
		t.Error("Expected namespace declaration")
	}
	if !strings.Contains(content, "public sealed class UserProfile") {
		t.Error("Expected UserProfile class")
	}
	if !strings.Contains(content, "/// User profile") {
		t.Error("Expected class comment")
	}
	if !strings.Contains(content, "/// User ID") {
		t.Error("Expected field comment")
	}
	if !strings.Contains(content, "public uint Id") {
		t.Error("Expected Id property")
	}
	if !strings.Contains(content, "public string Name") {
		t.Error("Expected Name property")
	}
	if !strings.Contains(content, "public enum UserType") {
		t.Error("Expected UserType enum")
	}
	if !strings.Contains(content, "public static class UserTypeExtensions") {
		t.Error("Expected enum extensions helper")
	}
	if !strings.Contains(content, "public static UserType FromValue(int raw)") {
		t.Error("Expected enum FromValue helper")
	}
	if !strings.Contains(content, "public static UserProfile Rent()") {
		t.Error("Expected pool rent helper")
	}
	if !strings.Contains(content, "public static void Return(UserProfile? value)") {
		t.Error("Expected pool return helper")
	}
	if !strings.Contains(content, "public void Reset()") {
		t.Error("Expected reset helper")
	}
}

func TestCSharpNestedTypes(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "Test",
		Messages: map[string]*schema.Message{
			"Test": {
				Fields: map[string]*schema.Field{
					"tags":     {Type: "list<string>", Tag: 1},
					"metadata": {Type: "map<string, string>", Tag: 2},
				},
			},
		},
		Enums: map[string]*schema.Enum{},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := string(files[0].Content)
	if !strings.Contains(content, "List<string>") {
		t.Error("Expected List<string>")
	}
	if !strings.Contains(content, "Dictionary<string, string>") {
		t.Error("Expected Dictionary<string, string>")
	}
}
