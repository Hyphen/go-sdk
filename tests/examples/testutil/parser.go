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

	// Inject dev URI for NewNetInfo calls
	if strings.Contains(code, "hyphen.NewNetInfo(") {
		code = injectNetInfoDevURI(code)
	}

	// Inject dev URIs for NewLink calls
	if strings.Contains(code, "hyphen.NewLink(") {
		code = injectLinkDevURIs(code)
	}

	// Inject dev URIs for hyphen.New() calls that create NetInfo or Link
	if strings.Contains(code, "hyphen.New(") {
		code = injectClientDevURIs(code)
	}

	return code
}

// injectNetInfoDevURI adds WithNetInfoBaseURI for dev environment
func injectNetInfoDevURI(code string) string {
	return injectOption(code, "hyphen.NewNetInfo(",
		`hyphen.WithNetInfoBaseURI("https://dev.net.info")`)
}

// injectLinkDevURIs adds WithLinkURIs for dev environment
func injectLinkDevURIs(code string) string {
	return injectOption(code, "hyphen.NewLink(",
		`hyphen.WithLinkURIs([]string{"https://dev-api.hyphen.ai/api/organizations/{organizationId}/link/codes/"})`)
}

// injectClientDevURIs adds dev URIs to hyphen.New() calls
func injectClientDevURIs(code string) string {
	code = injectOption(code, "hyphen.New(",
		`hyphen.WithNetInfoBaseURI("https://dev.net.info")`)
	code = injectOption(code, "hyphen.New(",
		`hyphen.WithLinkURIs([]string{"https://dev-api.hyphen.ai/api/organizations/{organizationId}/link/codes/"})`)
	return code
}

// injectOption finds a constructor call and adds an option before its closing parenthesis
func injectOption(code, constructor, option string) string {
	startIdx := strings.Index(code, constructor)
	if startIdx == -1 {
		return code
	}

	// Find the matching closing parenthesis by counting parens
	parenCount := 0
	inString := false
	var stringChar rune
	closeIdx := -1

	for i := startIdx; i < len(code); i++ {
		ch := rune(code[i])

		// Track string literals to avoid counting parens inside strings
		if !inString && (ch == '"' || ch == '`' || ch == '\'') {
			inString = true
			stringChar = ch
		} else if inString && ch == stringChar && (i == 0 || code[i-1] != '\\') {
			inString = false
		}

		if !inString {
			if ch == '(' {
				parenCount++
			} else if ch == ')' {
				parenCount--
				if parenCount == 0 {
					closeIdx = i
					break
				}
			}
		}
	}

	if closeIdx == -1 {
		return code
	}

	// Check if there's already a trailing comma before the closing paren
	prefix := code[:closeIdx]
	trimmed := strings.TrimRight(prefix, " \t\n")
	if strings.HasSuffix(trimmed, ",") {
		// Already has trailing comma, just add the option
		return prefix + "\t\t" + option + ",\n\t" + code[closeIdx:]
	}
	// No trailing comma, add one
	return prefix + ",\n\t\t" + option + ",\n\t" + code[closeIdx:]
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
