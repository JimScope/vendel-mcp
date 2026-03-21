package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestListTemplatesTool_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_templates/records", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PaginatedResponse[SmsTemplate]{
			Page: 1, PerPage: 20, TotalPages: 1, TotalItems: 2,
			Items: []SmsTemplate{
				{ID: "tpl_1", Name: "Greeting", Body: "Hello {{name}}", Created: "2026-03-21T10:00:00Z"},
				{ID: "tpl_2", Name: "Alert", Body: "Alert: {{msg}}", Created: "2026-03-21T11:00:00Z"},
			},
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_templates", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	for _, want := range []string{"Templates (2):", "Greeting [ID: tpl_1]", "Hello {{name}}"} {
		if !strings.Contains(text, want) {
			t.Errorf("text should contain %q, got: %q", want, text)
		}
	}
}

func TestListTemplatesTool_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_templates/records", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PaginatedResponse[SmsTemplate]{
			Page: 1, PerPage: 20, TotalPages: 0, TotalItems: 0,
			Items: []SmsTemplate{},
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_templates", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	if text != "No templates found." {
		t.Errorf("text = %q, want %q", text, "No templates found.")
	}
}

func TestListTemplatesTool_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_templates/records", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "server error"})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_templates", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
}

func TestSendTemplateTool_Success_WithBatchID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_templates/records/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SmsTemplate{
			ID: "tpl_1", Name: "Greeting", Body: "Hello World",
		})
	})
	mux.HandleFunc("/api/sms/send", func(w http.ResponseWriter, r *http.Request) {
		var req SendSmsRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Body != "Hello World" {
			t.Errorf("Body = %q, want template body %q", req.Body, "Hello World")
		}
		json.NewEncoder(w).Encode(SendSmsResponse{
			BatchID: "batch_1", MessageIDs: []string{"msg_1"},
			RecipientsCount: 1, Status: "queued",
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "send_template", map[string]any{
		"template_id": "tpl_1",
		"recipients":  []string{"+1111111111"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	for _, want := range []string{`Template "Greeting" sent`, "Batch ID: batch_1"} {
		if !strings.Contains(text, want) {
			t.Errorf("text should contain %q, got: %q", want, text)
		}
	}
}

func TestSendTemplateTool_Success_NoBatchID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_templates/records/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SmsTemplate{
			ID: "tpl_1", Name: "Greeting", Body: "Hello World",
		})
	})
	mux.HandleFunc("/api/sms/send", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SendSmsResponse{
			MessageIDs: []string{"msg_1"}, RecipientsCount: 1, Status: "queued",
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "send_template", map[string]any{
		"template_id": "tpl_1",
		"recipients":  []string{"+1111111111"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	if strings.Contains(text, "Batch ID:") {
		t.Errorf("text should not contain 'Batch ID:', got: %q", text)
	}
}

func TestSendTemplateTool_InvalidTemplateID(t *testing.T) {
	mux := http.NewServeMux()
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "send_template", map[string]any{
		"template_id": "../../etc/passwd",
		"recipients":  []string{"+1111111111"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "invalid template_id") {
		t.Errorf("text should contain 'invalid template_id', got: %q", text)
	}
}

func TestSendTemplateTool_EmptyRecipients(t *testing.T) {
	mux := http.NewServeMux()
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "send_template", map[string]any{
		"template_id": "tpl_1",
		"recipients":  []string{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "recipients must not be empty") {
		t.Errorf("text should contain validation message, got: %q", text)
	}
}

func TestSendTemplateTool_FetchError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_templates/records/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "send_template", map[string]any{
		"template_id": "missing",
		"recipients":  []string{"+1111111111"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "fetch template") {
		t.Errorf("text should contain 'fetch template', got: %q", text)
	}
}

func TestSendTemplateTool_SendError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_templates/records/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SmsTemplate{
			ID: "tpl_1", Name: "Greeting", Body: "Hello World",
		})
	})
	mux.HandleFunc("/api/sms/send", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "quota exceeded"})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "send_template", map[string]any{
		"template_id": "tpl_1",
		"recipients":  []string{"+1111111111"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "send template") {
		t.Errorf("text should contain 'send template', got: %q", text)
	}
}
