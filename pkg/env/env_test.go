package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadEnv(t *testing.T) {
	t.Run("loads_the_default_env_file_when_it_exists", func(t *testing.T) {
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, ".env")
		err := os.WriteFile(envFile, []byte("THE_VAR=theValue"), 0644)
		assert.NoError(t, err)

		err = LoadEnv(&EnvOptions{Path: tempDir, Local: false})

		assert.NoError(t, err)
		assert.Equal(t, "theValue", os.Getenv("THE_VAR"))
		t.Cleanup(func() { os.Unsetenv("THE_VAR") })
	})

	t.Run("does_not_return_an_error_when_default_env_file_does_not_exist", func(t *testing.T) {
		tempDir := t.TempDir()

		err := LoadEnv(&EnvOptions{Path: tempDir, Local: false})

		assert.NoError(t, err)
	})

	t.Run("loads_env_local_file_when_local_is_true", func(t *testing.T) {
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, ".env")
		envLocalFile := filepath.Join(tempDir, ".env.local")
		err := os.WriteFile(envFile, []byte("THE_VAR=aValue"), 0644)
		assert.NoError(t, err)
		err = os.WriteFile(envLocalFile, []byte("THE_VAR=theLocalValue"), 0644)
		assert.NoError(t, err)

		err = LoadEnv(&EnvOptions{Path: tempDir, Local: true})

		assert.NoError(t, err)
		assert.Equal(t, "theLocalValue", os.Getenv("THE_VAR"))
		t.Cleanup(func() { os.Unsetenv("THE_VAR") })
	})

	t.Run("does_not_load_env_local_file_when_local_is_false", func(t *testing.T) {
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, ".env")
		envLocalFile := filepath.Join(tempDir, ".env.local")
		err := os.WriteFile(envFile, []byte("THE_VAR=theValue"), 0644)
		assert.NoError(t, err)
		err = os.WriteFile(envLocalFile, []byte("THE_VAR=shouldNotLoad"), 0644)
		assert.NoError(t, err)

		err = LoadEnv(&EnvOptions{Path: tempDir, Local: false})

		assert.NoError(t, err)
		assert.Equal(t, "theValue", os.Getenv("THE_VAR"))
		t.Cleanup(func() { os.Unsetenv("THE_VAR") })
	})

	t.Run("loads_environment_specific_file_when_environment_is_set", func(t *testing.T) {
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, ".env")
		envDevFile := filepath.Join(tempDir, ".env.development")
		err := os.WriteFile(envFile, []byte("THE_VAR=aValue"), 0644)
		assert.NoError(t, err)
		err = os.WriteFile(envDevFile, []byte("THE_VAR=theDevelopmentValue"), 0644)
		assert.NoError(t, err)

		err = LoadEnv(&EnvOptions{Path: tempDir, Environment: "development", Local: false})

		assert.NoError(t, err)
		assert.Equal(t, "theDevelopmentValue", os.Getenv("THE_VAR"))
		t.Cleanup(func() { os.Unsetenv("THE_VAR") })
	})

	t.Run("loads_environment_specific_local_file_when_environment_and_local_are_set", func(t *testing.T) {
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, ".env")
		envDevFile := filepath.Join(tempDir, ".env.development")
		envDevLocalFile := filepath.Join(tempDir, ".env.development.local")
		err := os.WriteFile(envFile, []byte("THE_VAR=aValue"), 0644)
		assert.NoError(t, err)
		err = os.WriteFile(envDevFile, []byte("THE_VAR=anotherValue"), 0644)
		assert.NoError(t, err)
		err = os.WriteFile(envDevLocalFile, []byte("THE_VAR=theDevLocalValue"), 0644)
		assert.NoError(t, err)

		err = LoadEnv(&EnvOptions{Path: tempDir, Environment: "development", Local: true})

		assert.NoError(t, err)
		assert.Equal(t, "theDevLocalValue", os.Getenv("THE_VAR"))
		t.Cleanup(func() { os.Unsetenv("THE_VAR") })
	})

	t.Run("uses_go_env_environment_variable_when_environment_is_not_set", func(t *testing.T) {
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, ".env")
		envProdFile := filepath.Join(tempDir, ".env.production")
		err := os.WriteFile(envFile, []byte("THE_VAR=aValue"), 0644)
		assert.NoError(t, err)
		err = os.WriteFile(envProdFile, []byte("THE_VAR=theProductionValue"), 0644)
		assert.NoError(t, err)
		os.Setenv("GO_ENV", "production")
		t.Cleanup(func() { os.Unsetenv("GO_ENV") })

		err = LoadEnv(&EnvOptions{Path: tempDir, Local: false})

		assert.NoError(t, err)
		assert.Equal(t, "theProductionValue", os.Getenv("THE_VAR"))
		t.Cleanup(func() { os.Unsetenv("THE_VAR") })
	})

	t.Run("does_not_load_environment_specific_files_when_environment_is_empty", func(t *testing.T) {
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, ".env")
		envDevFile := filepath.Join(tempDir, ".env.development")
		err := os.WriteFile(envFile, []byte("THE_VAR=theValue"), 0644)
		assert.NoError(t, err)
		err = os.WriteFile(envDevFile, []byte("THE_VAR=shouldNotLoad"), 0644)
		assert.NoError(t, err)

		err = LoadEnv(&EnvOptions{Path: tempDir, Environment: "", Local: false})

		assert.NoError(t, err)
		assert.Equal(t, "theValue", os.Getenv("THE_VAR"))
		t.Cleanup(func() { os.Unsetenv("THE_VAR") })
	})

	t.Run("uses_current_working_directory_when_path_is_not_provided", func(t *testing.T) {
		originalWd, err := os.Getwd()
		assert.NoError(t, err)
		tempDir := t.TempDir()
		err = os.Chdir(tempDir)
		assert.NoError(t, err)
		t.Cleanup(func() { os.Chdir(originalWd) })
		envFile := filepath.Join(tempDir, ".env")
		err = os.WriteFile(envFile, []byte("THE_VAR=theValue"), 0644)
		assert.NoError(t, err)

		err = LoadEnv(&EnvOptions{Local: false})

		assert.NoError(t, err)
		assert.Equal(t, "theValue", os.Getenv("THE_VAR"))
		t.Cleanup(func() { os.Unsetenv("THE_VAR") })
	})

	t.Run("defaults_to_local_true_when_opts_is_nil", func(t *testing.T) {
		originalWd, err := os.Getwd()
		assert.NoError(t, err)
		tempDir := t.TempDir()
		err = os.Chdir(tempDir)
		assert.NoError(t, err)
		t.Cleanup(func() { os.Chdir(originalWd) })
		envFile := filepath.Join(tempDir, ".env")
		envLocalFile := filepath.Join(tempDir, ".env.local")
		err = os.WriteFile(envFile, []byte("THE_VAR=aValue"), 0644)
		assert.NoError(t, err)
		err = os.WriteFile(envLocalFile, []byte("THE_VAR=theLocalValue"), 0644)
		assert.NoError(t, err)

		err = LoadEnv(nil)

		assert.NoError(t, err)
		assert.Equal(t, "theLocalValue", os.Getenv("THE_VAR"))
		t.Cleanup(func() { os.Unsetenv("THE_VAR") })
	})

	t.Run("loads_files_in_the_correct_priority_order", func(t *testing.T) {
		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, ".env")
		envLocalFile := filepath.Join(tempDir, ".env.local")
		envDevFile := filepath.Join(tempDir, ".env.development")
		envDevLocalFile := filepath.Join(tempDir, ".env.development.local")

		err := os.WriteFile(envFile, []byte("VAR1=base\nVAR2=base\nVAR3=base\nVAR4=base"), 0644)
		assert.NoError(t, err)
		err = os.WriteFile(envLocalFile, []byte("VAR2=local\nVAR3=local\nVAR4=local"), 0644)
		assert.NoError(t, err)
		err = os.WriteFile(envDevFile, []byte("VAR3=dev\nVAR4=dev"), 0644)
		assert.NoError(t, err)
		err = os.WriteFile(envDevLocalFile, []byte("VAR4=theDevLocal"), 0644)
		assert.NoError(t, err)

		err = LoadEnv(&EnvOptions{Path: tempDir, Environment: "development", Local: true})

		assert.NoError(t, err)
		assert.Equal(t, "base", os.Getenv("VAR1"))
		assert.Equal(t, "local", os.Getenv("VAR2"))
		assert.Equal(t, "dev", os.Getenv("VAR3"))
		assert.Equal(t, "theDevLocal", os.Getenv("VAR4"))
		t.Cleanup(func() {
			os.Unsetenv("VAR1")
			os.Unsetenv("VAR2")
			os.Unsetenv("VAR3")
			os.Unsetenv("VAR4")
		})
	})
}

func TestFileExists(t *testing.T) {
	t.Run("returns_true_when_file_exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(testFile, []byte("content"), 0644)
		assert.NoError(t, err)

		result := fileExists(testFile)

		assert.True(t, result)
	})

	t.Run("returns_false_when_file_does_not_exist", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "nonexistent.txt")

		result := fileExists(testFile)

		assert.False(t, result)
	})
}
