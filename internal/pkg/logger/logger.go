package logger

import (
	"log/slog"
	"os"
)

type Config struct {
	Level   string
	LokiURL string
	Labels  map[string]string
}

func New(cfg Config) *slog.Logger {
	var handlers []slog.Handler

	opts := &slog.HandlerOptions{
		Level: parseLevel(cfg.Level),
	}
	handlers = append(handlers, slog.NewJSONHandler(os.Stdout, opts))

	if cfg.LokiURL != "" {
		lokiHandler := NewLokiHandler(cfg.LokiURL, cfg.Labels)
		handlers = append(handlers, lokiHandler)
	}

	multiHandler := NewMultiHandler(handlers...)
	return slog.New(multiHandler)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
