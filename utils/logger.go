package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	logLevel LogLevel = LogLevelInfo
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

func init() {
	// Initialize log level from environment
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch levelStr {
	case "debug":
		logLevel = LogLevelDebug
	case "info":
		logLevel = LogLevelInfo
	case "warn", "warning":
		logLevel = LogLevelWarn
	case "error":
		logLevel = LogLevelError
	default:
		logLevel = LogLevelInfo
	}
}

// SetLogLevel sets the current log level
func SetLogLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		logLevel = LogLevelDebug
	case "info":
		logLevel = LogLevelInfo
	case "warn", "warning":
		logLevel = LogLevelWarn
	case "error":
		logLevel = LogLevelError
	}
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	if logLevel <= LogLevelDebug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	if logLevel <= LogLevelInfo {
		log.Printf("[INFO] "+format, v...)
	}
}

// Warn logs a warning message
func Warn(format string, v ...interface{}) {
	if logLevel <= LogLevelWarn {
		log.Printf("[WARN] "+format, v...)
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if logLevel <= LogLevelError {
		log.Printf("[ERROR] "+format, v...)
	}
}

// Fatal logs a fatal error and exits
func Fatal(format string, v ...interface{}) {
	log.Fatalf("[FATAL] "+format, v...)
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// FormatJSON formats a message for JSON logging
func FormatJSON(key, value string) string {
	return fmt.Sprintf(`%s="%s"`, key, value)
}
