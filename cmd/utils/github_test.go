package utils

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckLatestRelease(t *testing.T) {
	// Initialize logger
	InitLogger(true)

	// Save original environment variables
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", originalToken)

	// Test with no token
	os.Unsetenv("GITHUB_TOKEN")
	release, ok := CheckLatestRelease()
	assert.Equal(t, "[no GITHUB_TOKEN set]", release, "Should return no token message")
	assert.False(t, ok, "Should return false when no token is set")

	// Test with token but simulating API error
	// Set a dummy token that won't work with the GitHub API
	os.Setenv("GITHUB_TOKEN", "dummy-token")
	release, ok = CheckLatestRelease()
	assert.Equal(t, "", release, "Should return empty string on API error")
	assert.False(t, ok, "Should return false on API error")

	// We can't reliably test the successful API call in unit tests
	// as it would require a valid GitHub token and network access
}

func TestUpdatePackage(t *testing.T) {
	// Initialize logger
	InitLogger(true)

	// Save original environment variables
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", originalToken)

	// Test with no token
	os.Unsetenv("GITHUB_TOKEN")
	result := UpdatePackage()
	assert.False(t, result, "Should return false when no token is set")

	// Test with token but simulating API error
	os.Setenv("GITHUB_TOKEN", "dummy-token")
	result = UpdatePackage()
	assert.False(t, result, "Should return false on API error")

	// Create a test flag to simulate test mode
	if flag.Lookup("test.v") == nil {
		// This is a hack to simulate the test flag being set
		// which is used in the UpdatePackage function to skip actual download
		flag.Bool("test.v", true, "")
	}

	// We can't fully test the update functionality as it would modify the binary
	// but we've tested the token check logic and API error handling
}

// Note: We're not using mock transports for these tests to avoid
// adding complexity. The tests focus on the token presence logic and error handling.
