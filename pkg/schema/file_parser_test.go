package schema

import (
	"testing"
)

func TestParseFileBmsg(t *testing.T) {
	s, err := ParseFile("../../testdata/user.bmsg")
	if err != nil {
		t.Fatalf("ParseFile .bmsg: %v", err)
	}
	if s.Version != "bymsg/v1" {
		t.Errorf("Expected version bymsg/v1, got %s", s.Version)
	}
	if len(s.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(s.Messages))
	}
	if len(s.Enums) != 2 {
		t.Errorf("Expected 2 enums, got %d", len(s.Enums))
	}
}

func TestParseFileYAML(t *testing.T) {
	s, err := ParseFile("../../testdata/user.bmsg.yaml")
	if err != nil {
		t.Fatalf("ParseFile .yaml: %v", err)
	}
	if s.Version != "bymsg/v1" {
		t.Errorf("Expected version bymsg/v1, got %s", s.Version)
	}
	if len(s.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(s.Messages))
	}
}

func TestParseFileJSON(t *testing.T) {
	s, err := ParseFile("../../testdata/user.json")
	if err != nil {
		t.Fatalf("ParseFile .json: %v", err)
	}
	if s.Version != "bymsg/v1" {
		t.Errorf("Expected version bymsg/v1, got %s", s.Version)
	}
	if len(s.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(s.Messages))
	}
	if len(s.Enums) != 1 {
		t.Errorf("Expected 1 enum, got %d", len(s.Enums))
	}
	msg := s.Messages["UserProfile"]
	if msg.Fields["id"].Type != "uint32" {
		t.Errorf("Expected uint32, got %s", msg.Fields["id"].Type)
	}
}
