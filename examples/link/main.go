package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Hyphen/hyphen-go-sdk"
)

func main() {
	// Create a Link client using functional options
	link, err := hyphen.NewLink(
		hyphen.WithAPIKey("your_api_key"),
		hyphen.WithOrganizationID("your_organization_id"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create a short code
	fmt.Println("Creating a short code...")
	shortCode, err := link.CreateShortCode(
		ctx,
		"https://hyphen.ai",
		"test.h4n.link",
		&hyphen.CreateShortCodeOptions{
			Title: "Hyphen Homepage",
			Tags:  []string{"example", "homepage"},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created short code: %s\n", shortCode.Code)
	fmt.Printf("Short URL: https://%s/%s\n", shortCode.Domain, shortCode.Code)

	// Get the short code
	fmt.Println("\nGetting short code details...")
	retrieved, err := link.GetShortCode(ctx, shortCode.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Title: %s\n", retrieved.Title)
	fmt.Printf("Long URL: %s\n", retrieved.LongURL)

	// Update the short code
	fmt.Println("\nUpdating short code...")
	updated, err := link.UpdateShortCode(ctx, shortCode.ID, &hyphen.UpdateShortCodeOptions{
		Title: "Updated Title",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated title: %s\n", updated.Title)

	// Create a QR code
	fmt.Println("\nCreating QR code...")
	qrCode, err := link.CreateQRCode(ctx, shortCode.ID, &hyphen.CreateQRCodeOptions{
		Title: "My QR Code",
		Size:  hyphen.QRSizeMedium,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("QR Code ID: %s\n", qrCode.ID)
	fmt.Printf("QR Link: %s\n", qrCode.QRLink)

	// Get all tags
	fmt.Println("\nGetting all tags...")
	tags, err := link.GetTags(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Tags: %v\n", tags)

	// Get code stats
	fmt.Println("\nGetting code statistics...")
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0) // 1 month ago
	stats, err := link.GetCodeStats(ctx, shortCode.ID, startDate, endDate)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total clicks: %d\n", stats.Clicks.Total)

	// Clean up - delete QR code and short code
	fmt.Println("\nCleaning up...")
	err = link.DeleteQRCode(ctx, shortCode.ID, qrCode.ID)
	if err != nil {
		log.Printf("Error deleting QR code: %v\n", err)
	}

	err = link.DeleteShortCode(ctx, shortCode.ID)
	if err != nil {
		log.Printf("Error deleting short code: %v\n", err)
	}

	fmt.Println("Done!")
}
