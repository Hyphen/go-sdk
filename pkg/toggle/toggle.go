package toggle

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/Hyphen/go-sdk/internal/client"
)

// CustomAttributes represents custom key-value pairs for evaluation context
type CustomAttributes map[string]interface{}

// User represents user information for evaluation context
type User struct {
	ID               string           `json:"id"`
	Email            string           `json:"email,omitempty"`
	Name             string           `json:"name,omitempty"`
	CustomAttributes CustomAttributes `json:"customAttributes,omitempty"`
}

// Context represents the evaluation context for feature toggle evaluation
type Context struct {
	TargetingKey     string           `json:"targetingKey,omitempty"`
	IPAddress        string           `json:"ipAddress,omitempty"`
	CustomAttributes CustomAttributes `json:"customAttributes,omitempty"`
	User             *User            `json:"user,omitempty"`
}

// Evaluation represents a single toggle evaluation result
type Evaluation struct {
	Key          string      `json:"key"`
	Value        interface{} `json:"value"`
	Type         string      `json:"type"`
	Reason       interface{} `json:"reason,omitempty"`
	ErrorMessage string      `json:"errorMessage,omitempty"`
}

// EvaluationResponse represents the response from the toggle evaluation API
type EvaluationResponse struct {
	Toggles map[string]Evaluation `json:"toggles"`
}

// toggleEvaluation is the internal type for API requests
type toggleEvaluation struct {
	Application      string           `json:"application"`
	Environment      string           `json:"environment"`
	TargetingKey     string           `json:"targetingKey,omitempty"`
	IPAddress        string           `json:"ipAddress,omitempty"`
	CustomAttributes CustomAttributes `json:"customAttributes,omitempty"`
	User             *User            `json:"user,omitempty"`
}

// Options represents configuration options for the Toggle client
type Options struct {
	PublicAPIKey        string
	ApplicationID       string
	Environment         string
	DefaultContext      *Context
	HorizonURLs         []string
	DefaultTargetingKey string
}

// Option is a functional option for configuring the Toggle client
type Option func(*Options)

// WithPublicAPIKey sets the public API key for authentication
func WithPublicAPIKey(key string) Option {
	return func(o *Options) {
		o.PublicAPIKey = key
	}
}

// WithApplicationID sets the application ID for toggle evaluation
func WithApplicationID(id string) Option {
	return func(o *Options) {
		o.ApplicationID = id
	}
}

// WithEnvironment sets the environment (e.g., "development", "production")
func WithEnvironment(env string) Option {
	return func(o *Options) {
		o.Environment = env
	}
}

// WithDefaultContext sets the default evaluation context
func WithDefaultContext(ctx *Context) Option {
	return func(o *Options) {
		o.DefaultContext = ctx
	}
}

// WithHorizonURLs sets custom Horizon API URLs
func WithHorizonURLs(urls []string) Option {
	return func(o *Options) {
		o.HorizonURLs = urls
	}
}

// WithDefaultTargetingKey sets the default targeting key
func WithDefaultTargetingKey(key string) Option {
	return func(o *Options) {
		o.DefaultTargetingKey = key
	}
}

// Toggle is the client for feature flag management
type Toggle struct {
	publicAPIKey        string
	organizationID      string
	applicationID       string
	environment         string
	horizonURLs         []string
	defaultContext      *Context
	defaultTargetingKey string
	client              *client.Client
	errorHandler        func(error)
}

// New creates a new Toggle client with functional options
func New(options ...Option) (*Toggle, error) {
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	// Get public API key from options or environment
	publicAPIKey := opts.PublicAPIKey
	if publicAPIKey == "" {
		publicAPIKey = os.Getenv("HYPHEN_PUBLIC_API_KEY")
	}

	// Get application ID from options or environment
	applicationID := opts.ApplicationID
	if applicationID == "" {
		applicationID = os.Getenv("HYPHEN_APPLICATION_ID")
	}

	// Set environment, default to "development"
	environment := opts.Environment
	if environment == "" {
		environment = "development"
	}

	// Extract organization ID from public key
	var organizationID string
	if publicAPIKey != "" {
		organizationID = getOrgIDFromPublicKey(publicAPIKey)
	}

	// Set horizon URLs
	horizonURLs := opts.HorizonURLs
	if len(horizonURLs) == 0 {
		horizonURLs = getDefaultHorizonURLs(publicAPIKey)
	}

	// Set default targeting key
	defaultTargetingKey := opts.DefaultTargetingKey
	if defaultTargetingKey == "" {
		if opts.DefaultContext != nil {
			defaultTargetingKey = getTargetingKey(opts.DefaultContext, applicationID, environment)
		} else {
			defaultTargetingKey = generateTargetKey(applicationID, environment)
		}
	}

	t := &Toggle{
		publicAPIKey:        publicAPIKey,
		organizationID:      organizationID,
		applicationID:       applicationID,
		environment:         environment,
		horizonURLs:         horizonURLs,
		defaultContext:      opts.DefaultContext,
		defaultTargetingKey: defaultTargetingKey,
		client:              client.NewClient(""),
	}

	return t, nil
}

// SetErrorHandler sets a custom error handler function
func (t *Toggle) SetErrorHandler(handler func(error)) {
	t.errorHandler = handler
}

// emitError calls the error handler if set
func (t *Toggle) emitError(err error) {
	if t.errorHandler != nil {
		t.errorHandler(err)
	}
}

