//go:build acceptance

package acceptance

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/Hyphen/go-sdk/pkg/netinfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newNetInfoClient(t *testing.T) *netinfo.NetInfo {
	t.Helper()

	options := []netinfo.Option{
		netinfo.WithAPIKey(os.Getenv("HYPHEN_API_KEY")),
	}

	if strings.ToLower(os.Getenv("HYPHEN_DEV")) == "true" {
		options = append(options, netinfo.WithBaseURI("https://dev.net.info"))
	}

	client, err := netinfo.New(options...)
	require.NoError(t, err)
	return client
}

func TestNetInfoAcceptance(t *testing.T) {
	if os.Getenv("HYPHEN_API_KEY") == "" {
		t.Skip("HYPHEN_API_KEY not set")
	}

	ctx := context.Background()

	t.Run("GetIPInfo_returns_info_for_valid_ip", func(t *testing.T) {
		client := newNetInfoClient(t)

		result, err := client.GetIPInfo(ctx, "8.8.8.8")

		require.NoError(t, err)
		assert.Equal(t, "8.8.8.8", result.IP)
		assert.NotEmpty(t, result.Location.Country)
	})

	t.Run("GetIPInfo_returns_info_for_cloudflare_dns", func(t *testing.T) {
		client := newNetInfoClient(t)

		result, err := client.GetIPInfo(ctx, "1.1.1.1")

		require.NoError(t, err)
		assert.Equal(t, "1.1.1.1", result.IP)
		assert.NotEmpty(t, result.Location.Country)
	})

	t.Run("GetIPInfos_returns_info_for_multiple_ips", func(t *testing.T) {
		client := newNetInfoClient(t)

		ips := []string{"8.8.8.8", "1.1.1.1"}
		results, err := client.GetIPInfos(ctx, ips)

		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("GetIPInfos_returns_error_for_empty_array", func(t *testing.T) {
		client := newNetInfoClient(t)

		results, err := client.GetIPInfos(ctx, []string{})

		assert.Error(t, err)
		assert.Nil(t, results)
	})
}
