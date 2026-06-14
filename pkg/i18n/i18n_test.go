package i18n

import (
	"testing"
)

func TestGetMessage(t *testing.T) {
	Reset()

	msg := Get("compile.success")
	if msg != "Compilation successful" {
		t.Errorf("Expected 'Compilation successful', got '%s'", msg)
	}

	SetLocale("zh")
	msg = Get("compile.success")
	if msg != "编译成功" {
		t.Errorf("Expected '编译成功', got '%s'", msg)
	}

	msg = Get("unknown.key")
	if msg != "unknown.key" {
		t.Errorf("Expected 'unknown.key', got '%s'", msg)
	}

	Reset()
}

func TestGetDescription(t *testing.T) {
	SetLocale("zh")
	msg := GetDescription("用户ID", "User ID")
	if msg != "用户ID" {
		t.Errorf("Expected '用户ID', got '%s'", msg)
	}

	SetLocale("en")
	msg = GetDescription("用户ID", "User ID")
	if msg != "User ID" {
		t.Errorf("Expected 'User ID', got '%s'", msg)
	}

	msg = GetDescription("", "")
	if msg != "" {
		t.Errorf("Expected empty string, got '%s'", msg)
	}

	Reset()
}

func TestSupportedLocales(t *testing.T) {
	locales := SupportedLocales()
	if len(locales) != 2 {
		t.Errorf("Expected 2 supported locales, got %d", len(locales))
	}
}
