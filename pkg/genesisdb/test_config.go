package genesisdb

import (
	"bufio"
	"os"
	"strings"
)

// TestConfig holds configuration for tests
type TestConfig struct {
	APIURL     string
	APIVersion string
	AuthToken  string
	UseMocks   bool
}

// GetTestConfig returns test configuration from environment variables or defaults
func GetTestConfig() *TestConfig {
	// Load .env file if it exists
	loadEnvFile()

	// Check if we have real environment variables for testing
	apiURL := os.Getenv("TEST_GENESISDB_API_URL")
	apiVersion := os.Getenv("TEST_GENESISDB_API_VERSION")
	authToken := os.Getenv("TEST_GENESISDB_AUTH_TOKEN")

	if apiURL != "" && apiVersion != "" && authToken != "" {
		return &TestConfig{
			APIURL:     apiURL,
			APIVersion: apiVersion,
			AuthToken:  authToken,
			UseMocks:   false,
		}
	}

	// Fall back to default mock configuration
	return &TestConfig{
		APIURL:     "http://localhost:8080",
		APIVersion: "v1",
		AuthToken:  "secret",
		UseMocks:   true,
	}
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		return // .env file doesn't exist, that's OK
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only set if not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}