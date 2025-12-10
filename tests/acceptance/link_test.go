//go:build acceptance

package acceptance

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Hyphen/go-sdk/pkg/link"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newLinkClient(t *testing.T) *link.Link {
	t.Helper()

	options := []link.Option{
		link.WithAPIKey(os.Getenv("HYPHEN_API_KEY")),
		link.WithOrganizationID(os.Getenv("HYPHEN_ORGANIZATION_ID")),
	}

	if strings.ToLower(os.Getenv("HYPHEN_DEV")) == "true" {
		options = append(options, link.WithURIs([]string{
			"https://dev-api.hyphen.ai/api/organizations/{organizationId}/link/codes/",
		}))
	}

	client, err := link.New(options...)
	require.NoError(t, err)
	return client
}

func TestLinkAcceptance(t *testing.T) {
	if os.Getenv("HYPHEN_API_KEY") == "" {
		t.Fatal("HYPHEN_API_KEY not set")
	}
	if os.Getenv("HYPHEN_ORGANIZATION_ID") == "" {
		t.Fatal("HYPHEN_ORGANIZATION_ID not set")
	}
	if os.Getenv("HYPHEN_LINK_DOMAIN") == "" {
		t.Fatal("HYPHEN_LINK_DOMAIN not set")
	}

	ctx := context.Background()
	domain := os.Getenv("HYPHEN_LINK_DOMAIN")

	t.Run("CreateShortCode_creates_a_short_code", func(t *testing.T) {
		client := newLinkClient(t)
		title := fmt.Sprintf("test-create-%d", time.Now().UnixNano())

		shortCode, err := client.CreateShortCode(ctx, "https://hyphen.ai", domain, &link.CreateShortCodeOptions{
			Title: title,
			Tags:  []string{"acceptance-test"},
		})

		require.NoError(t, err)
		assert.NotEmpty(t, shortCode.ID)
		assert.NotEmpty(t, shortCode.Code)
		assert.Equal(t, "https://hyphen.ai", shortCode.LongURL)
		assert.Equal(t, title, shortCode.Title)

		// Cleanup
		t.Cleanup(func() {
			client.DeleteShortCode(ctx, shortCode.ID)
		})
	})

	t.Run("GetShortCode_retrieves_an_existing_short_code", func(t *testing.T) {
		client := newLinkClient(t)
		title := fmt.Sprintf("test-get-%d", time.Now().UnixNano())

		created, err := client.CreateShortCode(ctx, "https://hyphen.ai", domain, &link.CreateShortCodeOptions{
			Title: title,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			client.DeleteShortCode(ctx, created.ID)
		})

		retrieved, err := client.GetShortCode(ctx, created.ID)

		require.NoError(t, err)
		assert.Equal(t, created.ID, retrieved.ID)
		assert.Equal(t, created.Code, retrieved.Code)
		assert.Equal(t, title, retrieved.Title)
	})

	t.Run("UpdateShortCode_updates_an_existing_short_code", func(t *testing.T) {
		client := newLinkClient(t)
		originalTitle := fmt.Sprintf("test-update-original-%d", time.Now().UnixNano())
		updatedTitle := fmt.Sprintf("test-update-updated-%d", time.Now().UnixNano())

		created, err := client.CreateShortCode(ctx, "https://hyphen.ai", domain, &link.CreateShortCodeOptions{
			Title: originalTitle,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			client.DeleteShortCode(ctx, created.ID)
		})

		updated, err := client.UpdateShortCode(ctx, created.ID, &link.UpdateShortCodeOptions{
			Title: updatedTitle,
		})

		require.NoError(t, err)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, updatedTitle, updated.Title)
	})

	t.Run("DeleteShortCode_deletes_an_existing_short_code", func(t *testing.T) {
		client := newLinkClient(t)
		title := fmt.Sprintf("test-delete-%d", time.Now().UnixNano())

		created, err := client.CreateShortCode(ctx, "https://hyphen.ai", domain, &link.CreateShortCodeOptions{
			Title: title,
		})
		require.NoError(t, err)

		// Wait for eventual consistency
		time.Sleep(2 * time.Second)

		err = client.DeleteShortCode(ctx, created.ID)

		require.NoError(t, err)

		// Verify it's deleted by trying to get it
		_, err = client.GetShortCode(ctx, created.ID)
		assert.Error(t, err)
	})

	t.Run("CreateQRCode_creates_a_qr_code_for_short_code", func(t *testing.T) {
		client := newLinkClient(t)
		title := fmt.Sprintf("test-qr-%d", time.Now().UnixNano())

		shortCode, err := client.CreateShortCode(ctx, "https://hyphen.ai", domain, &link.CreateShortCodeOptions{
			Title: title,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			client.DeleteShortCode(ctx, shortCode.ID)
		})

		qrCode, err := client.CreateQRCode(ctx, shortCode.ID, &link.CreateQRCodeOptions{
			Title: "Test QR Code",
			Size:  link.QRSizeMedium,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, qrCode.ID)
		assert.NotEmpty(t, qrCode.QRLink)

		// Cleanup QR code
		t.Cleanup(func() {
			client.DeleteQRCode(ctx, shortCode.ID, qrCode.ID)
		})
	})

	t.Run("GetQRCodes_retrieves_qr_codes_for_short_code", func(t *testing.T) {
		client := newLinkClient(t)
		title := fmt.Sprintf("test-get-qr-%d", time.Now().UnixNano())

		shortCode, err := client.CreateShortCode(ctx, "https://hyphen.ai", domain, &link.CreateShortCodeOptions{
			Title: title,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			client.DeleteShortCode(ctx, shortCode.ID)
		})

		qrCode, err := client.CreateQRCode(ctx, shortCode.ID, &link.CreateQRCodeOptions{
			Title: "Test QR Code",
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			client.DeleteQRCode(ctx, shortCode.ID, qrCode.ID)
		})

		qrCodes, err := client.GetQRCodes(ctx, shortCode.ID, 1, 10)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, qrCodes.Total, 1)
		assert.NotEmpty(t, qrCodes.Data)
	})

	t.Run("DeleteQRCode_deletes_a_qr_code", func(t *testing.T) {
		client := newLinkClient(t)
		title := fmt.Sprintf("test-delete-qr-%d", time.Now().UnixNano())

		shortCode, err := client.CreateShortCode(ctx, "https://hyphen.ai", domain, &link.CreateShortCodeOptions{
			Title: title,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			time.Sleep(2 * time.Second)
			client.DeleteShortCode(ctx, shortCode.ID)
		})

		qrCode, err := client.CreateQRCode(ctx, shortCode.ID, &link.CreateQRCodeOptions{
			Title: "Test QR Code to Delete",
		})
		require.NoError(t, err)

		// Wait for eventual consistency
		time.Sleep(2 * time.Second)

		err = client.DeleteQRCode(ctx, shortCode.ID, qrCode.ID)

		require.NoError(t, err)
	})

	t.Run("GetTags_retrieves_all_tags", func(t *testing.T) {
		client := newLinkClient(t)

		tags, err := client.GetTags(ctx)

		require.NoError(t, err)
		assert.NotNil(t, tags)
	})

	t.Run("GetCodeStats_retrieves_statistics_for_short_code", func(t *testing.T) {
		client := newLinkClient(t)
		title := fmt.Sprintf("test-stats-%d", time.Now().UnixNano())

		shortCode, err := client.CreateShortCode(ctx, "https://hyphen.ai", domain, &link.CreateShortCodeOptions{
			Title: title,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			client.DeleteShortCode(ctx, shortCode.ID)
		})

		endDate := time.Now()
		startDate := endDate.AddDate(0, -1, 0)

		stats, err := client.GetCodeStats(ctx, shortCode.ID, startDate, endDate)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, stats.Clicks.Total, 0)
	})
}
