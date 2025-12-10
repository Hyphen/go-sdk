# Acceptance Tests

Acceptance tests validate the SDK against real Hyphen services. They require valid API credentials and will create/delete real resources.

## Environment Variables

### Required for Toggle Tests

| Variable | Description |
|----------|-------------|
| `HYPHEN_PUBLIC_API_KEY` | Public API key for Toggle service (starts with `public_`) |
| `HYPHEN_APPLICATION_ID` | Application ID for Toggle evaluations |

### Required for Toggle Admin (test setup/teardown)

| Variable | Description |
|----------|-------------|
| `HYPHEN_API_KEY` | API key with management permissions |
| `HYPHEN_ORGANIZATION_ID` | Organization ID (e.g., `org_...`) |
| `HYPHEN_PROJECT_ID` | Project ID for creating test toggles |

### Required for NetInfo Tests

| Variable | Description |
|----------|-------------|
| `HYPHEN_API_KEY` | API key for NetInfo service |

### Required for Link Tests

| Variable | Description |
|----------|-------------|
| `HYPHEN_API_KEY` | API key for Link service |
| `HYPHEN_ORGANIZATION_ID` | Organization ID |
| `HYPHEN_LINK_DOMAIN` | Domain for short codes (e.g., `test.h4n.link`) |

### Optional

| Variable | Description |
|----------|-------------|
| `HYPHEN_DEV` | Set to `true` to use dev environment endpoints |

## Running Tests

Tests use the `acceptance` build tag and are excluded from normal test runs.

Run all acceptance tests:

```bash
go test -tags=acceptance ./tests/acceptance/...
```

Run specific test suites:

```bash
# Toggle tests only
go test -tags=acceptance ./tests/acceptance/... -run TestToggleAcceptance

# NetInfo tests only
go test -tags=acceptance ./tests/acceptance/... -run TestNetInfoAcceptance

# Link tests only
go test -tags=acceptance ./tests/acceptance/... -run TestLinkAcceptance
```

Run with verbose output:

```bash
go test -tags=acceptance -v ./tests/acceptance/...
```

## Example Setup

```bash
export HYPHEN_PUBLIC_API_KEY="public_..."
export HYPHEN_API_KEY="your_api_key"
export HYPHEN_APPLICATION_ID="app_..."
export HYPHEN_ORGANIZATION_ID="org_..."
export HYPHEN_PROJECT_ID="proj_..."
export HYPHEN_LINK_DOMAIN="test.h4n.link"

go test -tags=acceptance -v ./tests/acceptance/...
```

## Notes

- Tests will fail if required environment variables are not set
- Toggle tests create temporary toggles and clean them up after each test
- Link tests create temporary short codes and QR codes, then clean them up
- Some tests include small delays for eventual consistency
