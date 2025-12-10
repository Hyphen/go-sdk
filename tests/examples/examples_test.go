//go:build examples

package examples

import (
	"os"
	"testing"

	"github.com/Hyphen/go-sdk/tests/examples/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requireEnvVars fails the test if any of the specified environment variables are not set
func requireEnvVars(t *testing.T, vars ...string) {
	t.Helper()
	for _, v := range vars {
		if os.Getenv(v) == "" {
			t.Fatalf("Required environment variable %s is not set", v)
		}
	}
}

func TestToggleExample(t *testing.T) {
	requireEnvVars(t, "HYPHEN_PUBLIC_API_KEY", "HYPHEN_APPLICATION_ID")

	code, err := os.ReadFile("../../examples/toggle/main.go")
	require.NoError(t, err)

	transformed := testutil.ReplaceCredentials(string(code))
	transformed = testutil.InjectDevURIs(transformed)
	transformed = testutil.AddOsImport(transformed)

	output, err := testutil.RunGoCode(t, transformed)
	require.NoError(t, err, "Toggle example failed with output: %s", output)

	assert.Contains(t, output, "Feature enabled:")
	assert.Contains(t, output, "Welcome message:")
	assert.Contains(t, output, "Max retries:")
	assert.Contains(t, output, "App config:")
	assert.Contains(t, output, "Feature enabled for user-123:")
}

func TestNetInfoExample(t *testing.T) {
	requireEnvVars(t, "HYPHEN_API_KEY")

	code, err := os.ReadFile("../../examples/netinfo/main.go")
	require.NoError(t, err)

	transformed := testutil.ReplaceCredentials(string(code))
	transformed = testutil.InjectDevURIs(transformed)
	transformed = testutil.AddOsImport(transformed)

	output, err := testutil.RunGoCode(t, transformed)
	require.NoError(t, err, "NetInfo example failed with output: %s", output)

	assert.Contains(t, output, "Getting info for single IP address...")
	assert.Contains(t, output, "IP:")
	assert.Contains(t, output, "Country:")
	assert.Contains(t, output, "Getting info for multiple IP addresses...")
}

func TestLinkExample(t *testing.T) {
	requireEnvVars(t, "HYPHEN_API_KEY", "HYPHEN_ORGANIZATION_ID", "HYPHEN_LINK_DOMAIN")

	code, err := os.ReadFile("../../examples/link/main.go")
	require.NoError(t, err)

	transformed := testutil.ReplaceCredentials(string(code))
	transformed = testutil.InjectDevURIs(transformed)
	transformed = testutil.AddOsImport(transformed)

	output, err := testutil.RunGoCode(t, transformed)
	require.NoError(t, err, "Link example failed with output: %s", output)

	// Verify the example output contains expected content
	assert.Contains(t, output, "Creating a short code...")
	assert.Contains(t, output, "Created short code:")
	assert.Contains(t, output, "Short URL:")
	assert.Contains(t, output, "Getting short code details...")
	assert.Contains(t, output, "Title:")
	assert.Contains(t, output, "Long URL:")
	assert.Contains(t, output, "Updating short code...")
	assert.Contains(t, output, "Updated title:")
	assert.Contains(t, output, "Creating QR code...")
	assert.Contains(t, output, "QR Code ID:")
	assert.Contains(t, output, "QR Link:")
	assert.Contains(t, output, "Getting all tags...")
	assert.Contains(t, output, "Tags:")
	assert.Contains(t, output, "Getting code statistics...")
	assert.Contains(t, output, "Cleaning up...")
	assert.Contains(t, output, "Done!")
}

func TestCompleteExample(t *testing.T) {
	requireEnvVars(t, "HYPHEN_PUBLIC_API_KEY", "HYPHEN_APPLICATION_ID", "HYPHEN_API_KEY", "HYPHEN_ORGANIZATION_ID")

	code, err := os.ReadFile("../../examples/complete/main.go")
	require.NoError(t, err)

	transformed := testutil.ReplaceCredentials(string(code))
	transformed = testutil.InjectDevURIs(transformed)
	transformed = testutil.AddOsImport(transformed)

	output, err := testutil.RunGoCode(t, transformed)
	require.NoError(t, err, "Complete example failed with output: %s", output)

	assert.Contains(t, output, "=== Toggle Service ===")
	assert.Contains(t, output, "Feature enabled:")
	assert.Contains(t, output, "=== NetInfo Service ===")
	assert.Contains(t, output, "IP:")
	assert.Contains(t, output, "Country:")
	assert.Contains(t, output, "=== Link Service ===")
	assert.Contains(t, output, "All services accessed successfully!")
}
