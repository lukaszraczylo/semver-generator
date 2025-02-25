package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)
func TestInitLogger(t *testing.T) {
	// Test with debug mode enabled
	logger := InitLogger(true)
	assert.NotNil(t, logger, "Logger should not be nil")
	assert.NotNil(t, Logger, "Global logger should not be nil")

	// Test with debug mode disabled
	logger = InitLogger(false)
	assert.NotNil(t, logger, "Logger should not be nil")
	assert.NotNil(t, Logger, "Global logger should not be nil")
}

func TestLoggingFunctions(t *testing.T) {
	// Initialize logger with debug mode
	InitLogger(true)

	// Just test that these don't panic
	Debug("Debug message", map[string]interface{}{"key": "value"})
	Info("Info message", map[string]interface{}{"key": "value"})
	Error("Error message", map[string]interface{}{"key": "value"})
	
	// Skip testing Critical as it might call os.Exit
	// Critical("Critical message", map[string]interface{}{"key": "value"})
	
	// Test passes if we get here without panicking
	assert.True(t, true)
}

func TestLoggingWithNilLogger(t *testing.T) {
	// Temporarily set logger to nil
	oldLogger := Logger
	Logger = nil
	defer func() { Logger = oldLogger }()

	// These should not panic
	Debug("Debug message", map[string]interface{}{"key": "value"})
	Info("Info message", map[string]interface{}{"key": "value"})
	Error("Error message", map[string]interface{}{"key": "value"})
	
	// Skip testing Critical as it might call os.Exit
	// Critical("Critical message", map[string]interface{}{"key": "value"})
	
	// Test passes if we get here without panicking
	assert.True(t, true)
}

// TestCriticalNilLogger tests that the Critical function doesn't panic with a nil logger
func TestCriticalNilLogger(t *testing.T) {
	// Save original logger and restore after test
	originalLogger := Logger
	defer func() { Logger = originalLogger }()
	
	// Set logger to nil
	Logger = nil
	
	// This should not panic
	Critical("Critical message", map[string]interface{}{"key": "value"})
	
	// Test passes if we get here without panicking
	assert.True(t, true)
}

// Note: We don't test Critical with an actual logger because it calls os.Exit