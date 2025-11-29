package env

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// EnvOptions contains options for loading environment variables
type EnvOptions struct {
	Path        string
	Environment string
	Local       bool
}

// LoadEnv loads environment variables based on the default .env file
// and the current environment. The loading order is:
// .env -> .env.local -> .env.<environment> -> .env.<environment>.local
func LoadEnv(opts *EnvOptions) error {
	if opts == nil {
		opts = &EnvOptions{Local: true}
	}

	// Set the current working directory if provided
	currentWorkingDirectory := opts.Path
	if currentWorkingDirectory == "" {
		var err error
		currentWorkingDirectory, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	// Load the default .env file
	envPath := filepath.Join(currentWorkingDirectory, ".env")
	if fileExists(envPath) {
		_ = godotenv.Load(envPath)
	}

	// Load the .env.local file if it exists
	if opts.Local {
		localEnvPath := filepath.Join(currentWorkingDirectory, ".env.local")
		if fileExists(localEnvPath) {
			_ = godotenv.Overload(localEnvPath)
		}
	}

	// Load the environment specific .env file
	environment := opts.Environment
	if environment == "" {
		environment = os.Getenv("GO_ENV")
	}

	if environment != "" {
		envSpecificPath := filepath.Join(currentWorkingDirectory, ".env."+environment)
		if fileExists(envSpecificPath) {
			_ = godotenv.Overload(envSpecificPath)
		}

		// Load the environment specific .env.local file if it exists
		if opts.Local {
			envLocalPath := filepath.Join(currentWorkingDirectory, ".env."+environment+".local")
			if fileExists(envLocalPath) {
				_ = godotenv.Overload(envLocalPath)
			}
		}
	}

	return nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
