package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

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
		http:    &http.Client{},
	}
}

func (c *VendelClient) request(method, path string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
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

	respBody, err := io.ReadAll(resp.Body)
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

func (c *VendelClient) get(path string, result any) error {
	data, err := c.request("GET", path, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

func (c *VendelClient) post(path string, body, result any) error {
	data, err := c.request("POST", path, body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

// SendSms sends an SMS via the Vendel API.
func (c *VendelClient) SendSms(input *SendSmsRequest) (*SendSmsResponse, error) {
	var result SendSmsResponse
	if err := c.post("/api/sms/send", input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetQuota returns the current user's quota and plan info.
func (c *VendelClient) GetQuota() (*QuotaResponse, error) {
	var result QuotaResponse
	if err := c.get("/api/plans/quota", &result); err != nil {
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
	// A bit ugly, but it works.
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
func listRecords[T any](c *VendelClient, collection string, params *ListParams) (*PaginatedResponse[T], error) {
	query := params.encode()
	path := fmt.Sprintf("/api/collections/%s/records%s", collection, query)
	var result PaginatedResponse[T]
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// getRecord fetches a single record from a PocketBase collection.
func getRecord[T any](c *VendelClient, collection, id string) (*T, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s", collection, id)
	var result T
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// createRecord creates a record in a PocketBase collection.
func createRecord[T any](c *VendelClient, collection string, data any) (*T, error) {
	path := fmt.Sprintf("/api/collections/%s/records", collection)
	var result T
	if err := c.post(path, data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
