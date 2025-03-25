package logger

import (
	"bytes"
	"finicky/window"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var memLog *bytes.Buffer
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

// createHandler creates a slog handler with the given writer
func createHandler(writer io.Writer) slog.Handler {
	return slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format time as ISO string with microseconds
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   slog.TimeKey,
					Value: slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05.000000")),
				}
			}
			return a
		},
	})
}

// Setup initializes the logger with basic configuration
func Setup() {
	// Start with in-memory logging
	memLog = &bytes.Buffer{}
	multiWriter := io.MultiWriter(memLog, os.Stdout, &windowWriter{})

	// Set the default logger
	slog.SetDefault(slog.New(createHandler(multiWriter)))
}

// SetupFile configures file logging if enabled
func SetupFile(shouldLog bool) error {
	slog.Debug("Setting up file logging", "shouldLog", shouldLog)
	if shouldLog {
		slog.Warn("Logging requests to disk. Logs may include sensitive information. Disable this by setting logRequests: false.")
	}


	if !shouldLog {
		// FIXME: We should recreate the logger if the file logging options has been changed while the app is running
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

	currentTime := time.Now().Format("2006-01-02_15-04-05.000")
	logFile := filepath.Join(logDir, fmt.Sprintf("Finicky_%s.log", currentTime))

	file, err = os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	slog.Info("Log file created", "path", logFile)

	// Write buffered logs to file
	if _, err := file.Write(memLog.Bytes()); err != nil {
		return fmt.Errorf("failed to write buffered logs: %v", err)
	}

	// Update writer to include file while preserving window writer
	multiWriter := io.MultiWriter(file, memLog, os.Stdout, &windowWriter{})

	// Update the default logger
	slog.SetDefault(slog.New(createHandler(multiWriter)))
	return nil
}

// Close properly closes the logger and any open file handles
func Close() {
	slog.Info("Application closed!")
	if file != nil {
		file.Close()
	}
}

// GetBufferedLogs returns all logs stored in memory
func GetBufferedLogs() string {
	return memLog.String()
}