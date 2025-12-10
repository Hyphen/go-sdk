package link

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Hyphen/go-sdk/internal/client"
)

// QRSize represents the size of a QR code
type QRSize string

const (
	QRSizeSmall  QRSize = "small"
	QRSizeMedium QRSize = "medium"
	QRSizeLarge  QRSize = "large"
)

// OrganizationID represents the organization identifier in responses
type OrganizationID struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CreateShortCodeOptions contains optional parameters for creating short codes
type CreateShortCodeOptions struct {
	Code  string   `json:"code,omitempty"`
	Title string   `json:"title,omitempty"`
	Tags  []string `json:"tags,omitempty"`
}

// ShortCodeResponse represents a short code response
type ShortCodeResponse struct {
	ID             string         `json:"id"`
	Code           string         `json:"code"`
	LongURL        string         `json:"long_url"`
	Domain         string         `json:"domain"`
	CreatedAt      string         `json:"createdAt"`
	Title          string         `json:"title,omitempty"`
	Tags           []string       `json:"tags,omitempty"`
	OrganizationID OrganizationID `json:"organizationId"`
}

// UpdateShortCodeOptions contains optional parameters for updating short codes
type UpdateShortCodeOptions struct {
	LongURL string   `json:"long_url,omitempty"`
	Title   string   `json:"title,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// GetShortCodesResponse represents a paginated response of short codes
type GetShortCodesResponse struct {
	Total    int                 `json:"total"`
	PageNum  int                 `json:"pageNum"`
	PageSize int                 `json:"pageSize"`
	Data     []ShortCodeResponse `json:"data"`
}

// CreateQRCodeOptions contains optional parameters for creating QR codes
type CreateQRCodeOptions struct {
	Title           string `json:"title,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	Color           string `json:"color,omitempty"`
	Size            QRSize `json:"size,omitempty"`
	Logo            string `json:"logo,omitempty"`
}

// QRCodeResponse represents a QR code response
type QRCodeResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title,omitempty"`
	QRCode      string `json:"qrCode"`
	QRCodeBytes []byte `json:"-"`
	QRLink      string `json:"qrLink"`
}

// GetQRCodesResponse represents a paginated response of QR codes
type GetQRCodesResponse struct {
	Total    int              `json:"total"`
	PageNum  int              `json:"pageNum"`
	PageSize int              `json:"pageSize"`
	Data     []QRCodeResponse `json:"data"`
}

// ClicksByDay represents daily click statistics
type ClicksByDay struct {
	Date   string `json:"date"`
	Total  int    `json:"total"`
	Unique int    `json:"unique"`
}

// ClicksStats represents click statistics
type ClicksStats struct {
	Total  int           `json:"total"`
	Unique int           `json:"unique"`
	ByDay  []ClicksByDay `json:"byDay"`
}

// GetCodeStatsResponse represents code statistics response
type GetCodeStatsResponse struct {
	Clicks    ClicksStats `json:"clicks"`
	Referrals []any       `json:"referrals"`
	Browsers  []any       `json:"browsers"`
	Devices   []any       `json:"devices"`
	Locations []any       `json:"locations"`
}

// Options represents configuration options for the Link client
type Options struct {
	URIs           []string
	OrganizationID string
	APIKey         string
}

// Option is a functional option for configuring the Link client
type Option func(*Options)

// WithAPIKey sets the API key for authentication
func WithAPIKey(key string) Option {
	return func(o *Options) {
		o.APIKey = key
	}
}

// WithOrganizationID sets the organization ID
func WithOrganizationID(id string) Option {
	return func(o *Options) {
		o.OrganizationID = id
	}
}

// WithURIs sets the base URIs for the Link service
func WithURIs(uris []string) Option {
	return func(o *Options) {
		o.URIs = uris
	}
}

// Link is the client for URL shortening services
type Link struct {
	uris           []string
	organizationID string
	apiKey         string
	client         client.HTTPClient
	errorHandler   func(error)
}

var defaultLinkURIs = []string{
	"https://api.hyphen.ai/api/organizations/{organizationId}/link/codes/",
}