// Get retrieves a toggle value with generic type support
func (t *Toggle) Get(ctx context.Context, toggleKey string, defaultValue interface{}, contextOverride *Context) (interface{}, error) {
	evalContext := t.buildEvaluationContext(contextOverride)

	headers := client.CreateHeaders(t.publicAPIKey)

	// Try each horizon URL in order
	var lastErr error
	for _, baseURL := range t.horizonURLs {
		url := fmt.Sprintf("%s/toggle/evaluate", strings.TrimSuffix(baseURL, "/"))

		resp, err := t.client.Post(ctx, url, evalContext, headers)
		if err != nil {
			lastErr = fmt.Errorf("request to %s failed: %w", baseURL, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			continue
		}

		var evalResp EvaluationResponse
		if err := json.Unmarshal(resp.Body, &evalResp); err != nil {
			lastErr = fmt.Errorf("failed to unmarshal response: %w", err)
			continue
		}

		if toggle, ok := evalResp.Toggles[toggleKey]; ok {
			return toggle.Value, nil
		}

		return defaultValue, nil
	}

	err := fmt.Errorf("all horizon URLs failed. Last error: %w", lastErr)
	t.emitError(err)
	return defaultValue, err
}

// GetBoolean retrieves a boolean toggle value
func (t *Toggle) GetBoolean(ctx context.Context, toggleKey string, defaultValue bool, contextOverride *Context) bool {
	val, err := t.Get(ctx, toggleKey, defaultValue, contextOverride)
	if err != nil {
		return defaultValue
	}

	if boolVal, ok := val.(bool); ok {
		return boolVal
	}

	return defaultValue
}

// GetString retrieves a string toggle value
func (t *Toggle) GetString(ctx context.Context, toggleKey string, defaultValue string, contextOverride *Context) string {
	val, err := t.Get(ctx, toggleKey, defaultValue, contextOverride)
	if err != nil {
		return defaultValue
	}

	if strVal, ok := val.(string); ok {
		return strVal
	}

	return defaultValue
}

// GetNumber retrieves a number toggle value (returns float64)
func (t *Toggle) GetNumber(ctx context.Context, toggleKey string, defaultValue float64, contextOverride *Context) float64 {
	val, err := t.Get(ctx, toggleKey, defaultValue, contextOverride)
	if err != nil {
		return defaultValue
	}

	if numVal, ok := val.(float64); ok {
		return numVal
	}

	return defaultValue
}

// GetObject retrieves an object toggle value
func (t *Toggle) GetObject(ctx context.Context, toggleKey string, defaultValue map[string]interface{}, contextOverride *Context) map[string]interface{} {
	val, err := t.Get(ctx, toggleKey, defaultValue, contextOverride)
	if err != nil {
		return defaultValue
	}

	if objVal, ok := val.(map[string]interface{}); ok {
		return objVal
	}

	return defaultValue
}

// buildEvaluationContext builds the evaluation context for API requests
func (t *Toggle) buildEvaluationContext(contextOverride *Context) *toggleEvaluation {
	eval := &toggleEvaluation{
		Application: t.applicationID,
		Environment: t.environment,
	}

	var ctx *Context
	if contextOverride != nil {
		ctx = contextOverride
	} else {
		ctx = t.defaultContext
	}

	if ctx != nil {
		eval.TargetingKey = ctx.TargetingKey
		eval.IPAddress = ctx.IPAddress
		eval.CustomAttributes = ctx.CustomAttributes
		eval.User = ctx.User
	}

	if eval.TargetingKey == "" {
		if ctx != nil {
			eval.TargetingKey = getTargetingKey(ctx, t.applicationID, t.environment)
		} else {
			eval.TargetingKey = t.defaultTargetingKey
		}
	}

	return eval
}

// getOrgIDFromPublicKey extracts the organization ID from a public API key
func getOrgIDFromPublicKey(publicKey string) string {
	if !strings.HasPrefix(publicKey, "public_") {
		return ""
	}

	keyWithoutPrefix := strings.TrimPrefix(publicKey, "public_")
	decoded, err := base64.StdEncoding.DecodeString(keyWithoutPrefix)
	if err != nil {
		return ""
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}

// getDefaultHorizonURL builds the default Horizon API URL
func getDefaultHorizonURL(publicKey string) string {
	if publicKey == "" {
		return "https://toggle.hyphen.cloud"
	}

	orgID := getOrgIDFromPublicKey(publicKey)
	if orgID == "" {
		return "https://toggle.hyphen.cloud"
	}

	return fmt.Sprintf("https://%s.toggle.hyphen.cloud", orgID)
}

// getDefaultHorizonURLs gets the default Horizon URLs for load balancing
func getDefaultHorizonURLs(publicKey string) []string {
	defaultURL := "https://toggle.hyphen.cloud"

	if publicKey == "" {
		return []string{defaultURL}
	}

	orgURL := getDefaultHorizonURL(publicKey)
	if orgURL == defaultURL {
		return []string{defaultURL}
	}

	return []string{orgURL, defaultURL}
}

// generateTargetKey generates a unique targeting key
func generateTargetKey(applicationID, environment string) string {
	randomSuffix := fmt.Sprintf("%d", rand.Int63())

	var components []string
	if applicationID != "" {
		components = append(components, applicationID)
	}
	if environment != "" {
		components = append(components, environment)
	}
	components = append(components, randomSuffix)

	return strings.Join(components, "-")
}

// getTargetingKey extracts targeting key from context with fallback logic
func getTargetingKey(ctx *Context, applicationID, environment string) string {
	if ctx.TargetingKey != "" {
		return ctx.TargetingKey
	}
	if ctx.User != nil && ctx.User.ID != "" {
		return ctx.User.ID
	}
	return generateTargetKey(applicationID, environment)
}
