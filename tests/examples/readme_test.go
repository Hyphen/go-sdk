//go:build examples

package examples

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Hyphen/go-sdk/tests/examples/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestREADMEExamples(t *testing.T) {
	// Require all env vars upfront - FAIL if missing
	requireEnvVars(t,
		"HYPHEN_PUBLIC_API_KEY",
		"HYPHEN_APPLICATION_ID",
		"HYPHEN_API_KEY",
		"HYPHEN_ORGANIZATION_ID",
		"HYPHEN_LINK_DOMAIN",
	)

	blocks, err := testutil.ExtractGoCodeBlocks("../../README.md")
	require.NoError(t, err)

	completeCount := 0
	for _, block := range blocks {
		// Skip incomplete snippets (no package main)
		if !testutil.IsCompleteProgram(block.Code) {
			continue
		}
		completeCount++

		testName := fmt.Sprintf("line_%d_%s", block.Line, block.Section)
		t.Run(testName, func(t *testing.T) {
			// Transform code: replace placeholders with env var lookups
			code := testutil.ReplaceCredentials(block.Code)
			code = testutil.InjectDevURIs(code)
			code = testutil.AddOsImport(code)

			// Run and capture output
			output, err := testutil.RunGoCode(t, code)
			require.NoError(t, err, "Example at line %d failed with output: %s", block.Line, output)

			// Verify output based on what the example prints
			verifyExampleOutput(t, block, output)
		})
	}

	// Ensure we found at least some complete examples
	require.Greater(t, completeCount, 0, "No complete Go examples found in README.md")
}

func verifyExampleOutput(t *testing.T, block testutil.CodeBlock, output string) {
	t.Helper()

	// Check for expected output patterns based on code content
	if strings.Contains(block.Code, "GetBoolean") {
		assert.Contains(t, output, ":", "Expected key:value output from GetBoolean example")
	}
	if strings.Contains(block.Code, "GetString") {
		assert.Contains(t, output, ":", "Expected key:value output from GetString example")
	}
	if strings.Contains(block.Code, "GetNumber") {
		assert.Contains(t, output, ":", "Expected key:value output from GetNumber example")
	}
	if strings.Contains(block.Code, "GetIPInfo") {
		// NetInfo example should output IP-related info
		assert.Contains(t, output, "IP", "Expected IP output from GetIPInfo example")
	}
	if strings.Contains(block.Code, "CreateShortCode") {
		assert.Contains(t, output, "Short", "Expected Short code output from CreateShortCode example")
	}
}
