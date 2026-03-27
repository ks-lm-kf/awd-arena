package logger

import (
	"log/slog"
	"os"
)

var log *slog.Logger

func Init(level string) {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l})
	log = slog.New(handler)
	slog.SetDefault(log)
}

func Get() *slog.Logger {
	if log == nil {
		Init("info")
	}
	return log
}

func Debug(msg string, args ...any) { Get().Debug(msg, args...) }
func Info(msg string, args ...any)  { Get().Info(msg, args...) }
func Warn(msg string, args ...any)  { Get().Warn(msg, args...) }
func Error(msg string, args ...any) { Get().Error(msg, args...) }
func With(args ...any) *slog.Logger { return Get().With(args...) }
