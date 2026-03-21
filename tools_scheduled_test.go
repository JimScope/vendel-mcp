package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestScheduleSmsTool_Success_AllFields(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/scheduled_sms/records", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["timezone"] != "America/New_York" {
			t.Errorf("timezone = %v, want America/New_York", body["timezone"])
		}
		if body["device_id"] != "dev_1" {
			t.Errorf("device_id = %v, want dev_1", body["device_id"])
		}
		if body["scheduled_at"] != "2026-04-01T10:00:00Z" {
			t.Errorf("scheduled_at = %v, want 2026-04-01T10:00:00Z", body["scheduled_at"])
		}
		if body["cron_expression"] != "0 9 * * MON" {
			t.Errorf("cron_expression = %v, want '0 9 * * MON'", body["cron_expression"])
		}
		json.NewEncoder(w).Encode(ScheduledSms{
			ID: "sch_1", Name: "Weekly Alert", ScheduleType: "recurring",
			Recipients: []string{"+1111111111"}, Body: "Hello",
			Timezone: "America/New_York", ScheduledAt: "2026-04-01T10:00:00Z",
			CronExpression: "0 9 * * MON", NextRunAt: "2026-04-07T09:00:00Z",
			Status: "active",
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "schedule_sms", map[string]any{
		"name":            "Weekly Alert",
		"recipients":      []string{"+1111111111"},
		"body":            "Hello",
		"schedule_type":   "recurring",
		"scheduled_at":    "2026-04-01T10:00:00Z",
		"cron_expression": "0 9 * * MON",
		"timezone":        "America/New_York",
		"device_id":       "dev_1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	for _, want := range []string{
		"SMS scheduled",
		"ID: sch_1",
		"Name: Weekly Alert",
		"Type: recurring",
		"Timezone: America/New_York",
		"Scheduled for: 2026-04-01T10:00:00Z",
		"Cron: 0 9 * * MON",
		"Next run: 2026-04-07T09:00:00Z",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("text should contain %q, got: %q", want, text)
		}
	}
}

func TestScheduleSmsTool_Success_DefaultTimezone(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/scheduled_sms/records", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["timezone"] != "UTC" {
			t.Errorf("timezone should default to UTC, got %v", body["timezone"])
		}
		json.NewEncoder(w).Encode(ScheduledSms{
			ID: "sch_2", Name: "One-time", ScheduleType: "one_time",
			Recipients: []string{"+1111111111"}, Body: "Hello",
			Timezone: "UTC", Status: "active",
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "schedule_sms", map[string]any{
		"name":          "One-time",
		"recipients":    []string{"+1111111111"},
		"body":          "Hello",
		"schedule_type": "one_time",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "Timezone: UTC") {
		t.Errorf("text should contain 'Timezone: UTC', got: %q", text)
	}
	// Should NOT have optional fields when not set in response
	for _, absent := range []string{"Scheduled for:", "Cron:", "Next run:"} {
		if strings.Contains(text, absent) {
			t.Errorf("text should not contain %q, got: %q", absent, text)
		}
	}
}

func TestScheduleSmsTool_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/scheduled_sms/records", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "validation failed"})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "schedule_sms", map[string]any{
		"name":          "Test",
		"recipients":    []string{"+1111111111"},
		"body":          "Hello",
		"schedule_type": "one_time",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "schedule SMS") {
		t.Errorf("text should contain 'schedule SMS', got: %q", text)
	}
}

func TestScheduleSmsTool_EmptyName(t *testing.T) {
	mux := http.NewServeMux()
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "schedule_sms", map[string]any{
		"name":          "",
		"recipients":    []string{"+1111111111"},
		"body":          "Hello",
		"schedule_type": "one_time",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "name must not be empty") {
		t.Errorf("text should contain validation message, got: %q", text)
	}
}

func TestScheduleSmsTool_EmptyRecipients(t *testing.T) {
	mux := http.NewServeMux()
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "schedule_sms", map[string]any{
		"name":          "Test",
		"recipients":    []string{},
		"body":          "Hello",
		"schedule_type": "one_time",
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

func TestScheduleSmsTool_EmptyBody(t *testing.T) {
	mux := http.NewServeMux()
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "schedule_sms", map[string]any{
		"name":          "Test",
		"recipients":    []string{"+1111111111"},
		"body":          "",
		"schedule_type": "one_time",
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

func TestScheduleSmsTool_InvalidScheduleType(t *testing.T) {
	mux := http.NewServeMux()
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "schedule_sms", map[string]any{
		"name":          "Test",
		"recipients":    []string{"+1111111111"},
		"body":          "Hello",
		"schedule_type": "invalid",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "invalid schedule_type") {
		t.Errorf("text should contain validation message, got: %q", text)
	}
}

func TestListScheduledTool_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/scheduled_sms/records", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PaginatedResponse[ScheduledSms]{
			Page: 1, PerPage: 20, TotalPages: 1, TotalItems: 1,
			Items: []ScheduledSms{{
				ID: "sch_1", Name: "Weekly Alert", Status: "active",
				ScheduleType: "recurring", Recipients: []string{"+1111111111"},
				NextRunAt: "2026-04-07T09:00:00Z", LastRunAt: "2026-03-31T09:00:00Z",
				CronExpression: "0 9 * * MON", Timezone: "America/New_York",
			}},
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_scheduled", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	for _, want := range []string{
		"Scheduled SMS (1):",
		"Weekly Alert [ACTIVE]",
		"Next run: 2026-04-07T09:00:00Z",
		"Last run: 2026-03-31T09:00:00Z",
		"Cron: 0 9 * * MON (America/New_York)",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("text should contain %q, got: %q", want, text)
		}
	}
}

func TestListScheduledTool_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/scheduled_sms/records", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PaginatedResponse[ScheduledSms]{
			Page: 1, PerPage: 20, TotalPages: 0, TotalItems: 0,
			Items: []ScheduledSms{},
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_scheduled", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	if text != "No scheduled SMS found." {
		t.Errorf("text = %q, want %q", text, "No scheduled SMS found.")
	}
}

func TestListScheduledTool_WithStatusFilter(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/scheduled_sms/records", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		if !strings.Contains(filter, `status="active"`) {
			t.Errorf("filter should contain status, got %q", filter)
		}
		json.NewEncoder(w).Encode(PaginatedResponse[ScheduledSms]{
			Page: 1, PerPage: 20, TotalPages: 1, TotalItems: 1,
			Items: []ScheduledSms{{
				ID: "sch_1", Name: "Test", Status: "active",
				ScheduleType: "one_time", Recipients: []string{"+1111111111"},
			}},
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_scheduled", map[string]any{
		"status": "active",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected IsError: %s", getToolText(t, result))
	}
}

func TestListScheduledTool_InvalidStatus(t *testing.T) {
	mux := http.NewServeMux()
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_scheduled", map[string]any{
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

func TestListScheduledTool_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/scheduled_sms/records", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "server error"})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_scheduled", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "list scheduled SMS") {
		t.Errorf("text should contain 'list scheduled SMS', got: %q", text)
	}
}