// New creates a new Link client with functional options
func New(options ...Option) (*Link, error) {
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	// Get API key from options or environment
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("HYPHEN_API_KEY")
	}

	if apiKey != "" && strings.HasPrefix(apiKey, "public_") {
		return nil, fmt.Errorf("API key cannot start with \"public_\"")
	}

	// Get organization ID from options or environment
	organizationID := opts.OrganizationID
	if organizationID == "" {
		organizationID = os.Getenv("HYPHEN_ORGANIZATION_ID")
	}

	// Set URIs
	uris := opts.URIs
	if len(uris) == 0 {
		uris = defaultLinkURIs
	}

	l := &Link{
		uris:           uris,
		organizationID: organizationID,
		apiKey:         apiKey,
		client:         client.NewClient(""),
	}

	return l, nil
}

// SetErrorHandler sets a custom error handler function
func (l *Link) SetErrorHandler(handler func(error)) {
	l.errorHandler = handler
}

// emitError calls the error handler if set
func (l *Link) emitError(err error) {
	if l.errorHandler != nil {
		l.errorHandler(err)
	}
}

// getURI constructs the URI for a specific request
func (l *Link) getURI(prefix1, prefix2, prefix3 string) (string, error) {
	if l.organizationID == "" {
		return "", fmt.Errorf("organization ID is required")
	}

	uri := strings.Replace(l.uris[0], "{organizationId}", l.organizationID, 1)

	if prefix1 != "" {
		if strings.HasSuffix(uri, "/") {
			uri = uri + prefix1 + "/"
		} else {
			uri = uri + "/" + prefix1
		}
	}

	if prefix2 != "" {
		if strings.HasSuffix(uri, "/") {
			uri = uri + prefix2 + "/"
		} else {
			uri = uri + "/" + prefix2
		}
	}

	if prefix3 != "" {
		if strings.HasSuffix(uri, "/") {
			uri = uri + prefix3 + "/"
		} else {
			uri = uri + "/" + prefix3
		}
	}

	return strings.TrimSuffix(uri, "/"), nil
}

// CreateShortCode creates a short code for a long URL
func (l *Link) CreateShortCode(ctx context.Context, longURL, domain string, opts *CreateShortCodeOptions) (*ShortCodeResponse, error) {
	uri, err := l.getURI("", "", "")
	if err != nil {
		l.emitError(err)
		return nil, err
	}

	body := map[string]interface{}{
		"long_url": longURL,
		"domain":   domain,
	}

	if opts != nil {
		if opts.Code != "" {
			body["code"] = opts.Code
		}
		if opts.Title != "" {
			body["title"] = opts.Title
		}
		if len(opts.Tags) > 0 {
			body["tags"] = opts.Tags
		}
	}

	headers := client.CreateHeaders(l.apiKey)
	resp, err := l.client.Post(ctx, uri, body, headers)
	if err != nil {
		err = fmt.Errorf("failed to create short code: %w", err)
		l.emitError(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		err = fmt.Errorf("failed to create short code: HTTP %d: %s", resp.StatusCode, resp.Status)
		l.emitError(err)
		return nil, err
	}

	var shortCode ShortCodeResponse
	if err := json.Unmarshal(resp.Body, &shortCode); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		l.emitError(err)
		return nil, err
	}

	return &shortCode, nil
}

// GetShortCode retrieves a short code by its code
func (l *Link) GetShortCode(ctx context.Context, code string) (*ShortCodeResponse, error) {
	uri, err := l.getURI(code, "", "")
	if err != nil {
		l.emitError(err)
		return nil, err
	}

	headers := client.CreateHeaders(l.apiKey)
	resp, err := l.client.Get(ctx, uri, headers)
	if err != nil {
		err = fmt.Errorf("failed to get short code: %w", err)
		l.emitError(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get short code: HTTP %d: %s", resp.StatusCode, resp.Status)
		l.emitError(err)
		return nil, err
	}

	var shortCode ShortCodeResponse
	if err := json.Unmarshal(resp.Body, &shortCode); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		l.emitError(err)
		return nil, err
	}

	return &shortCode, nil
}

