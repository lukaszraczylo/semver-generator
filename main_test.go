package main

import (
	"os"
	"testing"

	"github.com/lukaszraczylo/semver-generator/cmd"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	// Save original os.Args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set up test args to avoid actual execution
	os.Args = []string{"semver-gen", "--version"}

	// Save original cmd.PKG_VERSION and restore after test
	originalPkgVersion := cmd.PKG_VERSION
	defer func() { cmd.PKG_VERSION = originalPkgVersion }()

	// Set a test version
	PKG_VERSION = "test-version"

	// Test should not panic
	assert.NotPanics(t, func() {
		main()
	}, "main() should not panic")

	// Verify that the version was set correctly
	assert.Equal(t, "test-version", cmd.PKG_VERSION, "PKG_VERSION should be set correctly")
}
