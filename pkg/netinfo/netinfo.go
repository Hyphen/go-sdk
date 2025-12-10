package netinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Hyphen/go-sdk/internal/client"
)

// Location represents the geographic location information
type Location struct {
	Country    string  `json:"country"`
	Region     string  `json:"region"`
	City       string  `json:"city"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	PostalCode string  `json:"postalCode"`
	Timezone   string  `json:"timezone"`
	GeonameID  int     `json:"geonameId"`
}

// IPInfo represents information about an IP address
type IPInfo struct {
	IP       string   `json:"ip"`
	Type     string   `json:"type"`
	Location Location `json:"location"`
}

// IPInfoError represents an error response for IP information
type IPInfoError struct {
	IP           string `json:"ip"`
	Type         string `json:"type"`
	ErrorMessage string `json:"errorMessage"`
}

// IPInfosResponse represents the response for bulk IP info requests
type IPInfosResponse struct {
	Data []interface{} `json:"data"`
}

// Options represents configuration options for the NetInfo client
type Options struct {
	APIKey  string
	BaseURI string
}

// Option is a functional option for configuring the NetInfo client
type Option func(*Options)

// WithAPIKey sets the API key for authentication
func WithAPIKey(key string) Option {
	return func(o *Options) {
		o.APIKey = key
	}
}

// WithBaseURI sets the base URI for the NetInfo service
func WithBaseURI(uri string) Option {
	return func(o *Options) {
		o.BaseURI = uri
	}
}

// NetInfo is the client for geo information services
type NetInfo struct {
	apiKey       string
	baseURI      string
	client       *client.Client
	errorHandler func(error)
}

// New creates a new NetInfo client with functional options
func New(options ...Option) (*NetInfo, error) {
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	// Get API key from options or environment
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("HYPHEN_API_KEY")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key is required. Please provide it via options or set the HYPHEN_API_KEY environment variable")
	}

	if strings.HasPrefix(apiKey, "public_") {
		return nil, fmt.Errorf("the provided API key is a public API key. Please provide a valid non-public API key for authentication")
	}

	// Set base URI, default to "https://net.info"
	baseURI := opts.BaseURI
	if baseURI == "" {
		baseURI = "https://net.info"
	}

	n := &NetInfo{
		apiKey:  apiKey,
		baseURI: baseURI,
		client:  client.NewClient(baseURI),
	}

	return n, nil
}

// SetErrorHandler sets a custom error handler function
func (n *NetInfo) SetErrorHandler(handler func(error)) {
	n.errorHandler = handler
}

// emitError calls the error handler if set
func (n *NetInfo) emitError(err error) {
	if n.errorHandler != nil {
		n.errorHandler(err)
	}
}

// GetIPInfo fetches GeoIP information for a given IP address
func (n *NetInfo) GetIPInfo(ctx context.Context, ip string) (*IPInfo, error) {
	url := fmt.Sprintf("%s/ip/%s", strings.TrimSuffix(n.baseURI, "/"), ip)
	headers := client.CreateHeaders(n.apiKey)

	resp, err := n.client.Get(ctx, url, headers)
	if err != nil {
		err = fmt.Errorf("failed to fetch ip info: %w", err)
		n.emitError(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to fetch ip info: HTTP %d: %s", resp.StatusCode, resp.Status)
		n.emitError(err)
		return nil, err
	}

	var ipInfo IPInfo
	if err := json.Unmarshal(resp.Body, &ipInfo); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		n.emitError(err)
		return nil, err
	}

	return &ipInfo, nil
}

// GetIPInfos fetches GeoIP information for multiple IP addresses
func (n *NetInfo) GetIPInfos(ctx context.Context, ips []string) ([]interface{}, error) {
	if len(ips) == 0 {
		err := fmt.Errorf("the provided IPs array is invalid. It should be a non-empty array of strings")
		n.emitError(err)
		return nil, err
	}

	url := fmt.Sprintf("%s/ip", strings.TrimSuffix(n.baseURI, "/"))
	headers := client.CreateHeaders(n.apiKey)

	resp, err := n.client.Post(ctx, url, ips, headers)
	if err != nil {
		err = fmt.Errorf("failed to fetch ip infos: %w", err)
		n.emitError(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to fetch ip infos: HTTP %d: %s", resp.StatusCode, resp.Status)
		n.emitError(err)
		return nil, err
	}

	var response IPInfosResponse
	if err := json.Unmarshal(resp.Body, &response); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		n.emitError(err)
		return nil, err
	}

	return response.Data, nil
}
