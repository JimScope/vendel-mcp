package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestSendSmsTool_Success_WithBatchID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/sms/send", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SendSmsResponse{
			BatchID:         "batch_123",
			MessageIDs:      []string{"msg_1", "msg_2"},
			RecipientsCount: 2,
			Status:          "queued",
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "send_sms", map[string]any{
		"recipients": []string{"+1111111111", "+2222222222"},
		"body":       "Hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	for _, want := range []string{"SMS queued successfully", "Recipients: 2", "Batch ID: batch_123"} {
		if !strings.Contains(text, want) {
			t.Errorf("text should contain %q, got: %q", want, text)
		}
	}
	if result.IsError {
		t.Error("IsError should be false")
	}
}

func TestSendSmsTool_Success_NoBatchID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/sms/send", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SendSmsResponse{
			MessageIDs:      []string{"msg_1"},
			RecipientsCount: 1,
			Status:          "queued",
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "send_sms", map[string]any{
		"recipients": []string{"+1111111111"},
		"body":       "Hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	if strings.Contains(text, "Batch ID:") {
		t.Errorf("text should not contain 'Batch ID:', got: %q", text)
	}
}

func TestSendSmsTool_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/sms/send", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "quota exceeded"})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "send_sms", map[string]any{
		"recipients": []string{"+1111111111"},
		"body":       "Hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "send SMS") {
		t.Errorf("text should contain 'send SMS', got: %q", text)
	}
}

func TestSendSmsTool_WithDeviceID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/sms/send", func(w http.ResponseWriter, r *http.Request) {
		var req SendSmsRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.DeviceID != "dev_1" {
			t.Errorf("DeviceID = %q, want %q", req.DeviceID, "dev_1")
		}
		json.NewEncoder(w).Encode(SendSmsResponse{
			MessageIDs:      []string{"msg_1"},
			RecipientsCount: 1,
			Status:          "queued",
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "send_sms", map[string]any{
		"recipients": []string{"+1111111111"},
		"body":       "Hello",
		"device_id":  "dev_1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected IsError: %s", getToolText(t, result))
	}
}

func TestSendSmsTool_EmptyRecipients(t *testing.T) {
	mux := http.NewServeMux()
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "send_sms", map[string]any{
		"recipients": []string{},
		"body":       "Hello",
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

func TestSendSmsTool_EmptyBody(t *testing.T) {
	mux := http.NewServeMux()
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "send_sms", map[string]any{
		"recipients": []string{"+1111111111"},
		"body":       "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "body must not be empty") {
		t.Errorf("text should contain validation message, got: %q", text)
	}
}

func TestListMessagesTool_Outgoing(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_messages/records", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PaginatedResponse[SmsMessage]{
			Page: 1, PerPage: 20, TotalPages: 1, TotalItems: 1,
			Items: []SmsMessage{{
				ID: "msg_1", To: "+1234567890", Body: "Hello",
				Status: "sent", MessageType: "outgoing", Created: "2026-03-21T10:00:00Z",
			}},
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_messages", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	for _, want := range []string{"To: +1234567890", "[SENT]", "page 1/1"} {
		if !strings.Contains(text, want) {
			t.Errorf("text should contain %q, got: %q", want, text)
		}
	}
}

func TestListMessagesTool_Incoming(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_messages/records", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PaginatedResponse[SmsMessage]{
			Page: 1, PerPage: 20, TotalPages: 1, TotalItems: 1,
			Items: []SmsMessage{{
				ID: "msg_1", FromNumber: "+9876543210", Body: "Hi there",
				Status: "received", MessageType: "incoming", Created: "2026-03-21T10:00:00Z",
			}},
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_messages", map[string]any{
		"type": "incoming",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "From: +9876543210") {
		t.Errorf("text should contain 'From: +9876543210', got: %q", text)
	}
}

func TestListMessagesTool_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_messages/records", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PaginatedResponse[SmsMessage]{
			Page: 1, PerPage: 20, TotalPages: 0, TotalItems: 0,
			Items: []SmsMessage{},
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_messages", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	if text != "No messages found." {
		t.Errorf("text = %q, want %q", text, "No messages found.")
	}
}

