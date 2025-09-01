package logger

import (
	"fmt"
	"os"
)

// Logger represents a simple logger with debug mode capability
type Logger struct {
	debugMode bool
}

// New creates a new logger instance
func New(debugMode bool) *Logger {
	return &Logger{
		debugMode: debugMode,
	}
}

// Debug prints debug messages when debug mode is enabled
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.debugMode {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// Info prints informational messages
func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

// Error prints error messages
func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
}

// IsDebugMode returns whether debug mode is enabled
func (l *Logger) IsDebugMode() bool {
	return l.debugMode
}
