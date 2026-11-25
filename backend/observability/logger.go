package observability

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	mu     sync.RWMutex
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	file   *os.File
)

// Configure enables structured logs to stdout and, when configured, a file.
func Configure(levelName, format, filePath string) error {
	level := slog.LevelInfo
	switch strings.ToLower(levelName) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	writers := []io.Writer{os.Stdout}
	var opened *os.File
	if filePath != "" {
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}
		var err error
		opened, err = os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
		if err != nil {
			return err
		}
		writers = append(writers, opened)
	}
	output := io.MultiWriter(writers...)
	options := &slog.HandlerOptions{Level: level}
	var handler slog.Handler = slog.NewJSONHandler(output, options)
	if strings.EqualFold(format, "text") {
		handler = slog.NewTextHandler(output, options)
	}

	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		_ = file.Close()
	}
	file = opened
	logger = slog.New(handler)
	return nil
}

func Info(message string, args ...any)  { current().Info(message, args...) }
func Warn(message string, args ...any)  { current().Warn(message, args...) }
func Error(message string, args ...any) { current().Error(message, args...) }

func current() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return logger
}
