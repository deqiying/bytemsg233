package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseProto(t *testing.T) {
	data := []byte(`syntax = "proto3";

package protocol;

// ByteMsg233 schema: bymsg/v1
// ByteMsg233 protocolVersion: 7

enum PlayerState {
  PLAYER_STATE_UNKNOWN = 0;
  PLAYER_STATE_ACTIVE = 1;
}

message PlayerProfile {
  uint32 level = 1;
}

// ByteMsg233 packetId: 1001
message Player {
  uint64 id = 1;
  sint32 score = 2;
  string name = 3;
  repeated string tags = 4;
  map<string, uint32> attrs = 5;
  PlayerState state = 6;
  PlayerProfile profile = 7;
  float heat = 8;
  double power = 9;
  bytes payload = 10;
}
`)

	s, err := ParseProto(data)
	if err != nil {
		t.Fatalf("ParseProto failed: %v", err)
	}
	if s.Version != "bymsg/v1" || s.ProtocolVersion != 7 || s.Package != "protocol" {
		t.Fatalf("metadata mismatch: %#v", s)
	}
	if s.Enums["PlayerState"].Values["PLAYER_STATE_ACTIVE"] != 1 {
		t.Fatalf("enum value mismatch")
	}
	player := s.Messages["Player"]
	if player.PacketID != 1001 {
		t.Fatalf("packet id = %d, want 1001", player.PacketID)
	}
	expectedTypes := map[string]string{
		"id":      "uint64",
		"score":   "int32",
		"name":    "string",
		"tags":    "list<string>",
		"attrs":   "map<string, uint32>",
		"state":   "PlayerState",
		"profile": "PlayerProfile",
		"heat":    "float32",
		"power":   "float64",
		"payload": "bytes",
	}
	for fieldName, expectedType := range expectedTypes {
		if player.Fields[fieldName].Type != expectedType {
			t.Fatalf("%s type = %s, want %s", fieldName, player.Fields[fieldName].Type, expectedType)
		}
	}
}

func TestImportFileProto(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "protocol.proto")
	if err := os.WriteFile(path, []byte(`syntax = "proto3";
package protocol;
message Ping {
  uint64 id = 1;
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := ImportFile(path, nil)
	if err != nil {
		t.Fatalf("ImportFile proto failed: %v", err)
	}
	if s.Version != "bymsg/v1" {
		t.Fatalf("expected default schema version, got %s", s.Version)
	}
	if s.Messages["Ping"].Fields["id"].Type != "uint64" {
		t.Fatalf("expected uint64 field")
	}
}

func TestParseProtoPreservesDescriptions(t *testing.T) {
	data := []byte(`syntax = "proto3";

package protocol;

// zh: 用户类型
// en: User Type
enum UserType {
  USER_TYPE_UNKNOWN = 0;
  USER_TYPE_NORMAL = 1;
}

// zh: 用户资料
// en: User Profile
// ByteMsg233 packetId: 1001
message UserProfile {
  // zh: 用户ID
  // en: User ID
  uint64 id = 1;
  string name = 2; // zh: 用户名
  // 普通中文注释
  string nickname = 3;
}
`)

	s, err := ParseProto(data)
	if err != nil {
		t.Fatalf("ParseProto failed: %v", err)
	}
	enumDesc := s.Enums["UserType"].Description
	if enumDesc == nil || enumDesc.Zh != "用户类型" || enumDesc.En != "User Type" {
		t.Fatalf("enum description mismatch: %#v", enumDesc)
	}
	msg := s.Messages["UserProfile"]
	if msg.PacketID != 1001 {
		t.Fatalf("packet id = %d, want 1001", msg.PacketID)
	}
	if msg.Description == nil || msg.Description.Zh != "用户资料" || msg.Description.En != "User Profile" {
		t.Fatalf("message description mismatch: %#v", msg.Description)
	}
	idDesc := msg.Fields["id"].Description
	if idDesc == nil || idDesc.Zh != "用户ID" || idDesc.En != "User ID" {
		t.Fatalf("id description mismatch: %#v", idDesc)
	}
	nameDesc := msg.Fields["name"].Description
	if nameDesc == nil || nameDesc.Zh != "用户名" {
		t.Fatalf("name description mismatch: %#v", nameDesc)
	}
	nicknameDesc := msg.Fields["nickname"].Description
	if nicknameDesc == nil || nicknameDesc.Zh != "普通中文注释" {
		t.Fatalf("nickname description mismatch: %#v", nicknameDesc)
	}
}

func TestParseProtoRejectsUnsupportedSyntax(t *testing.T) {
	_, err := ParseProto([]byte(`syntax = "proto3";
package protocol;
message Bad {
  oneof value {
    string name = 1;
  }
}
`))
	if err == nil || !strings.Contains(err.Error(), "oneof is not supported") {
		t.Fatalf("expected unsupported oneof error, got %v", err)
	}
}

func TestParseProtoRejectsUnsupportedMapKey(t *testing.T) {
	_, err := ParseProto([]byte(`syntax = "proto3";
package protocol;
message Bad {
  map<float, string> weights = 1;
}
`))
	if err == nil || !strings.Contains(err.Error(), "map key type") {
		t.Fatalf("expected unsupported map key error, got %v", err)
	}
}

func TestParseProtoRejectsUnsupportedScalar(t *testing.T) {
	_, err := ParseProto([]byte(`syntax = "proto3";
package protocol;
message Bad {
  fixed32 id = 1;
}
`))
	if err == nil || !strings.Contains(err.Error(), `unknown type "fixed32"`) {
		t.Fatalf("expected unsupported scalar error, got %v", err)
	}
}

func TestParseProtoRejectsRepeatedMap(t *testing.T) {
	_, err := ParseProto([]byte(`syntax = "proto3";
package protocol;
message Bad {
  repeated map<string, uint32> attrs = 1;
}
`))
	if err == nil || !strings.Contains(err.Error(), "repeated map fields are not supported") {
		t.Fatalf("expected repeated map error, got %v", err)
	}
}

func TestParseProtoPacketCommentDoesNotLeakFromField(t *testing.T) {
	s, err := ParseProto([]byte(`syntax = "proto3";
package protocol;
message First {
  uint64 id = 1;
  // ByteMsg233 packetId: 999
}
message Second {
  uint64 id = 1;
}
`))
	if err != nil {
		t.Fatalf("ParseProto failed: %v", err)
	}
	if s.Messages["Second"].PacketID != 0 {
		t.Fatalf("field comment leaked into next message packet id: %d", s.Messages["Second"].PacketID)
	}
}
