package logger

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	}))
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
