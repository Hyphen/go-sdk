package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// RunGoCode writes code to a temp directory, sets up go.mod with local SDK,
// runs the code, and returns the combined stdout/stderr output.
func RunGoCode(t *testing.T, code string) (string, error) {
	t.Helper()

	tmpDir := t.TempDir()

	// Write main.go
	mainFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainFile, []byte(code), 0644)
	if err != nil {
		return "", err
	}

	// Initialize go module
	if err := runCmd(t, tmpDir, "go", "mod", "init", "example"); err != nil {
		return "", err
	}

	// Replace SDK with local version
	sdkPath, err := getSDKPath()
	if err != nil {
		return "", err
	}
	if err := runCmd(t, tmpDir, "go", "mod", "edit", "-replace", "github.com/Hyphen/hyphen-go-sdk="+sdkPath); err != nil {
		return "", err
	}

	// Tidy dependencies
	if err := runCmd(t, tmpDir, "go", "mod", "tidy"); err != nil {
		return "", err
	}

	// Run and capture output
	cmd := exec.Command("go", "run", "main.go")
	cmd.Dir = tmpDir
	cmd.Env = os.Environ() // Pass through env vars
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// runCmd runs a command in the specified directory and returns any error
func runCmd(t *testing.T, dir string, name string, args ...string) error {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command %s %v failed: %s\nOutput: %s", name, args, err, output)
		return err
	}
	return nil
}

// getSDKPath returns the absolute path to the SDK root directory
func getSDKPath() (string, error) {
	// Get the path to this file
	_, currentFile, _, _ := runtime.Caller(0)
	// Navigate up from tests/examples/testutil/ to SDK root
	sdkPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "..")
	return filepath.Abs(sdkPath)
}
