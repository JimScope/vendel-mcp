package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCheckQuotaTool_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/plans/quota", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(QuotaResponse{
			Plan: "pro", SmsSentThisMonth: 42, MaxSmsPerMonth: 1000,
			DevicesRegistered: 2, MaxDevices: 5, ResetDate: "2026-04-01",
			ScheduledSmsActive: 3, MaxScheduledSms: 10,
			IntegrationsCreated: 1, MaxIntegrations: 5,
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "check_quota", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := getToolText(t, result)
	for _, want := range []string{
		"Plan: pro",
		"SMS: 42/1000 used (958 remaining)",
		"Devices: 2/5",
		"Scheduled SMS: 3/10",
		"Integrations: 1/5",
		"Quota resets: 2026-04-01",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("text should contain %q, got: %q", want, text)
		}
	}
}

func TestCheckQuotaTool_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/plans/quota", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "unauthorized"})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	result, err := callTool(t, session, "check_quota", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("IsError should be true")
	}
	text := getToolText(t, result)
	if !strings.Contains(text, "check quota") {
		t.Errorf("text should contain 'check quota', got: %q", text)
	}
}

func TestQuotaResource_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/plans/quota", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(QuotaResponse{
			Plan: "pro", SmsSentThisMonth: 42, MaxSmsPerMonth: 1000,
			DevicesRegistered: 2, MaxDevices: 5, ResetDate: "2026-04-01",
		})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	res, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "vendel://quota",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Contents) == 0 {
		t.Fatal("expected at least one content element")
	}
	if res.Contents[0].MIMEType != "application/json" {
		t.Errorf("MIMEType = %q, want %q", res.Contents[0].MIMEType, "application/json")
	}
	if !strings.Contains(res.Contents[0].Text, "pro") {
		t.Errorf("text should contain 'pro', got: %q", res.Contents[0].Text)
	}
}

func TestQuotaResource_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/plans/quota", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "unauthorized"})
	})
	session, cleanup := setupMCP(t, mux)
	defer cleanup()

	_, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "vendel://quota",
	})
	if err == nil {
		t.Fatal("expected error from resource handler")
	}
}
