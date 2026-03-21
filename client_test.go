package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// NewVendelClient
// ---------------------------------------------------------------------------

func TestNewVendelClient_TrimsTrailingSlash(t *testing.T) {
	c := NewVendelClient("http://example.com/", "key")
	if c.baseURL != "http://example.com" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "http://example.com")
	}
	if c.apiKey != "key" {
		t.Errorf("apiKey = %q, want %q", c.apiKey, "key")
	}
}

func TestNewVendelClient_NoTrailingSlash(t *testing.T) {
	c := NewVendelClient("http://example.com", "key")
	if c.baseURL != "http://example.com" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "http://example.com")
	}
}

func TestNewVendelClient_HasTimeout(t *testing.T) {
	c := NewVendelClient("http://example.com", "key")
	if c.http.Timeout == 0 {
		t.Error("http.Client should have a non-zero timeout")
	}
}

// ---------------------------------------------------------------------------
// request() -- every branch
// ---------------------------------------------------------------------------

func TestRequest_GetNilBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want %q", got, "application/json")
		}
		if got := r.Header.Get("X-API-Key"); got != "vk_test" {
			t.Errorf("X-API-Key = %q, want %q", got, "vk_test")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "vk_test")
	data, err := c.request(context.Background(), "GET", "/ping", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["ok"] != true {
		t.Errorf(`got["ok"] = %v, want true`, got["ok"])
	}
}

