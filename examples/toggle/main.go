package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Hyphen/hyphen-go-sdk"
)

func main() {
	// Create a Toggle client using functional options
	toggle, err := hyphen.NewToggle(
		hyphen.WithPublicAPIKey("your_public_api_key"),
		hyphen.WithApplicationID("your_application_id"),
		hyphen.WithEnvironment("production"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Set up error handler
	toggle.SetErrorHandler(func(err error) {
		fmt.Printf("Toggle error: %v\n", err)
	})

	ctx := context.Background()

	// Get a boolean toggle
	enabled := toggle.GetBoolean(ctx, "feature-flag", false, nil)
	fmt.Printf("Feature enabled: %v\n", enabled)

	// Get a string toggle
	message := toggle.GetString(ctx, "welcome-message", "Hello World", nil)
	fmt.Printf("Welcome message: %s\n", message)

	// Get a number toggle
	maxRetries := toggle.GetNumber(ctx, "max-retries", 3.0, nil)
	fmt.Printf("Max retries: %.0f\n", maxRetries)

	// Get an object toggle
	config := toggle.GetObject(ctx, "app-config", map[string]interface{}{"theme": "light"}, nil)
	fmt.Printf("App config: %+v\n", config)

	// Use context override for specific user
	userContext := &hyphen.ToggleContext{
		TargetingKey: "user-123",
		User: &hyphen.ToggleUser{
			ID:    "user-123",
			Email: "john.doe@example.com",
			Name:  "John Doe",
		},
	}

	userEnabled := toggle.GetBoolean(ctx, "feature-flag", false, userContext)
	fmt.Printf("Feature enabled for user-123: %v\n", userEnabled)
}
