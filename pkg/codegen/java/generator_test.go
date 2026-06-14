package javagen

import (
	"strings"
	"testing"

	"github.com/neko233-com/bytemsg233/pkg/codegen"
	"github.com/neko233-com/bytemsg233/pkg/schema"
)

func TestJavaGenerator(t *testing.T) {
	gen := New()
	s := &schema.Schema{
		Version: "bymsg/v1",
		Package: "com.example",
		Messages: map[string]*schema.Message{
			"User": {
				Fields: map[string]*schema.Field{
					"name": {Type: "string", Tag: 1},
					"age":  {Type: "uint32", Tag: 2},
				},
			},
		},
		Enums: map[string]*schema.Enum{
			"Status": {
				Values: map[string]int{"ACTIVE": 0, "INACTIVE": 1},
			},
		},
	}

	files, err := gen.Generate(s, &codegen.GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := string(files[0].Content)
	if !strings.Contains(content, "package com.example;") {
		t.Error("Expected package")
	}
	if !strings.Contains(content, "public class User") {
		t.Error("Expected class")
	}
	if !strings.Contains(content, "private String name;") {
		t.Error("Expected String field")
	}
	if !strings.Contains(content, "public enum Status") {
		t.Error("Expected enum")
	}
	if !strings.Contains(content, "import java.util.List;") {
		t.Error("Expected List import")
	}
	if !strings.Contains(content, "import java.util.Map;") {
		t.Error("Expected Map import")
	}
}

func TestJavaNestedTypes(t *testing.T) {
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
	if !strings.Contains(content, "List<String>") {
		t.Error("Expected List<String>")
	}
	if !strings.Contains(content, "Map<String, String>") {
		t.Error("Expected Map<String, String>")
	}
}
