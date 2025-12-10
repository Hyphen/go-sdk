package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Hyphen/go-sdk"
)

func main() {
	// Load environment variables
	err := hyphen.LoadEnv(&hyphen.EnvOptions{
		Environment: "development",
		Local:       true,
	})
	if err != nil {
		log.Printf("Warning: Could not load env files: %v\n", err)
	}

	// Create a complete Hyphen client with all services using functional options
	client, err := hyphen.New(
		hyphen.WithPublicAPIKey("your_public_api_key"),
		hyphen.WithAPIKey("your_api_key"),
		hyphen.WithApplicationID("your_application_id"),
		hyphen.WithEnvironment("production"),
		hyphen.WithOrganizationID("your_organization_id"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Use Toggle service
	if client.Toggle != nil {
		fmt.Println("=== Toggle Service ===")
		enabled := client.Toggle.GetBoolean(ctx, "feature-flag", false, nil)
		fmt.Printf("Feature enabled: %v\n\n", enabled)
	}

	// Use NetInfo service
	if client.NetInfo != nil {
		fmt.Println("=== NetInfo Service ===")
		ipInfo, err := client.NetInfo.GetIPInfo(ctx, "8.8.8.8")
		if err != nil {
			log.Printf("NetInfo error: %v\n", err)
		} else {
			fmt.Printf("IP: %s, Country: %s\n\n", ipInfo.IP, ipInfo.Location.Country)
		}
	}

	// Use Link service
	if client.Link != nil {
		fmt.Println("=== Link Service ===")
		tags, err := client.Link.GetTags(ctx)
		if err != nil {
			log.Printf("Link error: %v\n", err)
		} else {
			fmt.Printf("Available tags: %v\n\n", tags)
		}
	}

	fmt.Println("All services accessed successfully!")
}
