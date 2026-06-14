package schema

import (
	"testing"
)

func TestBmsgParser(t *testing.T) {
	input := `schema: bymsg/v1
package: com.example.user

enum UserType {
    ADMIN = 0
    USER = 1
    GUEST = 2
}

message UserProfile {
    uint32 id = 1 // "用户ID" | "User ID"
    string name = 2 // "用户名" | "Username"
    list<string> tags = 3
    map<string, string> metadata = 4
}
`

	schema, err := ParseBmsg([]byte(input))
	if err != nil {
		t.Fatalf("ParseBmsg failed: %v", err)
	}

	if schema.Version != "bymsg/v1" {
		t.Errorf("Expected version 'bymsg/v1', got '%s'", schema.Version)
	}

	if schema.Package != "com.example.user" {
		t.Errorf("Expected package 'com.example.user', got '%s'", schema.Package)
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

	msg, ok := schema.Messages["UserProfile"]
	if !ok {
		t.Fatal("Expected message 'UserProfile'")
	}

	if len(msg.Fields) != 4 {
		t.Errorf("Expected 4 fields, got %d", len(msg.Fields))
	}

	idField, ok := msg.Fields["id"]
	if !ok {
		t.Fatal("Expected field 'id'")
	}
	if idField.Type != "uint32" {
		t.Errorf("Expected type 'uint32', got '%s'", idField.Type)
	}
	if idField.Tag != 1 {
		t.Errorf("Expected tag 1, got %d", idField.Tag)
	}
	if idField.Description == nil {
		t.Error("Expected description")
	} else {
		if idField.Description.Zh != "用户ID" {
			t.Errorf("Expected zh '用户ID', got '%s'", idField.Description.Zh)
		}
		if idField.Description.En != "User ID" {
			t.Errorf("Expected en 'User ID', got '%s'", idField.Description.En)
		}
	}

	tagsField := msg.Fields["tags"]
	if tagsField.Type != "list<string>" {
		t.Errorf("Expected type 'list<string>', got '%s'", tagsField.Type)
	}

	metadataField := msg.Fields["metadata"]
	if metadataField.Type != "map<string, string>" {
		t.Errorf("Expected type 'map<string, string>', got '%s'", metadataField.Type)
	}
}

func TestBmsgFileParsing(t *testing.T) {
	data := []byte(`schema: bymsg/v1
package: com.example.user

enum UserType {
    ADMIN = 0
    USER = 1
    GUEST = 2
}

enum Status {
    ACTIVE = 0
    INACTIVE = 1
    BANNED = 2
}

message UserProfile {
    uint32 id = 1 // "用户ID" | "User ID"
    string name = 2 // "用户名" | "Username"
    string email = 3 // "邮箱" | "Email"
    list<string> tags = 4 // "标签" | "Tags"
    map<string, string> metadata = 5 // "元数据" | "Metadata"
    Address address = 6 // "地址" | "Address"
}

message Address {
    string street = 1 // "街道" | "Street"
    string city = 2 // "城市" | "City"
    string zip = 3 // "邮编" | "Zip Code"
}
`)

	schema, err := ParseBmsg(data)
	if err != nil {
		t.Fatalf("ParseBmsg failed: %v", err)
	}

	if len(schema.Enums) != 2 {
		t.Errorf("Expected 2 enums, got %d", len(schema.Enums))
	}

	if len(schema.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(schema.Messages))
	}

	userMsg := schema.Messages["UserProfile"]
	if len(userMsg.Fields) != 6 {
		t.Errorf("Expected 6 fields in UserProfile, got %d", len(userMsg.Fields))
	}
}
