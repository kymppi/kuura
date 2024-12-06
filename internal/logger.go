package kuura

import (
	"log/slog"
	"os"
	"sync"
)

var (
	logger     *slog.Logger
	loggerOnce sync.Once
	logLevel   = slog.LevelInfo
)

func SetLoggerDebugMode(enabled bool) {
	if enabled {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}
}

func ProvideLogger() *slog.Logger {
	loggerOnce.Do(func() {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))
	})
	return logger
}
