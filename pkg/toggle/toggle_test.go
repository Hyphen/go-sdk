package toggle

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("creates_a_new_toggle_client_with_the_provided_options", func(t *testing.T) {
		thePublicAPIKey := "public_dGVzdC1vcmc6c2VjcmV0"
		theApplicationID := "theApplicationID"
		theEnvironment := "theEnvironment"

		result, err := New(
			WithPublicAPIKey(thePublicAPIKey),
			WithApplicationID(theApplicationID),
			WithEnvironment(theEnvironment),
		)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected result to not be nil")
		}
		if result.publicAPIKey != thePublicAPIKey {
			t.Errorf("Expected publicAPIKey to be %s, got %s", thePublicAPIKey, result.publicAPIKey)
		}
		if result.applicationID != theApplicationID {
			t.Errorf("Expected applicationID to be %s, got %s", theApplicationID, result.applicationID)
		}
		if result.environment != theEnvironment {
			t.Errorf("Expected environment to be %s, got %s", theEnvironment, result.environment)
		}
	})

	t.Run("uses_development_as_default_environment_when_not_provided", func(t *testing.T) {
		result, err := New(
			WithPublicAPIKey("public_dGVzdC1vcmc6c2VjcmV0"),
			WithApplicationID("anApplicationID"),
		)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result.environment != "development" {
			t.Errorf("Expected environment to be development, got %s", result.environment)
		}
	})

	t.Run("extracts_organization_id_from_public_key", func(t *testing.T) {
		thePublicAPIKey := "public_dGVzdC1vcmc6c2VjcmV0"
		theExpectedOrgID := "test-org"

		result, err := New(
			WithPublicAPIKey(thePublicAPIKey),
			WithApplicationID("anApplicationID"),
		)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result.organizationID != theExpectedOrgID {
			t.Errorf("Expected organizationID to be %s, got %s", theExpectedOrgID, result.organizationID)
		}
	})
}

func TestGetBoolean(t *testing.T) {
	t.Run("returns_the_expected_boolean_value_when_successful", func(t *testing.T) {
		theToggleKey := "theToggleKey"
		theExpectedValue := true
		theApplicationID := "theApplicationID"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := EvaluationResponse{
				Toggles: map[string]Evaluation{
					theToggleKey: {
						Key:   theToggleKey,
						Value: theExpectedValue,
						Type:  "boolean",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		t.Cleanup(func() { server.Close() })

		toggle, err := New(
			WithPublicAPIKey("public_dGVzdC1vcmc6c2VjcmV0"),
			WithApplicationID(theApplicationID),
			WithHorizonURLs([]string{server.URL}),
		)
		if err != nil {
			t.Fatalf("Failed to create toggle client: %v", err)
		}

		result := toggle.GetBoolean(context.Background(), theToggleKey, false, nil)

		if result != theExpectedValue {
			t.Errorf("Expected %v, got %v", theExpectedValue, result)
		}
	})

	t.Run("returns_the_default_value_when_the_request_fails", func(t *testing.T) {
		theDefaultValue := false

		toggle, err := New(
			WithPublicAPIKey("public_dGVzdC1vcmc6c2VjcmV0"),
			WithApplicationID("anApplicationID"),
			WithHorizonURLs([]string{"http://invalid-url-that-does-not-exist.local"}),
		)
		if err != nil {
			t.Fatalf("Failed to create toggle client: %v", err)
		}

		result := toggle.GetBoolean(context.Background(), "aToggleKey", theDefaultValue, nil)

		if result != theDefaultValue {
			t.Errorf("Expected %v, got %v", theDefaultValue, result)
		}
	})
}

func TestGetString(t *testing.T) {
	t.Run("returns_the_expected_string_value_when_successful", func(t *testing.T) {
		theToggleKey := "theToggleKey"
		theExpectedValue := "theValue"
		theApplicationID := "theApplicationID"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := EvaluationResponse{
				Toggles: map[string]Evaluation{
					theToggleKey: {
						Key:   theToggleKey,
						Value: theExpectedValue,
						Type:  "string",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		t.Cleanup(func() { server.Close() })

		toggle, err := New(
			WithPublicAPIKey("public_dGVzdC1vcmc6c2VjcmV0"),
			WithApplicationID(theApplicationID),
			WithHorizonURLs([]string{server.URL}),
		)
		if err != nil {
			t.Fatalf("Failed to create toggle client: %v", err)
		}

		result := toggle.GetString(context.Background(), theToggleKey, "aDefaultValue", nil)

		if result != theExpectedValue {
			t.Errorf("Expected %s, got %s", theExpectedValue, result)
		}
	})
}

func TestGetOrgIDFromPublicKey(t *testing.T) {
	t.Run("returns_the_organization_id_from_a_valid_public_key", func(t *testing.T) {
		thePublicKey := "public_dGVzdC1vcmc6c2VjcmV0"
		theExpectedOrgID := "test-org"

		result := getOrgIDFromPublicKey(thePublicKey)

		if result != theExpectedOrgID {
			t.Errorf("Expected %s, got %s", theExpectedOrgID, result)
		}
	})

	t.Run("returns_empty_string_for_invalid_public_key", func(t *testing.T) {
		thePublicKey := "invalid_key"

		result := getOrgIDFromPublicKey(thePublicKey)

		if result != "" {
			t.Errorf("Expected empty string, got %s", result)
		}
	})
}

func TestGenerateTargetKey(t *testing.T) {
	t.Run("generates_a_target_key_with_app_and_environment", func(t *testing.T) {
		theApplicationID := "theApp"
		theEnvironment := "theEnv"

		result := generateTargetKey(theApplicationID, theEnvironment)

		if result == "" {
			t.Error("Expected non-empty target key")
		}
	})
}
