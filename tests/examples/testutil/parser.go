package testutil

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// CodeBlock represents a Go code block extracted from a markdown file
type CodeBlock struct {
	Code    string
	Line    int    // Line number in source for debugging
	Section string // README section name (if applicable)
}

// ExtractGoCodeBlocks parses a markdown file and returns all ```go blocks
func ExtractGoCodeBlocks(markdownPath string) ([]CodeBlock, error) {
	file, err := os.Open(markdownPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var blocks []CodeBlock
	var currentBlock strings.Builder
	var inCodeBlock bool
	var blockStartLine int
	var currentSection string

	scanner := bufio.NewScanner(file)
	lineNum := 0

	headingRegex := regexp.MustCompile(`^#+\s+(.+)$`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Track section headers
		if matches := headingRegex.FindStringSubmatch(line); matches != nil {
			currentSection = matches[1]
			continue
		}

		// Check for code block start
		if strings.HasPrefix(line, "```go") {
			inCodeBlock = true
			blockStartLine = lineNum
			currentBlock.Reset()
			continue
		}

		// Check for code block end
		if inCodeBlock && strings.HasPrefix(line, "```") {
			inCodeBlock = false
			blocks = append(blocks, CodeBlock{
				Code:    currentBlock.String(),
				Line:    blockStartLine,
				Section: sanitizeSection(currentSection),
			})
			continue
		}

		// Accumulate code block content
		if inCodeBlock {
			if currentBlock.Len() > 0 {
				currentBlock.WriteString("\n")
			}
			currentBlock.WriteString(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return blocks, nil
}

// ReplaceCredentials replaces placeholder strings with os.Getenv calls
func ReplaceCredentials(code string) string {
	replacements := map[string]string{
		`"your_public_api_key"`:  `os.Getenv("HYPHEN_PUBLIC_API_KEY")`,
		`"your_api_key"`:         `os.Getenv("HYPHEN_API_KEY")`,
		`"your_application_id"`:  `os.Getenv("HYPHEN_APPLICATION_ID")`,
		`"your_organization_id"`: `os.Getenv("HYPHEN_ORGANIZATION_ID")`,
		`"test.h4n.link"`:        `os.Getenv("HYPHEN_LINK_DOMAIN")`,
	}
	for placeholder, replacement := range replacements {
		code = strings.ReplaceAll(code, placeholder, replacement)
	}
	return code
}

// InjectDevURIs adds dev endpoint options when HYPHEN_DEV=true
func InjectDevURIs(code string) string {
	if os.Getenv("HYPHEN_DEV") != "true" {
		return code
	}

	// Inject dev URI for NetInfoOptions
	if strings.Contains(code, "hyphen.NetInfoOptions{}") {
		code = strings.ReplaceAll(code, "hyphen.NetInfoOptions{}",
			`hyphen.NetInfoOptions{BaseURI: "https://dev.net.info"}`)
	}

	// Inject dev URIs for LinkOptions
	if strings.Contains(code, "hyphen.LinkOptions{}") {
		code = strings.ReplaceAll(code, "hyphen.LinkOptions{}",
			`hyphen.LinkOptions{URIs: []string{"https://dev-api.hyphen.ai/api/organizations/{organizationId}/link/codes/"}}`)
	}

	// Inject dev URIs for ClientOptions (hyphen.New())
	if strings.Contains(code, "hyphen.ClientOptions{") {
		code = injectClientOptionsDevURIs(code)
	}

	return code
}

// injectClientOptionsDevURIs adds dev URIs to hyphen.ClientOptions struct
func injectClientOptionsDevURIs(code string) string {
	// Find ClientOptions struct and add dev fields before the closing brace
	// This is a simplified approach - look for the pattern and inject fields

	// Pattern: hyphen.ClientOptions{...}
	// We need to add:
	//   NetInfoBaseURI: "https://dev.net.info",
	//   LinkURIs: []string{"https://dev-api.hyphen.ai/api/organizations/{organizationId}/link/codes/"},

	// Find where ClientOptions ends (the closing brace before the next statement)
	startIdx := strings.Index(code, "hyphen.ClientOptions{")
	if startIdx == -1 {
		return code
	}

	// Find the matching closing brace
	braceCount := 0
	inString := false
	var stringChar rune
	closeIdx := -1

	for i := startIdx; i < len(code); i++ {
		ch := rune(code[i])

		// Track string literals
		if !inString && (ch == '"' || ch == '`' || ch == '\'') {
			inString = true
			stringChar = ch
		} else if inString && ch == stringChar && (i == 0 || code[i-1] != '\\') {
			inString = false
		}

		if !inString {
			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
				if braceCount == 0 {
					closeIdx = i
					break
				}
			}
		}
	}

	if closeIdx == -1 {
		return code
	}

	// Check if NetInfoBaseURI is already set
	optionsContent := code[startIdx:closeIdx]
	devFields := ""

	if !strings.Contains(optionsContent, "NetInfoBaseURI") {
		devFields += "\n\t\tNetInfoBaseURI: \"https://dev.net.info\","
	}
	if !strings.Contains(optionsContent, "LinkURIs") {
		devFields += "\n\t\tLinkURIs: []string{\"https://dev-api.hyphen.ai/api/organizations/{organizationId}/link/codes/\"},"
	}

	if devFields == "" {
		return code
	}

	// Insert before the closing brace
	return code[:closeIdx] + devFields + "\n\t" + code[closeIdx:]
}

// AddOsImport ensures "os" is in the imports if os.Getenv is used
func AddOsImport(code string) string {
	if !strings.Contains(code, "os.Getenv") {
		return code
	}

	// Check if os is already imported
	if strings.Contains(code, `"os"`) {
		return code
	}

	// Find the import block and add os to it
	// Handle both single import and import block formats

	// Single import format: import "..."
	singleImportRegex := regexp.MustCompile(`import\s+"([^"]+)"`)
	if singleImportRegex.MatchString(code) {
		// Convert single import to block and add os
		return singleImportRegex.ReplaceAllString(code, `import (
	"$1"
	"os"
)`)
	}

	// Import block format: import ( ... )
	importBlockRegex := regexp.MustCompile(`import\s*\(\s*\n`)
	if importBlockRegex.MatchString(code) {
		return importBlockRegex.ReplaceAllString(code, "import (\n\t\"os\"\n")
	}

	return code
}

// IsCompleteProgram checks if code has package main and func main
func IsCompleteProgram(code string) bool {
	hasPackageMain := strings.Contains(code, "package main")
	hasFuncMain := regexp.MustCompile(`func\s+main\s*\(`).MatchString(code)
	return hasPackageMain && hasFuncMain
}

// sanitizeSection converts a section name to a valid test name component
func sanitizeSection(section string) string {
	// Replace spaces and special characters with underscores
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, section)

	// Collapse multiple underscores
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}

	// Trim leading/trailing underscores
	result = strings.Trim(result, "_")

	return result
}
