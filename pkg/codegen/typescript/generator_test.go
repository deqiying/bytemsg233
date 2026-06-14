package tsgen

import (
	"strings"
	"testing"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

func TestTypeScriptGenerator(t *testing.T) {
	gen := New()
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
				Values: map[string]int{"ADMIN": 0, "USER": 1},
			},
		},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := string(files[0].Content)
	if !strings.Contains(content, "export interface UserProfile") {
		t.Error("Expected UserProfile interface")
	}
	if !strings.Contains(content, "id: number") {
		t.Error("Expected id: number")
	}
	if !strings.Contains(content, "name: string") {
		t.Error("Expected name: string")
	}
	if !strings.Contains(content, "export enum UserType") {
		t.Error("Expected UserType enum")
	}
}

func TestTypeScriptNestedTypes(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "test",
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
	if !strings.Contains(content, "string[]") {
		t.Error("Expected string[]")
	}
	if !strings.Contains(content, "Record<string, string>") {
		t.Error("Expected Record<string, string>")
	}
}
