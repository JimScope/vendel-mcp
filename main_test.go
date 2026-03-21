package main

import (
	"strings"
	"testing"
)

func TestCreateServer(t *testing.T) {
	t.Run("empty baseURL", func(t *testing.T) {
		srv, err := createServer("", "vk_123")
		if srv != nil {
			t.Error("expected nil server")
		}
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "VENDEL_URL") {
			t.Errorf("error should mention VENDEL_URL, got: %v", err)
		}
	})

	t.Run("empty apiKey", func(t *testing.T) {
		srv, err := createServer("http://localhost:8090", "")
		if srv != nil {
			t.Error("expected nil server")
		}
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "VENDEL_API_KEY") {
			t.Errorf("error should mention VENDEL_API_KEY, got: %v", err)
		}
	})

	t.Run("both empty", func(t *testing.T) {
		srv, err := createServer("", "")
		if srv != nil {
			t.Error("expected nil server")
		}
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "environment variables are required") {
			t.Errorf("error should mention 'environment variables are required', got: %v", err)
		}
	})

	t.Run("valid params", func(t *testing.T) {
		srv, err := createServer("http://localhost:8090", "vk_test123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if srv == nil {
			t.Fatal("expected non-nil server")
		}
	})
}
