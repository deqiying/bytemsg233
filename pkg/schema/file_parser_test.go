package schema

import (
	"os"
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

func TestParseFileBmsgYAMLCompatibility(t *testing.T) {
	t.Run(".bmsg can contain the default JSON DSL", func(t *testing.T) {
		dir := t.TempDir()
		path := dir + "/user.bmsg"
		data := []byte(`{
  "schema": "bymsg/v1",
  "package": "com.example.user",
  "User": {
    "id": { "type": "uint32", "tag": 1 },
    "name": { "type": "string", "tag": 2 }
  }
}`)
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}

		s, err := ParseFile(path)
		if err != nil {
			t.Fatalf("ParseFile JSON .bmsg: %v", err)
		}
		if _, ok := s.Messages["User"]; !ok {
			t.Fatal("Expected User message")
		}
	})
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

func TestParseFileNativeJSON(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/game.bmsg.json"
	data := []byte(`{
  "schema": "bymsg/v1",
  "package": "com.example.game",
  "enums": {
    "HeroState": {
      "values": {
        "IDLE": 0,
        "MOVING": 1
      }
    }
  },
  "Hero": {
    "description": {
      "zh": "英雄",
      "en": "Hero"
    },
    "id": {
      "type": "uint32",
      "tag": 1
    },
    "state": {
      "type": "HeroState",
      "tag": 2
    }
  }
}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	s, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile native json: %v", err)
	}
	if len(s.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(s.Messages))
	}
	if s.Messages["Hero"].Fields["state"].Type != "HeroState" {
		t.Fatalf("Expected Hero.state to use HeroState")
	}
	if len(s.Enums) != 1 {
		t.Fatalf("Expected 1 enum, got %d", len(s.Enums))
	}
}
