package gocodegen

import (
	"strings"
	"testing"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

func TestGoGenerator(t *testing.T) {
	gen := New()

	if gen.Name() != "go" {
		t.Errorf("Expected name 'go', got '%s'", gen.Name())
	}

	if gen.FileExtension() != ".go" {
		t.Errorf("Expected extension '.go', got '%s'", gen.FileExtension())
	}

	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "user",
		Messages: map[string]*schema.Message{
			"UserProfile": {
				Fields: map[string]*schema.Field{
					"id":   {Type: "uint32", Tag: 1},
					"name": {Type: "string", Tag: 2},
				},
			},
		},
		Enums: map[string]*schema.Enum{
			"UserType": {
				Values: map[string]int{
					"ADMIN": 0,
					"USER":  1,
				},
			},
		},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}

	content := string(files[0].Content)

	if !strings.Contains(content, "package user") {
		t.Error("Expected package declaration")
	}
	if !strings.Contains(content, "type UserProfile struct") {
		t.Error("Expected UserProfile struct")
	}
	if !strings.Contains(content, "Id uint32") {
		t.Error("Expected Id field")
	}
	if !strings.Contains(content, "Name string") {
		t.Error("Expected Name field")
	}
	if !strings.Contains(content, "type UserType int32") {
		t.Error("Expected UserType enum")
	}
	if !strings.Contains(content, "UserTypeADMIN UserType = 0") {
		t.Error("Expected ADMIN constant")
	}
}

func TestGoGeneratorNestedTypes(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "test",
		Messages: map[string]*schema.Message{
			"Test": {
				Fields: map[string]*schema.Field{
					"tags":     {Type: "list<string>", Tag: 1},
					"metadata": {Type: "map<string, string>", Tag: 2},
					"nested":   {Type: "map<string, list<uint32>>", Tag: 3},
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
	if !strings.Contains(content, "Tags []string") {
		t.Error("Expected Tags []string")
	}
	if !strings.Contains(content, "Metadata map[string]string") {
		t.Error("Expected Metadata map[string]string")
	}
}
