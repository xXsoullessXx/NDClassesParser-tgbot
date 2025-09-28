package logger_test

import (
	"os"
	"strings"
	"testing"

	"NDClasses/clients/logger"
)

func TestNew(t *testing.T) {
	// Test creating logger with debug mode enabled
	l := logger.New(true)
	if !l.IsDebugMode() {
		t.Error("Expected debug mode to be enabled")
	}

	// Test creating logger with debug mode disabled
	l = logger.New(false)
	if l.IsDebugMode() {
		t.Error("Expected debug mode to be disabled")
	}
}

func TestIsDebugMode(t *testing.T) {
	l := logger.New(true)
	if !l.IsDebugMode() {
		t.Error("IsDebugMode should return true when debug mode is enabled")
	}

	l = logger.New(false)
	if l.IsDebugMode() {
		t.Error("IsDebugMode should return false when debug mode is disabled")
	}
}

func TestDebug(t *testing.T) {
	// Test debug output when debug mode is enabled
	l := logger.New(true)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	l.Debug("Test debug message: %s", "value")

	w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "[DEBUG] Test debug message: value") {
		t.Errorf("Expected debug output, got: %s", output)
	}

	// Test no output when debug mode is disabled
	l = logger.New(false)

	r2, w2, _ := os.Pipe()
	os.Stdout = w2

	l.Debug("This should not appear")

	w2.Close()
	os.Stdout = oldStdout

	buf2 := make([]byte, 1024)
	n2, _ := r2.Read(buf2)
	output2 := string(buf2[:n2])

	if len(output2) > 0 {
		t.Errorf("Expected no debug output when debug mode disabled, got: %s", output2)
	}
}

func TestInfo(t *testing.T) {
	l := logger.New(false)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	l.Info("Test info message: %d", 42)

	w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "[INFO] Test info message: 42") {
		t.Errorf("Expected info output, got: %s", output)
	}
}

func TestError(t *testing.T) {
	l := logger.New(false)

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	l.Error("Test error message: %s", "error details")

	w.Close()
	os.Stderr = oldStderr

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "[ERROR] Test error message: error details") {
		t.Errorf("Expected error output, got: %s", output)
	}
}

// Test basic functionality
func TestLoggerBasic(t *testing.T) {
	// Test that logger is created correctly
	l := logger.New(true)
	if !l.IsDebugMode() {
		t.Error("Debug mode should be set")
	}

	l = logger.New(false)
	if l.IsDebugMode() {
		t.Error("Debug mode should not be set")
	}
}
