// Package shared contains shared code for all lib modules.
package shared

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

var (
	// Logger is the logger for the app.
	Logger *slog.Logger
	done   bool
)

func deriveLoggingLevel() slog.Level {
	defaultLevel := slog.LevelInfo
	rawValue, exists := os.LookupEnv("dafavorites_logging_level")
	if !exists {
		return defaultLevel
	}

	value, found := map[string]slog.Level{
		"info":  slog.LevelInfo,
		"debug": slog.LevelDebug,
		"error": slog.LevelError,
	}[strings.ToLower(rawValue)]
	if !found {
		return defaultLevel
	}
	return value
}

// InitLogging initializes logging.
func InitLogging() {
	if done {
		return
	}
	file, err := os.OpenFile(
		"dafavorites.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0640)
	if err != nil {
		panic(err)
	}
	initLogger(file, deriveLoggingLevel())
}

// InitTestLogging creates an slog logger that writes to t.Log.
func InitTestLogging(tb testing.TB) {
	initLogger(&testWriter{tb: tb}, slog.LevelDebug)
}

func initLogger(writer io.Writer, level slog.Level) {
	options := &slog.HandlerOptions{Level: level}
	Logger = slog.New(slog.NewTextHandler(writer, options))
	done = true
}

type testWriter struct {
	tb testing.TB
}

func (w testWriter) Write(p []byte) (n int, err error) {
	w.tb.Log(string(p))
	return len(p), nil
}
