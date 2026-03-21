package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestListDevicesTool_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_devices/records", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PaginatedResponse[SmsDevice]{
			Page: 1, PerPage: 20, TotalPages: 1, TotalItems: 2,
			Items: []SmsDevice{
				{ID: "dev_1", Name: "Phone A", DeviceType: "android", PhoneNumber: "+1111111111"},
				{ID: "dev_2", Name: "Modem B", DeviceType: "modem", PhoneNumber: "+2222222222"},
			},
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_devices", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	for _, want := range []string{"Devices (2):", "Phone A (android)", "Modem B (modem)"} {
		if !strings.Contains(text, want) {
			t.Errorf("text should contain %q, got: %q", want, text)
		}
	}
}

func TestListDevicesTool_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_devices/records", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PaginatedResponse[SmsDevice]{
			Page: 1, PerPage: 20, TotalPages: 0, TotalItems: 0,
			Items: []SmsDevice{},
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_devices", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	if text != "No devices registered." {
		t.Errorf("text = %q, want %q", text, "No devices registered.")
	}
}

func TestListDevicesTool_WithFilter(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_devices/records", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		if !strings.Contains(filter, `device_type="android"`) {
			t.Errorf("filter should contain device_type, got %q", filter)
		}
		json.NewEncoder(w).Encode(PaginatedResponse[SmsDevice]{
			Page: 1, PerPage: 20, TotalPages: 1, TotalItems: 1,
			Items: []SmsDevice{
				{ID: "dev_1", Name: "Phone A", DeviceType: "android", PhoneNumber: "+1111111111"},
			},
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_devices", map[string]any{
		"device_type": "android",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected IsError: %s", getToolText(t, result))
	}
}

func TestListDevicesTool_InvalidDeviceType(t *testing.T) {
	mux := http.NewServeMux()
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_devices", map[string]any{
		"device_type": "satellite",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "invalid device_type") {
		t.Errorf("text should contain validation message, got: %q", text)
	}
}

func TestListDevicesTool_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/sms_devices/records", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "server error"})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "list_devices", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "list devices") {
		t.Errorf("text should contain 'list devices', got: %q", text)
	}
}
