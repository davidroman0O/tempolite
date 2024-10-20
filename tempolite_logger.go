package tempolite

import (
	"context"
	"log/slog"
	"os"
)

// Level represents the severity of a log message
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger is the interface that wraps the basic logging methods.
type Logger interface {
	Debug(ctx context.Context, msg string, keysAndValues ...interface{})
	Info(ctx context.Context, msg string, keysAndValues ...interface{})
	Warn(ctx context.Context, msg string, keysAndValues ...interface{})
	Error(ctx context.Context, msg string, keysAndValues ...interface{})
	WithFields(fields map[string]interface{}) Logger
}

type defaultLogger struct {
	logger *slog.Logger
}

func NewDefaultLogger() Logger {
	return &defaultLogger{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

func (l *defaultLogger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.logger.DebugContext(ctx, msg, keysAndValues...)
}

func (l *defaultLogger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.logger.InfoContext(ctx, msg, keysAndValues...)
}

func (l *defaultLogger) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.logger.WarnContext(ctx, msg, keysAndValues...)
}

func (l *defaultLogger) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.logger.ErrorContext(ctx, msg, keysAndValues...)
}

func (l *defaultLogger) WithFields(fields map[string]interface{}) Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &defaultLogger{logger: l.logger.With(args...)}
}
