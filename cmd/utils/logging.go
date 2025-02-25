package utils

import (
	"os"

	libpack_logging "github.com/lukaszraczylo/graphql-monitoring-proxy/logging"
)

// Logger is a global logger instance
var Logger *libpack_logging.Logger

// InitLogger initializes the logger with the specified debug level
func InitLogger(debug bool) *libpack_logging.Logger {
	Logger = libpack_logging.New()
	if debug {
		Logger.SetOutput(os.Stdout).SetMinLogLevel(libpack_logging.LEVEL_DEBUG)
	}
	return Logger
}

// Debug logs a debug message
func Debug(message string, pairs map[string]interface{}) {
	if Logger != nil {
		Logger.Debug(&libpack_logging.LogMessage{
			Message: message,
			Pairs:   pairs,
		})
	}
}

// Info logs an info message
func Info(message string, pairs map[string]interface{}) {
	if Logger != nil {
		Logger.Info(&libpack_logging.LogMessage{
			Message: message,
			Pairs:   pairs,
		})
	}
}

// Error logs an error message
func Error(message string, pairs map[string]interface{}) {
	if Logger != nil {
		Logger.Error(&libpack_logging.LogMessage{
			Message: message,
			Pairs:   pairs,
		})
	}
}

// Critical logs a critical message
func Critical(message string, pairs map[string]interface{}) {
	if Logger != nil {
		Logger.Critical(&libpack_logging.LogMessage{
			Message: message,
			Pairs:   pairs,
		})
	}
}