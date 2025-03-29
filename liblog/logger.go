package liblog

import (
	"log/slog"
	"os"
	"sync"
)

var once = sync.Once{}

func init() {
	once.Do(func() {
		_ = New(WithLogLevel(LogLevelInfo))
	})
}

type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

type options struct {
	level LogLevel
}

type Option func(*options)

func WithLogLevel(level LogLevel) Option {
	return func(o *options) {
		o.level = level
	}
}

func New(opts ...Option) *slog.Logger {
	options := options{}
	for _, o := range opts {
		o(&options)
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     mapLogLevel(options.level),
		AddSource: true,
	})

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger
}

func mapLogLevel(level LogLevel) slog.Level {
	switch level {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelInfo:
		return slog.LevelInfo
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelError
	}
}
