package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestTextResult(t *testing.T) {
	result := textResult("hello")

	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content element, got %d", len(result.Content))
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}

	if tc.Text != "hello" {
		t.Errorf("text = %q, want %q", tc.Text, "hello")
	}
	if result.IsError {
		t.Error("IsError should be false")
	}
}

func TestErrorResult(t *testing.T) {
	err := errors.New("quota exceeded")
	result := errorResult("send SMS", err)

	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content element, got %d", len(result.Content))
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}

	if !result.IsError {
		t.Error("IsError should be true")
	}
	if !strings.Contains(tc.Text, "send SMS") {
		t.Errorf("text should contain 'send SMS', got %q", tc.Text)
	}
	if !strings.Contains(tc.Text, "quota exceeded") {
		t.Errorf("text should contain 'quota exceeded', got %q", tc.Text)
	}
	want := "Failed to send SMS: quota exceeded"
	if tc.Text != want {
		t.Errorf("text = %q, want %q", tc.Text, want)
	}
}

func TestValidateRecordID(t *testing.T) {
	valid := []string{"abc123def456ghi", "msg_42", "tpl1", "UPPER", "mix3d"}
	for _, id := range valid {
		if !validateRecordID(id) {
			t.Errorf("validateRecordID(%q) should be true", id)
		}
	}

	invalid := []string{"", "../etc/passwd", "id/../../x", "id?q=1", "id&x=1", "id#frag", "has space", "a\x00b"}
	for _, id := range invalid {
		if validateRecordID(id) {
			t.Errorf("validateRecordID(%q) should be false", id)
		}
	}
}

func TestValidateEnum(t *testing.T) {
	allowed := []string{"outgoing", "incoming"}

	if !validateEnum("outgoing", allowed) {
		t.Error("validateEnum should return true for 'outgoing'")
	}
	if !validateEnum("incoming", allowed) {
		t.Error("validateEnum should return true for 'incoming'")
	}
	if validateEnum("invalid", allowed) {
		t.Error("validateEnum should return false for 'invalid'")
	}
	if validateEnum("", allowed) {
		t.Error("validateEnum should return false for empty string")
	}
}