func TestListMessagesTool_WithTypeAndStatusFilter(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_messages/records", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		if !strings.Contains(filter, `message_type="outgoing"`) {
			t.Errorf("filter should contain message_type, got %q", filter)
		}
		if !strings.Contains(filter, `status="sent"`) {
			t.Errorf("filter should contain status, got %q", filter)
		}
		json.NewEncoder(w).Encode(PaginatedResponse[SmsMessage]{
			Page: 1, PerPage: 20, TotalPages: 1, TotalItems: 1,
			Items: []SmsMessage{{
				ID: "msg_1", To: "+1234567890", Body: "Hello",
				Status: "sent", MessageType: "outgoing", Created: "2026-03-21T10:00:00Z",
			}},
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_messages", map[string]any{
		"type":   "outgoing",
		"status": "sent",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected IsError: %s", getToolText(t, result))
	}
}

func TestListMessagesTool_DefaultPageAndPerPage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_messages/records", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("page"); got != "1" {
			t.Errorf("page = %q, want %q (default)", got, "1")
		}
		if got := r.URL.Query().Get("perPage"); got != "20" {
			t.Errorf("perPage = %q, want %q (default)", got, "20")
		}
		json.NewEncoder(w).Encode(PaginatedResponse[SmsMessage]{
			Page: 1, PerPage: 20, TotalPages: 0, TotalItems: 0,
			Items: []SmsMessage{},
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	// Explicitly omit page and per_page to trigger defaults
	result, err := callTool(t, session, "list_messages", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = getToolText(t, result)
}

func TestListMessagesTool_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_messages/records", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "server error"})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_messages", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
}

func TestListMessagesTool_InvalidType(t *testing.T) {
	mux := http.NewServeMux()
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_messages", map[string]any{
		"type": "invalid",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "invalid type") {
		t.Errorf("text should contain validation message, got: %q", text)
	}
}

func TestListMessagesTool_InvalidStatus(t *testing.T) {
	mux := http.NewServeMux()
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_messages", map[string]any{
		"status": "invalid",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "invalid status") {
		t.Errorf("text should contain validation message, got: %q", text)
	}
}

func TestGetMessageTool_InvalidID(t *testing.T) {
	mux := http.NewServeMux()
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	for _, id := range []string{"", "../etc/passwd", "id/../../x", "has space"} {
		result, err := callTool(t, session, "get_message", map[string]any{
			"message_id": id,
		})
		if err != nil {
			t.Fatalf("unexpected error for id %q: %v", id, err)
		}
		if !result.IsError {
			t.Errorf("IsError should be true for id %q", id)
		}
		text := getToolText(t, result)
		if !strings.Contains(text, "invalid message_id") {
			t.Errorf("text should contain 'invalid message_id' for id %q, got: %q", id, text)
		}
	}
}

func TestGetMessageTool_Outgoing(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_messages/records/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SmsMessage{
			ID: "msg_1", To: "+1234567890", Body: "Hello",
			Status: "sent", MessageType: "outgoing", Created: "2026-03-21T10:00:00Z",
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "get_message", map[string]any{
		"message_id": "msg_1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	for _, want := range []string{"Message: msg_1", "To: +1234567890", "Type: outgoing"} {
		if !strings.Contains(text, want) {
			t.Errorf("text should contain %q, got: %q", want, text)
		}
	}
}

func TestGetMessageTool_Incoming(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_messages/records/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SmsMessage{
			ID: "msg_2", FromNumber: "+9876543210", Body: "Hi",
			Status: "received", MessageType: "incoming", Created: "2026-03-21T10:00:00Z",
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "get_message", map[string]any{
		"message_id": "msg_2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "From: +9876543210") {
		t.Errorf("text should contain 'From: +9876543210', got: %q", text)
	}
}

func TestGetMessageTool_WithOptionalFields(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_messages/records/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SmsMessage{
			ID: "msg_3", To: "+1234567890", Body: "Hello",
			Status: "failed", MessageType: "outgoing", Created: "2026-03-21T10:00:00Z",
			SentAt: "2026-03-21T10:01:00Z", DeliveredAt: "2026-03-21T10:02:00Z",
			ErrorMessage: "timeout", BatchID: "batch_1", Device: "dev_1",
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "get_message", map[string]any{
		"message_id": "msg_3",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	for _, want := range []string{"Sent at:", "Delivered at:", "Error: timeout", "Batch ID: batch_1", "Device: dev_1"} {
		if !strings.Contains(text, want) {
			t.Errorf("text should contain %q, got: %q", want, text)
		}
	}
}

func TestGetMessageTool_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_messages/records/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "get_message", map[string]any{
		"message_id": "missing",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "get message") {
		t.Errorf("text should contain 'get message', got: %q", text)
	}
}
