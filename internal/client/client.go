package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient defines the interface for making HTTP requests
type HTTPClient interface {
	Get(ctx context.Context, url string, headers map[string]string) (*Response, error)
	Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error)
	Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error)
	Patch(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error)
	Delete(ctx context.Context, url string, headers map[string]string) (*Response, error)
}

// Response wraps the HTTP response
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       []byte
}

// Client is the base HTTP client for the SDK
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new HTTP client
func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

// NewClientWithHTTPClient creates a new client with a custom HTTP client
func NewClientWithHTTPClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	return c.do(ctx, http.MethodGet, url, nil, headers)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.do(ctx, http.MethodPost, url, body, headers)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.do(ctx, http.MethodPut, url, body, headers)
}

// Patch performs a PATCH request
func (c *Client) Patch(ctx context.Context, url string, body interface{}, headers map[string]string) (*Response, error) {
	return c.do(ctx, http.MethodPatch, url, body, headers)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	return c.do(ctx, http.MethodDelete, url, nil, headers)
}

// do performs the actual HTTP request
func (c *Client) do(ctx context.Context, method, url string, body interface{}, headers map[string]string) (*Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
		Body:       respBody,
	}, nil
}

// CreateHeaders creates a standard set of headers with optional API key
func CreateHeaders(apiKey string) map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
	if apiKey != "" {
		headers["x-api-key"] = apiKey
	}
	return headers
}
