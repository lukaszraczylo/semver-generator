package cmd

import (
	"os"
	"testing"

	"github.com/lukaszraczylo/semver-generator/cmd/utils"
	"github.com/spf13/cobra"
	assertions "github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	// Save original os.Args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set up test args to avoid actual execution
	os.Args = []string{"semver-gen", "--version"}

	// Initialize logger
	utils.InitLogger(true)

	// Create a custom rootCmd for testing
	originalRootCmd := rootCmd
	defer func() { rootCmd = originalRootCmd }()

	// Create a test command that doesn't actually execute anything
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	// Add all the required flags to the test command
	testCmd.Flags().Bool("version", false, "Print version information")
	testCmd.Flags().String("repository", "test-repo", "Repository URL")
	testCmd.Flags().String("branch", "test-branch", "Repository branch")
	testCmd.Flags().String("config", "test-config", "Config file path")

	rootCmd = testCmd

	// Execute should not panic
	assertions.NotPanics(t, func() {
		Execute()
	}, "Execute should not panic")
}

func TestSetupCobra(t *testing.T) {
	// Initialize logger
	utils.InitLogger(true)

	// Create a test Setup instance
	testRepo := &Setup{}

	// Create a test command with flags
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.Flags().String("repository", "test-repo", "")
	cmd.Flags().String("branch", "test-branch", "")
	cmd.Flags().String("config", "test-config", "")

	// Save original rootCmd and restore after test
	originalRootCmd := rootCmd
	defer func() { rootCmd = originalRootCmd }()
	rootCmd = cmd

	// Set up test params
	originalParams := params
	defer func() { params = originalParams }()
	params = myParams{
		varUseLocal: true,
	}

	// Test setupCobra
	assertions.NotPanics(t, func() {
		testRepo.setupCobra()
	}, "setupCobra should not panic")

	// Verify values were set correctly
	assertions.Equal(t, "test-repo", testRepo.RepositoryName, "Repository name should be set")
	assertions.Equal(t, "test-branch", testRepo.RepositoryBranch, "Repository branch should be set")
	assertions.Equal(t, "test-config", testRepo.LocalConfigFile, "Config file should be set")
	assertions.True(t, testRepo.UseLocal, "UseLocal should be set to true")
}
