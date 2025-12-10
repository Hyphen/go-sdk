# Hyphen Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/Hyphen/go-sdk.svg)](https://pkg.go.dev/github.com/Hyphen/go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/Hyphen/go-sdk)](https://goreportcard.com/report/github.com/Hyphen/go-sdk)

The Hyphen Go SDK is a Go library that allows developers to easily integrate Hyphen's feature flag service [Toggle](https://hyphen.ai/toggle), secret management service [ENV](https://hyphen.ai/env), and geo information service [Net Info](https://hyphen.ai/net-info) into their Go applications.

## Table of Contents

- [Installation](#installation)
- [Basic Usage with Hyphen](#basic-usage-with-hyphen)
- [Toggle - Feature Flag Service](#toggle---feature-flag-service)
  - [Toggle Options](#toggle-options)
  - [Toggle API](#toggle-api)
  - [Toggle Error Handling](#toggle-error-handling)
  - [Toggle Environment Variables](#toggle-environment-variables)
  - [Toggle Self-Hosted](#toggle-self-hosted)
- [ENV - Secret Management Service](#env---secret-management-service)
- [Net Info - Geo Information Service](#net-info---geo-information-service)
- [Link - Short Code Service](#link---short-code-service)
- [Contributing](#contributing)
- [License and Copyright](#license-and-copyright)

## Installation

To install the Hyphen Go SDK, you can use `go get`:

```bash
go get github.com/Hyphen/go-sdk
```

## Basic Usage with Hyphen

There are many ways to use the Hyphen Go SDK. To get started, you can create an instance of the `Client` using functional options:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Hyphen/go-sdk"
)

func main() {
	client, err := hyphen.New(
		hyphen.WithPublicAPIKey("your_public_api_key"),
		hyphen.WithApplicationID("your_application_id"),
	)
	if err != nil {
		log.Fatal(err)
	}

	result := client.Toggle.GetBoolean(context.Background(), "hyphen-sdk-boolean", false, nil)
	fmt.Printf("Boolean toggle value: %v\n", result)
}
```

You can also use context for more advanced targeting:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Hyphen/go-sdk"
)

func main() {
	ctx := &hyphen.ToggleContext{
		TargetingKey: "user-123",
		IPAddress:    "203.0.113.42",
		CustomAttributes: hyphen.ToggleCustomAttrs{
			"subscriptionLevel": "premium",
			"region":            "us-east",
		},
		User: &hyphen.ToggleUser{
			ID:    "user-123",
			Email: "john.doe@example.com",
			Name:  "John Doe",
			CustomAttributes: hyphen.ToggleCustomAttrs{
				"role": "admin",
			},
		},
	}

	client, err := hyphen.New(
		hyphen.WithPublicAPIKey("your_public_api_key"),
		hyphen.WithApplicationID("your_application_id"),
		hyphen.WithDefaultContext(ctx),
	)
	if err != nil {
		log.Fatal(err)
	}

	result := client.Toggle.GetBoolean(context.Background(), "hyphen-sdk-boolean", false, nil)
	fmt.Printf("Boolean toggle value: %v\n", result)
}
```

## Toggle - Feature Flag Service

[Toggle](https://hyphen.ai/toggle) is our feature flag service that allows you to control the rollout of new features to your users.

### Toggle Options

| Option | Description |
|--------|-------------|
| `WithPublicAPIKey(key)` | The public API key for your Hyphen project. Must start with "public_". |
| `WithApplicationID(id)` | The application ID for your Hyphen project. |
| `WithEnvironment(env)` | The environment for your Hyphen project (e.g., "production"). Defaults to "development". |
| `WithDefaultContext(ctx)` | The default context to use when one is not passed to getter methods. |
| `WithHorizonURLs(urls)` | Array of Horizon endpoint URLs for load balancing and failover. |
| `WithDefaultTargetingKey(key)` | Default targeting key to use if one cannot be derived from context. |

### Toggle API

#### Creating a Toggle Client

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Hyphen/go-sdk"
)

func main() {
	toggle, err := hyphen.NewToggle(
		hyphen.WithPublicAPIKey("your_public_api_key"),
		hyphen.WithApplicationID("your_application_id"),
		hyphen.WithEnvironment("production"),
	)
	if err != nil {
		log.Fatal(err)
	}

	enabled := toggle.GetBoolean(context.Background(), "my-feature", false, nil)
	fmt.Printf("Feature enabled: %v\n", enabled)
}
```

#### Getting Toggle Values

```go
// Get a boolean toggle
enabled := toggle.GetBoolean(ctx, "feature-flag", false, nil)

// Get a string toggle
message := toggle.GetString(ctx, "welcome-message", "Hello!", nil)

// Get a number toggle
maxRetries := toggle.GetNumber(ctx, "max-retries", 3.0, nil)

// Get an object toggle
config := toggle.GetObject(ctx, "app-config", map[string]interface{}{"theme": "light"}, nil)
```

#### Context Override

You can override the context for a single request:

```go
overrideContext := &hyphen.ToggleContext{
	TargetingKey: "user-456",
	User: &hyphen.ToggleUser{
		ID:    "user-456",
		Email: "jane.doe@example.com",
	},
}

result := toggle.GetBoolean(ctx, "feature-flag", false, overrideContext)
```

### Toggle Error Handling

The SDK provides error handling through a callback mechanism:

```go
toggle, err := hyphen.NewToggle(
	hyphen.WithPublicAPIKey("your_public_api_key"),
	hyphen.WithApplicationID("your_application_id"),
)
if err != nil {
	panic(err)
}

// Set error handler
toggle.SetErrorHandler(func(err error) {
	fmt.Printf("Toggle error: %v\n", err)
})

// When an error occurs, the handler will be called and the default value will be returned
result := toggle.GetBoolean(ctx, "feature-flag", false, nil)
```

### Toggle Environment Variables

You can use environment variables to set the `PublicAPIKey` and `ApplicationID`:

```bash
export HYPHEN_PUBLIC_API_KEY=your_public_api_key
export HYPHEN_APPLICATION_ID=your_application_id
```

The SDK will automatically check for these environment variables during initialization.

### Toggle Self-Hosted

If you are using a self-hosted version of Hyphen, you can use the `WithHorizonURLs` option:

```go
toggle, err := hyphen.NewToggle(
	hyphen.WithPublicAPIKey("your_public_api_key"),
	hyphen.WithApplicationID("your_application_id"),
	hyphen.WithHorizonURLs([]string{"https://your-self-hosted-horizon-url"}),
)
```

## ENV - Secret Management Service

Hyphen's secret management service [ENV](https://hyphen.ai/env) allows you to manage your environment variables in a secure way.

### Loading Environment Variables

```go
package main

import (
	"github.com/Hyphen/go-sdk"
)

func main() {
	// Load default .env files
	err := hyphen.LoadEnv(nil)
	if err != nil {
		panic(err)
	}

	// Or with options
	err = hyphen.LoadEnv(&hyphen.EnvOptions{
		Path:        "/path/to/your/env/files/",
		Environment: "development",
		Local:       true,
	})
	if err != nil {
		panic(err)
	}
}
```

The loading order is:
```
.env -> .env.local -> .env.<environment> -> .env.<environment>.local
```

## Net Info - Geo Information Service

The Hyphen Go SDK provides a `NetInfo` client for fetching geo information about IP addresses.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Hyphen/go-sdk"
)

func main() {
	netInfo, err := hyphen.NewNetInfo(
		hyphen.WithAPIKey("your_api_key"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Get info for a single IP
	ipInfo, err := netInfo.GetIPInfo(context.Background(), "8.8.8.8")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("IP Info: %+v\n", ipInfo)

	// Get info for multiple IPs
	ips := []string{"8.8.8.8", "1.1.1.1"}
	ipInfos, err := netInfo.GetIPInfos(context.Background(), ips)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("IP Infos: %+v\n", ipInfos)
}
```

## Link - Short Code Service

The Hyphen Go SDK provides a `Link` client for creating and managing short codes and QR codes.

### Creating a Short Code

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Hyphen/go-sdk"
)

func main() {
	link, err := hyphen.NewLink(
		hyphen.WithAPIKey("your_api_key"),
		hyphen.WithOrganizationID("your_organization_id"),
	)
	if err != nil {
		log.Fatal(err)
	}

	shortCode, err := link.CreateShortCode(
		context.Background(),
		"https://hyphen.ai",
		"test.h4n.link",
		&hyphen.CreateShortCodeOptions{
			Title: "My Short Code",
			Tags:  []string{"sdk-test"},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Short Code: %+v\n", shortCode)
}
```

### Other Link Operations

```go
// Get a short code
shortCode, err := link.GetShortCode(ctx, "code_1234567890")

// Update a short code
updated, err := link.UpdateShortCode(ctx, "code_1234567890", &hyphen.UpdateShortCodeOptions{
	Title: "Updated Title",
})

// Delete a short code
err = link.DeleteShortCode(ctx, "code_1234567890")

// Create a QR code
qrCode, err := link.CreateQRCode(ctx, "code_1234567890", &hyphen.CreateQRCodeOptions{
	Title: "My QR Code",
	Size:  hyphen.QRSizeMedium,
})

// Get QR codes for a short code
qrCodes, err := link.GetQRCodes(ctx, "code_1234567890", 1, 10)

// Delete a QR code
err = link.DeleteQRCode(ctx, "code_1234567890", "qr_1234567890")
```

## All Available Options

The SDK uses a unified functional options pattern. Here are all available options:

| Option | Used By | Description |
|--------|---------|-------------|
| `WithAPIKey(key)` | Link, NetInfo | API key for authentication |
| `WithPublicAPIKey(key)` | Toggle | Public API key (must start with "public_") |
| `WithApplicationID(id)` | Toggle | Application ID |
| `WithEnvironment(env)` | Toggle | Environment name (defaults to "development") |
| `WithOrganizationID(id)` | Link | Organization ID |
| `WithDefaultContext(ctx)` | Toggle | Default evaluation context |
| `WithHorizonURLs(urls)` | Toggle | Custom Horizon endpoint URLs |
| `WithDefaultTargetingKey(key)` | Toggle | Default targeting key |
| `WithNetInfoBaseURI(uri)` | NetInfo | Custom base URI |
| `WithLinkURIs(uris)` | Link | Custom Link service URIs |

## Contributing

We welcome contributions to the Hyphen Go SDK! If you have an idea for a new feature, bug fix, or improvement, please follow these steps:

1. Fork the repository
2. Create a new branch for your feature or bug fix
3. Make your changes and commit them with a clear message
4. Push your changes to your forked repository
5. Create a pull request to the main repository

## Testing

To run the tests:

```bash
go test ./...
```

To run tests with coverage:

```bash
go test -cover ./...
```

## License and Copyright

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

The copyright for this project is held by Hyphen, Inc. All rights reserved.
