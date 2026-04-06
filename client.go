package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxResponseSize = 10 << 20 // 10 MB

// VendelClient is an HTTP client wrapper for the Vendel API.
type VendelClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewVendelClient creates a new Vendel API client.
func NewVendelClient(baseURL, apiKey string) *VendelClient {
	return &VendelClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *VendelClient) request(ctx context.Context, method, path string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr map[string]any
		if json.Unmarshal(respBody, &apiErr) == nil {
			for _, key := range []string{"message", "detail", "error"} {
				if msg, ok := apiErr[key].(string); ok && msg != "" {
					return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, msg)
				}
			}
		}
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, resp.Status)
	}

	return respBody, nil
}

func (c *VendelClient) get(ctx context.Context, path string, result any) error {
	data, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

func (c *VendelClient) post(ctx context.Context, path string, body, result any) error {
	data, err := c.request(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

// SendSms sends an SMS via the Vendel API.
func (c *VendelClient) SendSms(ctx context.Context, input *SendSmsRequest) (*SendSmsResponse, error) {
	var result SendSmsResponse
	if err := c.post(ctx, "/api/sms/send", input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SendSmsTemplate sends an SMS using a saved template via the Vendel API.
func (c *VendelClient) SendSmsTemplate(ctx context.Context, input *SendSmsTemplateRequest) (*SendSmsResponse, error) {
	var result SendSmsResponse
	if err := c.post(ctx, "/api/sms/send-template", input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetQuota returns the current user's quota and plan info.
func (c *VendelClient) GetQuota(ctx context.Context) (*QuotaResponse, error) {
	var result QuotaResponse
	if err := c.get(ctx, "/api/plans/quota", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMessageStatus returns the status of a single SMS message.
func (c *VendelClient) GetMessageStatus(ctx context.Context, messageID string) (*MessageStatusResponse, error) {
	var result MessageStatusResponse
	if err := c.get(ctx, "/api/sms/status/"+messageID, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBatchStatus returns the status of all messages in a batch.
func (c *VendelClient) GetBatchStatus(ctx context.Context, batchID string) (*BatchStatusResponse, error) {
	var result BatchStatusResponse
	if err := c.get(ctx, "/api/sms/batch/"+batchID, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListContacts lists contacts with optional search and group filter.
func (c *VendelClient) ListContacts(ctx context.Context, page, perPage int, search, groupID string) (*ContactListResponse, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		params.Set("per_page", strconv.Itoa(perPage))
	}
	if search != "" {
		params.Set("search", search)
	}
	if groupID != "" {
		params.Set("group_id", groupID)
	}
	path := "/api/contacts"
	if q := params.Encode(); q != "" {
		path += "?" + q
	}
	var result ContactListResponse
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListContactGroups lists contact groups.
func (c *VendelClient) ListContactGroups(ctx context.Context, page, perPage int) (*ContactGroupListResponse, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		params.Set("per_page", strconv.Itoa(perPage))
	}
	path := "/api/contacts/groups"
	if q := params.Encode(); q != "" {
		path += "?" + q
	}
	var result ContactGroupListResponse
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListParams holds query parameters for collection list endpoints.
type ListParams struct {
	Filter  string
	Page    int
	PerPage int
	Sort    string
}

// Encode encodes the parameters into a query string.
func (p *ListParams) Encode() string {
	params := url.Values{}
	if p.Filter != "" {
		params.Set("filter", p.Filter)
	}
	if p.Page > 0 {
		params.Set("page", strconv.Itoa(p.Page))
	}
	if p.PerPage > 0 {
		params.Set("perPage", strconv.Itoa(p.PerPage))
	}
	if p.Sort != "" {
		params.Set("sort", p.Sort)
	}
	encoded := params.Encode()
	if encoded != "" {
		return "?" + encoded
	}
	return ""
}

// listRecords fetches a paginated list from a PocketBase collection.
func listRecords[T any](ctx context.Context, c *VendelClient, collection string, params *ListParams) (*PaginatedResponse[T], error) {
	query := params.Encode()
	path := fmt.Sprintf("/api/collections/%s/records%s", collection, query)
	var result PaginatedResponse[T]
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// getRecord fetches a single record from a PocketBase collection.
func getRecord[T any](ctx context.Context, c *VendelClient, collection, id string) (*T, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s", collection, id)
	var result T
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// createRecord creates a record in a PocketBase collection.
func createRecord[T any](ctx context.Context, c *VendelClient, collection string, data any) (*T, error) {
	path := fmt.Sprintf("/api/collections/%s/records", collection)
	var result T
	if err := c.post(ctx, path, data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