func TestRequest_PostWithBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["hello"] != "world" {
			t.Errorf(`body["hello"] = %q, want %q`, body["hello"], "world")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"created":true}`))
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "vk_test")
	data, err := c.request(context.Background(), "POST", "/create", map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["created"] != true {
		t.Errorf(`got["created"] = %v, want true`, got["created"])
	}
}

func TestRequest_MarshalError(t *testing.T) {
	c := NewVendelClient("http://localhost", "key")
	// channels cannot be marshaled to JSON
	_, err := c.request(context.Background(), "POST", "/x", make(chan int))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "marshal request") {
		t.Errorf("error should contain 'marshal request', got: %v", err)
	}
}

func TestRequest_NewRequestError(t *testing.T) {
	c := NewVendelClient("http://localhost", "key")
	// method with space is invalid for http.NewRequest
	_, err := c.request(context.Background(), "BAD METHOD", "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "create request") {
		t.Errorf("error should contain 'create request', got: %v", err)
	}
}

func TestRequest_NetworkError(t *testing.T) {
	// Port 1 on localhost will refuse connections
	c := NewVendelClient("http://127.0.0.1:1", "key")
	_, err := c.request(context.Background(), "GET", "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "send request") {
		t.Errorf("error should contain 'send request', got: %v", err)
	}
}

func TestRequest_Error400_MessageKey(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "bad"})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	_, err := c.request(context.Background(), "GET", "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "API error 400") {
		t.Errorf("error should contain 'API error 400', got: %v", err)
	}
	if !strings.Contains(err.Error(), "bad") {
		t.Errorf("error should contain 'bad', got: %v", err)
	}
}

func TestRequest_Error400_DetailKey(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"detail": "missing"})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	_, err := c.request(context.Background(), "GET", "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "API error 400") {
		t.Errorf("error should contain 'API error 400', got: %v", err)
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("error should contain 'missing', got: %v", err)
	}
}

func TestRequest_Error400_ErrorKey(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "quota"})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	_, err := c.request(context.Background(), "GET", "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "API error 400") {
		t.Errorf("error should contain 'API error 400', got: %v", err)
	}
	if !strings.Contains(err.Error(), "quota") {
		t.Errorf("error should contain 'quota', got: %v", err)
	}
}

func TestRequest_Error500_NonJSONBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	_, err := c.request(context.Background(), "GET", "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "API error 500") {
		t.Errorf("error should contain 'API error 500', got: %v", err)
	}
	if !strings.Contains(err.Error(), "500 Internal Server Error") {
		t.Errorf("error should contain '500 Internal Server Error', got: %v", err)
	}
}

func TestRequest_Error400_NoMatchingKey(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"other_key": "value"})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	_, err := c.request(context.Background(), "GET", "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "API error 400") {
		t.Errorf("error should contain 'API error 400', got: %v", err)
	}
	// Falls through to the status-line fallback
	if !strings.Contains(err.Error(), "400 Bad Request") {
		t.Errorf("error should contain '400 Bad Request', got: %v", err)
	}
}

func TestRequest_Success200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	data, err := c.request(context.Background(), "GET", "/ok", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `{"status":"ok"}` {
		t.Errorf("data = %q, want %q", string(data), `{"status":"ok"}`)
	}
}

// ---------------------------------------------------------------------------
// get() / post() wrappers
// ---------------------------------------------------------------------------

func TestGet_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/test" {
			t.Errorf("path = %s, want /api/test", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"key": "value"})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	var result map[string]string
	err := c.get(context.Background(), "/api/test", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf(`result["key"] = %q, want %q`, result["key"], "value")
	}
}

func TestGet_RequestError(t *testing.T) {
	c := NewVendelClient("http://127.0.0.1:1", "key")
	var result map[string]string
	err := c.get(context.Background(), "/api/test", &result)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "send request") {
		t.Errorf("error should contain 'send request', got: %v", err)
	}
}

func TestGet_UnmarshalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	var result map[string]string
	err := c.get(context.Background(), "/api/test", &result)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPost_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/create" {
			t.Errorf("path = %s, want /api/create", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"id": "abc123"})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	var result map[string]string
	err := c.post(context.Background(), "/api/create", map[string]string{"name": "test"}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["id"] != "abc123" {
		t.Errorf(`result["id"] = %q, want %q`, result["id"], "abc123")
	}
}

func TestPost_RequestError(t *testing.T) {
	c := NewVendelClient("http://127.0.0.1:1", "key")
	var result map[string]string
	err := c.post(context.Background(), "/api/create", map[string]string{"a": "b"}, &result)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "send request") {
		t.Errorf("error should contain 'send request', got: %v", err)
	}
}

func TestPost_UnmarshalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	var result map[string]string
	err := c.post(context.Background(), "/api/create", map[string]string{"a": "b"}, &result)
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// SendSms
// ---------------------------------------------------------------------------

func TestSendSms_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sms/send" {
			t.Errorf("path = %s, want /api/sms/send", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}

		var req SendSmsRequest
		json.NewDecoder(r.Body).Decode(&req)
		if len(req.Recipients) != 1 || req.Recipients[0] != "+1234567890" {
			t.Errorf("Recipients = %v, want [+1234567890]", req.Recipients)
		}
		if req.Body != "Hello" {
			t.Errorf("Body = %q, want %q", req.Body, "Hello")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SendSmsResponse{
			BatchID:         "batch_1",
			MessageIDs:      []string{"msg_1"},
			RecipientsCount: 1,
			Status:          "queued",
		})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "vk_test")
	resp, err := c.SendSms(context.Background(), &SendSmsRequest{
		Recipients: []string{"+1234567890"},
		Body:       "Hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.BatchID != "batch_1" {
		t.Errorf("BatchID = %q, want %q", resp.BatchID, "batch_1")
	}
	if len(resp.MessageIDs) != 1 || resp.MessageIDs[0] != "msg_1" {
		t.Errorf("MessageIDs = %v, want [msg_1]", resp.MessageIDs)
	}
	if resp.RecipientsCount != 1 {
		t.Errorf("RecipientsCount = %d, want 1", resp.RecipientsCount)
	}
	if resp.Status != "queued" {
		t.Errorf("Status = %q, want %q", resp.Status, "queued")
	}
}

func TestSendSms_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "quota exceeded"})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "vk_test")
	resp, err := c.SendSms(context.Background(), &SendSmsRequest{
		Recipients: []string{"+1"},
		Body:       "test",
	})
	if resp != nil {
		t.Errorf("expected nil response, got %v", resp)
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "quota exceeded") {
		t.Errorf("error should contain 'quota exceeded', got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetQuota
// ---------------------------------------------------------------------------

func TestGetQuota_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/plans/quota" {
			t.Errorf("path = %s, want /api/plans/quota", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("method = %s, want GET", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(QuotaResponse{
			Plan:              "pro",
			SmsSentThisMonth:  42,
			MaxSmsPerMonth:    1000,
			DevicesRegistered: 2,
			MaxDevices:        5,
			ResetDate:         "2026-04-01",
		})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "vk_test")
	q, err := c.GetQuota(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Plan != "pro" {
		t.Errorf("Plan = %q, want %q", q.Plan, "pro")
	}
	if q.SmsSentThisMonth != 42 {
		t.Errorf("SmsSentThisMonth = %d, want 42", q.SmsSentThisMonth)
	}
	if q.MaxSmsPerMonth != 1000 {
		t.Errorf("MaxSmsPerMonth = %d, want 1000", q.MaxSmsPerMonth)
	}
	if q.DevicesRegistered != 2 {
		t.Errorf("DevicesRegistered = %d, want 2", q.DevicesRegistered)
	}
	if q.MaxDevices != 5 {
		t.Errorf("MaxDevices = %d, want 5", q.MaxDevices)
	}
	if q.ResetDate != "2026-04-01" {
		t.Errorf("ResetDate = %q, want %q", q.ResetDate, "2026-04-01")
	}
}

func TestGetQuota_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "unauthorized"})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "vk_test")
	q, err := c.GetQuota(context.Background())
	if q != nil {
		t.Errorf("expected nil quota, got %v", q)
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("error should contain 'unauthorized', got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ListParams.Encode()
// ---------------------------------------------------------------------------

func TestListParamsEncode(t *testing.T) {
	tests := []struct {
		name     string
		params   ListParams
		expected []string // substrings that must appear in the output
		empty    bool     // true if we expect ""
	}{
		{
			name: "all params set",
			params: ListParams{
				Filter:  `status="sent"`,
				Page:    2,
				PerPage: 50,
				Sort:    "-created",
			},
			expected: []string{
				"filter=",
				"page=2",
				"perPage=50",
				"sort=-created",
			},
		},
		{
			name:   "no params",
			params: ListParams{},
			empty:  true,
		},
		{
			name: "only filter",
			params: ListParams{
				Filter: `name="test"`,
			},
			expected: []string{"?", "filter="},
		},
		{
			name: "only page",
			params: ListParams{
				Page: 3,
			},
			expected: []string{"?", "page=3"},
		},
		{
			name: "only perPage",
			params: ListParams{
				PerPage: 10,
			},
			expected: []string{"?", "perPage=10"},
		},
		{
			name: "only sort",
			params: ListParams{
				Sort: "name",
			},
			expected: []string{"?", "sort=name"},
		},
		{
			name: "page and sort only",
			params: ListParams{
				Page: 1,
				Sort: "-updated",
			},
			expected: []string{"?", "page=1", "sort=-updated"},
		},
		{
			name: "zero page and perPage are excluded",
			params: ListParams{
				Page:    0,
				PerPage: 0,
				Filter:  "x",
			},
			expected: []string{"?", "filter=x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.Encode()
			if tt.empty {
				if result != "" {
					t.Errorf("expected empty string, got %q", result)
				}
				return
			}
			if len(result) == 0 {
				t.Error("expected non-empty query string")
			}
			if result[0] != '?' {
				t.Errorf("query string must start with '?', got %q", string(result[0]))
			}
			for _, s := range tt.expected {
				if !strings.Contains(result, s) {
					t.Errorf("result %q should contain %q", result, s)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// listRecords (generic, package-level)
// ---------------------------------------------------------------------------

func TestListRecords_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/sms_devices/records" {
			t.Errorf("path = %s, want /api/collections/sms_devices/records", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("method = %s, want GET", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PaginatedResponse[SmsDevice]{
			Page:       1,
			PerPage:    20,
			TotalPages: 1,
			TotalItems: 1,
			Items: []SmsDevice{
				{ID: "dev_1", Name: "Phone", DeviceType: "android", PhoneNumber: "+1111111111"},
			},
		})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "vk_test")
	result, err := listRecords[SmsDevice](context.Background(), c, "sms_devices", &ListParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalItems != 1 {
		t.Errorf("TotalItems = %d, want 1", result.TotalItems)
	}
	if len(result.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(result.Items))
	}
	if result.Items[0].ID != "dev_1" {
		t.Errorf("Items[0].ID = %q, want %q", result.Items[0].ID, "dev_1")
	}
}

func TestListRecords_WithParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.RawQuery
		if !strings.Contains(q, "page=2") {
			t.Errorf("query should contain page=2, got %q", q)
		}
		if !strings.Contains(q, "perPage=5") {
			t.Errorf("query should contain perPage=5, got %q", q)
		}
		if !strings.Contains(q, "sort=-created") {
			t.Errorf("query should contain sort=-created, got %q", q)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(PaginatedResponse[SmsDevice]{
			Page: 2, PerPage: 5, TotalPages: 3, TotalItems: 12,
			Items: []SmsDevice{},
		})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "vk_test")
	result, err := listRecords[SmsDevice](context.Background(), c, "sms_devices", &ListParams{
		Page:    2,
		PerPage: 5,
		Sort:    "-created",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Page != 2 {
		t.Errorf("Page = %d, want 2", result.Page)
	}
}

func TestListRecords_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "forbidden"})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "vk_test")
	result, err := listRecords[SmsDevice](context.Background(), c, "sms_devices", &ListParams{})
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("error should contain 'forbidden', got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// getRecord (generic, package-level)
// ---------------------------------------------------------------------------

func TestGetRecord_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/sms_messages/records/msg_42" {
			t.Errorf("path = %s, want /api/collections/sms_messages/records/msg_42", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("method = %s, want GET", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SmsMessage{
			ID:          "msg_42",
			To:          "+1234567890",
			Body:        "Test message",
			Status:      "sent",
			MessageType: "outgoing",
			Created:     "2026-03-21T10:00:00Z",
		})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "vk_test")
	msg, err := getRecord[SmsMessage](context.Background(), c, "sms_messages", "msg_42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != "msg_42" {
		t.Errorf("ID = %q, want %q", msg.ID, "msg_42")
	}
	if msg.Body != "Test message" {
		t.Errorf("Body = %q, want %q", msg.Body, "Test message")
	}
	if msg.Status != "sent" {
		t.Errorf("Status = %q, want %q", msg.Status, "sent")
	}
}

func TestGetRecord_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "vk_test")
	msg, err := getRecord[SmsMessage](context.Background(), c, "sms_messages", "missing_id")
	if msg != nil {
		t.Errorf("expected nil, got %v", msg)
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should contain 'not found', got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// createRecord (generic, package-level)
// ---------------------------------------------------------------------------

func TestCreateRecord_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/sms_templates/records" {
			t.Errorf("path = %s, want /api/collections/sms_templates/records", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "greeting" {
			t.Errorf(`body["name"] = %q, want %q`, body["name"], "greeting")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SmsTemplate{
			ID:      "tpl_1",
			Name:    "greeting",
			Body:    "Hello {{name}}",
			Created: "2026-03-21T12:00:00Z",
		})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "vk_test")
	tpl, err := createRecord[SmsTemplate](context.Background(), c, "sms_templates", map[string]string{
		"name": "greeting",
		"body": "Hello {{name}}",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tpl.ID != "tpl_1" {
		t.Errorf("ID = %q, want %q", tpl.ID, "tpl_1")
	}
	if tpl.Name != "greeting" {
		t.Errorf("Name = %q, want %q", tpl.Name, "greeting")
	}
}

func TestCreateRecord_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "validation failed"})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "vk_test")
	tpl, err := createRecord[SmsTemplate](context.Background(), c, "sms_templates", map[string]string{"name": ""})
	if tpl != nil {
		t.Errorf("expected nil, got %v", tpl)
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("error should contain 'validation failed', got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// request() edge case: JSON error body with empty string value for known key
// ---------------------------------------------------------------------------

func TestRequest_Error400_EmptyMessageValue(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		// "message" key exists but value is empty -- should skip to fallback
		json.NewEncoder(w).Encode(map[string]string{"message": ""})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	_, err := c.request(context.Background(), "GET", "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "API error 400") {
		t.Errorf("error should contain 'API error 400', got: %v", err)
	}
	if !strings.Contains(err.Error(), "400 Bad Request") {
		t.Errorf("error should contain '400 Bad Request', got: %v", err)
	}
}

// The "message" key has a non-string value (number) -- the .(string) assertion
// fails, so it falls through to the status-line fallback.
func TestRequest_Error400_NonStringMessageValue(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"message": 123})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	_, err := c.request(context.Background(), "GET", "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "API error 400") {
		t.Errorf("error should contain 'API error 400', got: %v", err)
	}
	if !strings.Contains(err.Error(), "400 Bad Request") {
		t.Errorf("error should contain '400 Bad Request', got: %v", err)
	}
}

// Verify that the first matching key wins (message before detail).
func TestRequest_Error400_MessageAndDetailBothPresent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "from_message",
			"detail":  "from_detail",
		})
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	_, err := c.request(context.Background(), "GET", "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	// "message" is checked first
	if !strings.Contains(err.Error(), "from_message") {
		t.Errorf("error should contain 'from_message' (message key wins), got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Context cancellation
// ---------------------------------------------------------------------------

func TestRequest_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	_, err := c.request(ctx, "GET", "/x", nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

// ---------------------------------------------------------------------------
// Response size limit (io.LimitReader)
// ---------------------------------------------------------------------------

func TestRequest_ResponseSizeLimited(t *testing.T) {
	overflow := 1024
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		data := make([]byte, maxResponseSize+overflow)
		for i := range data {
			data[i] = 'x'
		}
		w.Write(data)
	}))
	defer ts.Close()

	c := NewVendelClient(ts.URL, "key")
	body, err := c.request(context.Background(), "GET", "/big", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(body) != maxResponseSize {
		t.Errorf("body length = %d, want %d (maxResponseSize)", len(body), maxResponseSize)
	}
}
