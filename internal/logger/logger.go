package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/SantiagoBobrik/agent-pulse/internal/config"
	"github.com/lmittmann/tint"
)

var defaultLogger *slog.Logger

func init() {
	writers := []io.Writer{os.Stderr}

	if f, err := openLogFile(); err == nil {
		writers = append(writers, f)
	}

	defaultLogger = slog.New(tint.NewHandler(io.MultiWriter(writers...), &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	}))
}

func openLogFile() (*os.File, error) {
	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "server.log")
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

func SetOutput(w io.Writer) {
	defaultLogger = slog.New(tint.NewHandler(w, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	}))
}

func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

func Fatal(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
	os.Exit(1)
}