// GetShortCodes retrieves all short codes for the organization
func (l *Link) GetShortCodes(ctx context.Context, titleSearch string, tags []string, pageNumber, pageSize int) (*GetShortCodesResponse, error) {
	uri, err := l.getURI("", "", "")
	if err != nil {
		l.emitError(err)
		return nil, err
	}

	// Build query parameters
	params := url.Values{}
	if titleSearch != "" {
		params.Add("title", titleSearch)
	}
	if len(tags) > 0 {
		params.Add("tags", strings.Join(tags, ","))
	}
	if pageNumber > 0 {
		params.Add("pageNum", fmt.Sprintf("%d", pageNumber))
	}
	if pageSize > 0 {
		params.Add("pageSize", fmt.Sprintf("%d", pageSize))
	}

	if len(params) > 0 {
		uri = uri + "?" + params.Encode()
	}

	headers := client.CreateHeaders(l.apiKey)
	resp, err := l.client.Get(ctx, uri, headers)
	if err != nil {
		err = fmt.Errorf("failed to get short codes: %w", err)
		l.emitError(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get short codes: HTTP %d: %s", resp.StatusCode, resp.Status)
		l.emitError(err)
		return nil, err
	}

	var response GetShortCodesResponse
	if err := json.Unmarshal(resp.Body, &response); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		l.emitError(err)
		return nil, err
	}

	return &response, nil
}

// GetTags retrieves all tags associated with the organization's short codes
func (l *Link) GetTags(ctx context.Context) ([]string, error) {
	uri, err := l.getURI("tags", "", "")
	if err != nil {
		l.emitError(err)
		return nil, err
	}

	headers := client.CreateHeaders(l.apiKey)
	resp, err := l.client.Get(ctx, uri, headers)
	if err != nil {
		err = fmt.Errorf("failed to get tags: %w", err)
		l.emitError(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get tags: HTTP %d: %s", resp.StatusCode, resp.Status)
		l.emitError(err)
		return nil, err
	}

	var tags []string
	if err := json.Unmarshal(resp.Body, &tags); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		l.emitError(err)
		return nil, err
	}

	return tags, nil
}

// GetCodeStats retrieves statistics for a specific short code
func (l *Link) GetCodeStats(ctx context.Context, code string, startDate, endDate time.Time) (*GetCodeStatsResponse, error) {
	uri, err := l.getURI(code, "stats", "")
	if err != nil {
		l.emitError(err)
		return nil, err
	}

	// Build query parameters
	params := url.Values{}
	params.Add("startDate", startDate.Format(time.RFC3339))
	params.Add("endDate", endDate.Format(time.RFC3339))
	uri = uri + "?" + params.Encode()

	headers := client.CreateHeaders(l.apiKey)
	resp, err := l.client.Get(ctx, uri, headers)
	if err != nil {
		err = fmt.Errorf("failed to get code stats: %w", err)
		l.emitError(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get code stats: HTTP %d: %s", resp.StatusCode, resp.Status)
		l.emitError(err)
		return nil, err
	}

	var stats GetCodeStatsResponse
	if err := json.Unmarshal(resp.Body, &stats); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		l.emitError(err)
		return nil, err
	}

	return &stats, nil
}

// UpdateShortCode updates a short code
func (l *Link) UpdateShortCode(ctx context.Context, code string, opts *UpdateShortCodeOptions) (*ShortCodeResponse, error) {
	uri, err := l.getURI(code, "", "")
	if err != nil {
		l.emitError(err)
		return nil, err
	}

	headers := client.CreateHeaders(l.apiKey)
	resp, err := l.client.Patch(ctx, uri, opts, headers)
	if err != nil {
		err = fmt.Errorf("failed to update short code: %w", err)
		l.emitError(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to update short code: HTTP %d: %s", resp.StatusCode, resp.Status)
		l.emitError(err)
		return nil, err
	}

	var shortCode ShortCodeResponse
	if err := json.Unmarshal(resp.Body, &shortCode); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		l.emitError(err)
		return nil, err
	}

	return &shortCode, nil
}

