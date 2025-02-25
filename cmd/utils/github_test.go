package utils

import (
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

	// We can't reliably test with a token in CI environments
	// Just verify the no-token case works as expected
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

	// We can't fully test the update functionality as it would modify the binary
	// but we can test the token check logic
}

// Note: We're not using mock transports for these tests to avoid
// adding complexity. The tests focus on the token presence logic.