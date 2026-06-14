package schema

import (
	"testing"
)

func TestParseSchema(t *testing.T) {
	yamlContent := `
schema: bymsg/v1
package: com.example.user

messages:
  UserProfile:
    fields:
      id:
        type: uint32
        tag: 1
        description:
          zh: "用户ID"
          en: "User ID"
      name:
        type: string
        tag: 2
    description:
      zh: "用户资料"
      en: "User Profile"

enums:
  UserType:
    values:
      ADMIN: 0
      USER: 1
      GUEST: 2
    description:
      zh: "用户类型"
      en: "User Type"
`

	schema, err := Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if schema.Version != "bymsg/v1" {
		t.Errorf("Expected version 'bymsg/v1', got '%s'", schema.Version)
	}

	if schema.Package != "com.example.user" {
		t.Errorf("Expected package 'com.example.user', got '%s'", schema.Package)
	}

	msg, ok := schema.Messages["UserProfile"]
	if !ok {
		t.Fatal("Expected message 'UserProfile'")
	}

	if len(msg.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(msg.Fields))
	}

	if msg.Fields["id"].Type != "uint32" {
		t.Errorf("Expected field 'id' type 'uint32', got '%s'", msg.Fields["id"].Type)
	}

	if msg.Fields["id"].Tag != 1 {
		t.Errorf("Expected field 'id' tag 1, got %d", msg.Fields["id"].Tag)
	}

	if msg.Fields["id"].Description == nil {
		t.Error("Expected field 'id' to have description")
	} else {
		if msg.Fields["id"].Description.Zh != "用户ID" {
			t.Errorf("Expected zh description '用户ID', got '%s'", msg.Fields["id"].Description.Zh)
		}
		if msg.Fields["id"].Description.En != "User ID" {
			t.Errorf("Expected en description 'User ID', got '%s'", msg.Fields["id"].Description.En)
		}
	}

	enum, ok := schema.Enums["UserType"]
	if !ok {
		t.Fatal("Expected enum 'UserType'")
	}

	if len(enum.Values) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(enum.Values))
	}

	if enum.Values["ADMIN"] != 0 {
		t.Errorf("Expected ADMIN=0, got %d", enum.Values["ADMIN"])
	}
}

func TestParseNestedTypes(t *testing.T) {
	yamlContent := `
schema: bymsg/v1
package: com.example.test

messages:
  TestMessage:
    fields:
      tags:
        type: list<string>
        tag: 1
      metadata:
        type: map<string, string>
        tag: 2
      nested:
        type: "map<string, list<uint32>>"
        tag: 3
`

	schema, err := Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	msg := schema.Messages["TestMessage"]
	if msg.Fields["tags"].Type != "list<string>" {
		t.Errorf("Expected 'list<string>', got '%s'", msg.Fields["tags"].Type)
	}

	if msg.Fields["metadata"].Type != "map<string, string>" {
		t.Errorf("Expected 'map<string, string>', got '%s'", msg.Fields["metadata"].Type)
	}
}
