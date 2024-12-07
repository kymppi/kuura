package kuura

import (
	"log/slog"
	"os"
	"sync"
)

type LoggerConfig struct {
	DebugEnabled  bool
	PrettyEnabled bool
}

type LogFormat int

const (
	FormatJSON LogFormat = iota
	FormatPretty
)

type Logger struct {
	mu     sync.RWMutex
	logger *slog.Logger
	config LoggerConfig
}

func NewLogger(cfg LoggerConfig) *Logger {
	return &Logger{
		config: cfg,
	}
}

func (l *Logger) Get() *slog.Logger {
	l.mu.RLock()
	if l.logger != nil {
		defer l.mu.RUnlock()
		return l.logger
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logger != nil {
		return l.logger
	}

	var handler slog.Handler
	logLevel := slog.LevelInfo
	if l.config.DebugEnabled {
		logLevel = slog.LevelDebug
	}

	if l.config.PrettyEnabled {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: logLevel == slog.LevelDebug,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})
	}

	l.logger = slog.New(handler)
	return l.logger
}
