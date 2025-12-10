//go:build acceptance

package acceptance

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Hyphen/go-sdk/pkg/toggle"
	"github.com/Hyphen/go-sdk/tests/acceptance/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newToggleClient(t *testing.T) *toggle.Toggle {
	t.Helper()

	options := []toggle.Option{
		toggle.WithPublicAPIKey(os.Getenv("HYPHEN_PUBLIC_API_KEY")),
		toggle.WithApplicationID(os.Getenv("HYPHEN_APPLICATION_ID")),
	}

	if strings.ToLower(os.Getenv("HYPHEN_DEV")) == "true" {
		options = append(options, toggle.WithHorizonURLs([]string{"https://dev-horizon.hyphen.ai"}))
	}

	client, err := toggle.New(options...)
	require.NoError(t, err)
	return client
}

func TestToggleAcceptance(t *testing.T) {
	admin := testutil.NewToggleAdmin()
	if !admin.IsConfigured() {
		t.Skip("Toggle admin not configured (missing HYPHEN_API_KEY, HYPHEN_ORGANIZATION_ID, or HYPHEN_PROJECT_ID)")
	}

	if os.Getenv("HYPHEN_PUBLIC_API_KEY") == "" {
		t.Skip("HYPHEN_PUBLIC_API_KEY not set")
	}

	ctx := context.Background()

	t.Run("GetBoolean_returns_true_when_toggle_default_is_true", func(t *testing.T) {
		toggleKey := fmt.Sprintf("test-bool-true-%d", time.Now().UnixNano())
		err := admin.CreateBooleanToggle(ctx, toggleKey, true)
		require.NoError(t, err)
		t.Cleanup(func() { admin.DeleteToggle(ctx, toggleKey) })

		client := newToggleClient(t)

		result := client.GetBoolean(ctx, toggleKey, false, nil)

		assert.True(t, result)
	})

	t.Run("GetBoolean_returns_false_when_toggle_default_is_false", func(t *testing.T) {
		toggleKey := fmt.Sprintf("test-bool-false-%d", time.Now().UnixNano())
		err := admin.CreateBooleanToggle(ctx, toggleKey, false)
		require.NoError(t, err)
		t.Cleanup(func() { admin.DeleteToggle(ctx, toggleKey) })

		client := newToggleClient(t)

		result := client.GetBoolean(ctx, toggleKey, true, nil)

		assert.False(t, result)
	})

	t.Run("GetBoolean_returns_default_value_for_nonexistent_toggle", func(t *testing.T) {
		client := newToggleClient(t)

		theDefaultValue := true

		result := client.GetBoolean(ctx, "nonexistent-toggle-key", theDefaultValue, nil)

		assert.Equal(t, theDefaultValue, result)
	})

	t.Run("GetString_returns_configured_value", func(t *testing.T) {
		toggleKey := fmt.Sprintf("test-string-%d", time.Now().UnixNano())
		theExpectedValue := "the-configured-string-value"
		err := admin.CreateStringToggle(ctx, toggleKey, theExpectedValue)
		require.NoError(t, err)
		t.Cleanup(func() { admin.DeleteToggle(ctx, toggleKey) })

		client := newToggleClient(t)

		result := client.GetString(ctx, toggleKey, "a-default-value", nil)

		assert.Equal(t, theExpectedValue, result)
	})

	t.Run("GetString_returns_default_value_for_nonexistent_toggle", func(t *testing.T) {
		client := newToggleClient(t)

		theDefaultValue := "the-default-string-value"

		result := client.GetString(ctx, "nonexistent-toggle-key", theDefaultValue, nil)

		assert.Equal(t, theDefaultValue, result)
	})

	t.Run("GetNumber_returns_configured_value", func(t *testing.T) {
		toggleKey := fmt.Sprintf("test-number-%d", time.Now().UnixNano())
		theExpectedValue := 42.5
		err := admin.CreateNumberToggle(ctx, toggleKey, theExpectedValue)
		require.NoError(t, err)
		t.Cleanup(func() { admin.DeleteToggle(ctx, toggleKey) })

		client := newToggleClient(t)

		result := client.GetNumber(ctx, toggleKey, 0.0, nil)

		assert.Equal(t, theExpectedValue, result)
	})

	t.Run("GetNumber_returns_default_value_for_nonexistent_toggle", func(t *testing.T) {
		client := newToggleClient(t)

		theDefaultValue := 99.9

		result := client.GetNumber(ctx, "nonexistent-toggle-key", theDefaultValue, nil)

		assert.Equal(t, theDefaultValue, result)
	})

	t.Run("GetBoolean_returns_targeted_value_when_user_id_matches", func(t *testing.T) {
		toggleKey := fmt.Sprintf("test-targeting-user-%d", time.Now().UnixNano())
		// JSONLogic: if user.id == "the-vip-user", return true
		targets := []testutil.Target{
			{
				Logic: `{"==": [{"var": "user.id"}, "the-vip-user"]}`,
				Value: true,
			},
		}
		err := admin.CreateBooleanToggleWithTargets(ctx, toggleKey, false, targets)
		require.NoError(t, err)
		t.Cleanup(func() { admin.DeleteToggle(ctx, toggleKey) })

		client := newToggleClient(t)

		// Act - with matching user ID
		resultWithMatch := client.GetBoolean(ctx, toggleKey, false, &toggle.Context{
			User: &toggle.User{
				ID: "the-vip-user",
			},
		})

		// Act - with non-matching user ID
		resultWithoutMatch := client.GetBoolean(ctx, toggleKey, false, &toggle.Context{
			User: &toggle.User{
				ID: "a-regular-user",
			},
		})

		// Assert
		assert.True(t, resultWithMatch, "should return targeted value for matching user")
		assert.False(t, resultWithoutMatch, "should return default value for non-matching user")
	})

	t.Run("GetString_returns_targeted_value_based_on_custom_attribute", func(t *testing.T) {
		toggleKey := fmt.Sprintf("test-targeting-attr-%d", time.Now().UnixNano())
		// JSONLogic: if customAttributes.plan == "premium", return "premium-feature"
		targets := []testutil.Target{
			{
				Logic: `{"==": [{"var": "customAttributes.plan"}, "premium"]}`,
				Value: "the-premium-feature-value",
			},
		}
		err := admin.CreateStringToggleWithTargets(ctx, toggleKey, "the-default-feature-value", targets)
		require.NoError(t, err)
		t.Cleanup(func() { admin.DeleteToggle(ctx, toggleKey) })

		client := newToggleClient(t)

		// Act - with matching custom attribute
		resultPremium := client.GetString(ctx, toggleKey, "a-fallback", &toggle.Context{
			CustomAttributes: toggle.CustomAttributes{
				"plan": "premium",
			},
		})

		// Act - with non-matching custom attribute
		resultFree := client.GetString(ctx, toggleKey, "a-fallback", &toggle.Context{
			CustomAttributes: toggle.CustomAttributes{
				"plan": "free",
			},
		})

		// Assert
		assert.Equal(t, "the-premium-feature-value", resultPremium, "should return targeted value for premium plan")
		assert.Equal(t, "the-default-feature-value", resultFree, "should return default value for free plan")
	})

	t.Run("GetBoolean_returns_targeted_value_based_on_targeting_key", func(t *testing.T) {
		toggleKey := fmt.Sprintf("test-targeting-key-%d", time.Now().UnixNano())
		// JSONLogic: if targetingKey == "the-beta-tester", return true
		targets := []testutil.Target{
			{
				Logic: `{"==": [{"var": "targetingKey"}, "the-beta-tester"]}`,
				Value: true,
			},
		}
		err := admin.CreateBooleanToggleWithTargets(ctx, toggleKey, false, targets)
		require.NoError(t, err)
		t.Cleanup(func() { admin.DeleteToggle(ctx, toggleKey) })

		client := newToggleClient(t)

		// Act - with matching targeting key
		resultBeta := client.GetBoolean(ctx, toggleKey, false, &toggle.Context{
			TargetingKey: "the-beta-tester",
		})

		// Act - with non-matching targeting key
		resultRegular := client.GetBoolean(ctx, toggleKey, false, &toggle.Context{
			TargetingKey: "a-regular-user",
		})

		// Assert
		assert.True(t, resultBeta, "should return targeted value for beta tester")
		assert.False(t, resultRegular, "should return default value for regular user")
	})
}
