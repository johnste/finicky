package logger

import (
	"bytes"
	"finicky/window"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var memLog *bytes.Buffer
var multiWriter io.Writer
var file *os.File

// windowWriter implements io.Writer to send logs to the window
type windowWriter struct{}

func (w *windowWriter) Write(p []byte) (n int, err error) {
	// Remove trailing newline if present
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}

	window.SendMessageToWebView("log", msg)
	return len(p), nil
}

// Setup initializes the logger with basic configuration
func Setup() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// Start with in-memory logging
	memLog = &bytes.Buffer{}
	multiWriter = io.MultiWriter(memLog, os.Stdout, &windowWriter{})
	log.SetOutput(multiWriter)
}

// SetupFile configures file logging if enabled
func SetupFile(shouldLog bool) error {
	if !shouldLog {
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	logDir := filepath.Join(homeDir, "Library", "Logs", "Finicky")
	err = os.MkdirAll(logDir, 0755) // Create directory if it doesn't exist
	if err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	currentTime := time.Now().Format("2025-01-02_15-04-05.000")
	logFile := filepath.Join(logDir, fmt.Sprintf("Finicky_%s.log", currentTime))

	file, err = os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// Write buffered logs to file
	if _, err := file.Write(memLog.Bytes()); err != nil {
		return fmt.Errorf("failed to write buffered logs: %v", err)
	}

	// Update writer to include file while preserving window writer
	multiWriter = io.MultiWriter(file, memLog, os.Stdout, &windowWriter{})
	log.SetOutput(multiWriter)
	return nil
}

// Close properly closes the logger and any open file handles
func Close() {
	log.Println("Application closed!")
	if file != nil {
		file.Close()
	}
}

// GetBufferedLogs returns all logs stored in memory
func GetBufferedLogs() string {
	return memLog.String()
}