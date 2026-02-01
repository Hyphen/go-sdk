#!/bin/bash
#
# Repro script for toggle availability issue
#
# This demonstrates that toggles created via the Management API
# may appear in the Horizon API initially, then disappear after
# a short time. This is the opposite of expected eventual consistency.
#
# Required environment variables:
#   HYPHEN_ORGANIZATION_ID
#   HYPHEN_PROJECT_ID
#   HYPHEN_API_KEY
#   HYPHEN_PUBLIC_API_KEY
#   HYPHEN_APPLICATION_ID
#
# Optional:
#   HYPHEN_DEV=true (uses dev endpoints, otherwise uses production)

set -e

# Check required env vars
for var in HYPHEN_ORGANIZATION_ID HYPHEN_PROJECT_ID HYPHEN_API_KEY HYPHEN_PUBLIC_API_KEY HYPHEN_APPLICATION_ID; do
  if [ -z "${!var}" ]; then
    echo "Error: $var is not set"
    exit 1
  fi
done

# Set endpoints based on HYPHEN_DEV
if [ "$HYPHEN_DEV" = "true" ]; then
  API_BASE="https://dev-api.hyphen.ai"
  HORIZON_BASE="https://dev-horizon.hyphen.ai"
else
  API_BASE="https://api.hyphen.ai"
  HORIZON_BASE="https://toggle.hyphen.cloud"
fi

ORG_ID="$HYPHEN_ORGANIZATION_ID"
PROJECT_ID="$HYPHEN_PROJECT_ID"
API_KEY="$HYPHEN_API_KEY"
PUBLIC_KEY="$HYPHEN_PUBLIC_API_KEY"
APP_ID="$HYPHEN_APPLICATION_ID"

# Generate unique toggle key
TOGGLE_KEY="test-repro-$(date +%s)"

echo "=== Configuration ==="
echo "API Base: ${API_BASE}"
echo "Horizon Base: ${HORIZON_BASE}"
echo "Toggle Key: ${TOGGLE_KEY}"
echo ""

echo "=== Step 1: Create toggle via Management API ==="
curl -s -X POST "${API_BASE}/api/organizations/${ORG_ID}/projects/${PROJECT_ID}/toggles/" \
  -H "Content-Type: application/json" \
  -H "x-api-key: ${API_KEY}" \
  -d '{
    "key": "'"${TOGGLE_KEY}"'",
    "type": "boolean",
    "targets": [],
    "defaultValue": true,
    "description": "Repro test for timing issue"
  }' | jq .

echo ""
echo "=== Step 2: Check toggle count from Management API ==="
curl -s -X GET "${API_BASE}/api/organizations/${ORG_ID}/projects/${PROJECT_ID}/toggles/" \
  -H "x-api-key: ${API_KEY}" | jq '{total: .total, pageSize: .pageSize, pageNum: .pageNum, returnedCount: (.data | length)}'

echo ""
echo "=== Step 3: Check every 5s for 60s ==="
for i in $(seq 0 5 60); do
  RESPONSE=$(curl -s -X POST "${HORIZON_BASE}/toggle/evaluate" \
    -H "Content-Type: application/json" \
    -H "x-api-key: ${PUBLIC_KEY}" \
    -d '{
      "application": "'"${APP_ID}"'",
      "environment": "development",
      "targetingKey": "test-user"
    }')

  TOGGLE_COUNT=$(echo "$RESPONSE" | jq '.toggles | length')
  TOGGLE_VALUE=$(echo "$RESPONSE" | jq -c ".toggles[\"${TOGGLE_KEY}\"] // \"NOT FOUND\"")
  echo "${i}s: ${TOGGLE_VALUE} (${TOGGLE_COUNT} toggles returned)"

  if [ $i -lt 60 ]; then
    sleep 5
  fi
done

echo ""
echo "=== Summary ==="
echo "Toggle key: ${TOGGLE_KEY}"
echo ""
echo "Expected: Toggle should return {\"value\": true, ...} consistently"
echo "Actual: Toggle often appears then disappears, or vice versa"