// DeleteShortCode deletes a short code
func (l *Link) DeleteShortCode(ctx context.Context, code string) error {
	uri, err := l.getURI(code, "", "")
	if err != nil {
		l.emitError(err)
		return err
	}

	headers := client.CreateHeaders(l.apiKey)
	resp, err := l.client.Delete(ctx, uri, headers)
	if err != nil {
		err = fmt.Errorf("failed to delete short code: %w", err)
		l.emitError(err)
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		err = fmt.Errorf("failed to delete short code: HTTP %d: %s", resp.StatusCode, resp.Status)
		l.emitError(err)
		return err
	}

	return nil
}

// CreateQRCode creates a QR code for a specific short code
func (l *Link) CreateQRCode(ctx context.Context, code string, opts *CreateQRCodeOptions) (*QRCodeResponse, error) {
	uri, err := l.getURI(code, "qrs", "")
	if err != nil {
		l.emitError(err)
		return nil, err
	}

	headers := client.CreateHeaders(l.apiKey)
	resp, err := l.client.Post(ctx, uri, opts, headers)
	if err != nil {
		err = fmt.Errorf("failed to create QR code: %w", err)
		l.emitError(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		err = fmt.Errorf("failed to create QR code: HTTP %d: %s", resp.StatusCode, resp.Status)
		l.emitError(err)
		return nil, err
	}

	var qrCode QRCodeResponse
	if err := json.Unmarshal(resp.Body, &qrCode); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		l.emitError(err)
		return nil, err
	}

	return &qrCode, nil
}

// GetQRCode retrieves a QR code by its ID
func (l *Link) GetQRCode(ctx context.Context, code, qrID string) (*QRCodeResponse, error) {
	uri, err := l.getURI(code, "qrs", qrID)
	if err != nil {
		l.emitError(err)
		return nil, err
	}

	headers := client.CreateHeaders(l.apiKey)
	resp, err := l.client.Get(ctx, uri, headers)
	if err != nil {
		err = fmt.Errorf("failed to get QR code: %w", err)
		l.emitError(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get QR code: HTTP %d: %s", resp.StatusCode, resp.Status)
		l.emitError(err)
		return nil, err
	}

	var qrCode QRCodeResponse
	if err := json.Unmarshal(resp.Body, &qrCode); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		l.emitError(err)
		return nil, err
	}

	return &qrCode, nil
}

// GetQRCodes retrieves all QR codes for a short code
func (l *Link) GetQRCodes(ctx context.Context, code string, pageNumber, pageSize int) (*GetQRCodesResponse, error) {
	uri, err := l.getURI(code, "qrs", "")
	if err != nil {
		l.emitError(err)
		return nil, err
	}

	// Build query parameters
	params := url.Values{}
	if pageNumber > 0 {
		params.Add("pageNum", fmt.Sprintf("%d", pageNumber))
	}
	if pageSize > 0 {
		params.Add("pageSize", fmt.Sprintf("%d", pageSize))
	}

	if len(params) > 0 {
		uri = uri + "?" + params.Encode()
	}

	headers := client.CreateHeaders(l.apiKey)
	resp, err := l.client.Get(ctx, uri, headers)
	if err != nil {
		err = fmt.Errorf("failed to get QR codes: %w", err)
		l.emitError(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get QR codes: HTTP %d: %s", resp.StatusCode, resp.Status)
		l.emitError(err)
		return nil, err
	}

	var response GetQRCodesResponse
	if err := json.Unmarshal(resp.Body, &response); err != nil {
		err = fmt.Errorf("failed to unmarshal response: %w", err)
		l.emitError(err)
		return nil, err
	}

	return &response, nil
}

// DeleteQRCode deletes a QR code by its ID
func (l *Link) DeleteQRCode(ctx context.Context, code, qrID string) error {
	uri, err := l.getURI(code, "qrs", qrID)
	if err != nil {
		l.emitError(err)
		return err
	}

	headers := client.CreateHeaders(l.apiKey)
	resp, err := l.client.Delete(ctx, uri, headers)
	if err != nil {
		err = fmt.Errorf("failed to delete QR code: %w", err)
		l.emitError(err)
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		err = fmt.Errorf("failed to delete QR code: HTTP %d: %s", resp.StatusCode, resp.Status)
		l.emitError(err)
		return err
	}

	return nil
}
