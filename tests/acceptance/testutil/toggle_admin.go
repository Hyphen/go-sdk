package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// ToggleAdmin provides methods to create and delete toggles via the Hyphen Management API.
// This is used for test setup and teardown in acceptance tests.
type ToggleAdmin struct {
	apiKey         string
	organizationID string
	projectID      string
	baseURL        string
	client         *http.Client
}

// NewToggleAdmin creates a new ToggleAdmin from environment variables.
// Required environment variables:
//   - HYPHEN_API_KEY: API key with management permissions
//   - HYPHEN_ORGANIZATION_ID: Organization ID (e.g., org_...)
//   - HYPHEN_PROJECT_ID: Project ID
//
// Optional environment variables:
//   - HYPHEN_DEV: Set to "true" to use dev-api.hyphen.ai
func NewToggleAdmin() *ToggleAdmin {
	baseURL := "https://api.hyphen.ai"
	if os.Getenv("HYPHEN_DEV") == "true" {
		baseURL = "https://dev-api.hyphen.ai"
	}

	return &ToggleAdmin{
		apiKey:         os.Getenv("HYPHEN_API_KEY"),
		organizationID: os.Getenv("HYPHEN_ORGANIZATION_ID"),
		projectID:      os.Getenv("HYPHEN_PROJECT_ID"),
		baseURL:        baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsConfigured returns true if all required environment variables are set.
func (a *ToggleAdmin) IsConfigured() bool {
	return a.apiKey != "" && a.organizationID != "" && a.projectID != ""
}

// Target represents a targeting rule for a toggle.
type Target struct {
	Logic string      `json:"logic"` // JSONLogic expression
	Value interface{} `json:"value"` // Value to return if logic evaluates to true
}

type createToggleRequest struct {
	Key          string      `json:"key"`
	Type         string      `json:"type"`
	Targets      []Target    `json:"targets"`
	DefaultValue interface{} `json:"defaultValue"`
	Description  string      `json:"description,omitempty"`
}

// CreateBooleanToggle creates a boolean toggle with the given key and default value.
func (a *ToggleAdmin) CreateBooleanToggle(ctx context.Context, key string, defaultValue bool) error {
	return a.createToggle(ctx, key, "boolean", defaultValue, nil)
}

// CreateStringToggle creates a string toggle with the given key and default value.
func (a *ToggleAdmin) CreateStringToggle(ctx context.Context, key string, defaultValue string) error {
	return a.createToggle(ctx, key, "string", defaultValue, nil)
}

// CreateNumberToggle creates a number toggle with the given key and default value.
func (a *ToggleAdmin) CreateNumberToggle(ctx context.Context, key string, defaultValue float64) error {
	return a.createToggle(ctx, key, "number", defaultValue, nil)
}

// CreateBooleanToggleWithTargets creates a boolean toggle with targeting rules.
func (a *ToggleAdmin) CreateBooleanToggleWithTargets(ctx context.Context, key string, defaultValue bool, targets []Target) error {
	return a.createToggle(ctx, key, "boolean", defaultValue, targets)
}

// CreateStringToggleWithTargets creates a string toggle with targeting rules.
func (a *ToggleAdmin) CreateStringToggleWithTargets(ctx context.Context, key string, defaultValue string, targets []Target) error {
	return a.createToggle(ctx, key, "string", defaultValue, targets)
}

func (a *ToggleAdmin) createToggle(ctx context.Context, key, toggleType string, defaultValue interface{}, targets []Target) error {
	url := fmt.Sprintf("%s/api/organizations/%s/projects/%s/toggles/",
		a.baseURL, a.organizationID, a.projectID)

	if targets == nil {
		targets = []Target{}
	}

	reqBody := createToggleRequest{
		Key:          key,
		Type:         toggleType,
		Targets:      targets,
		DefaultValue: defaultValue,
		Description:  "Created by acceptance test",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Allow time for eventual consistency
	time.Sleep(500 * time.Millisecond)

	return nil
}

// DeleteToggle deletes a toggle by its key.
func (a *ToggleAdmin) DeleteToggle(ctx context.Context, key string) error {
	url := fmt.Sprintf("%s/api/organizations/%s/projects/%s/toggles/%s",
		a.baseURL, a.organizationID, a.projectID, key)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
