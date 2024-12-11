package tempolite

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
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

type LogFormat string

const (
	TextFormat LogFormat = "text"
	JSONFormat LogFormat = "json"
)

func NewDefaultLogger(level slog.Leveler, format LogFormat) Logger {
	var handler slog.Handler

	// jsonFile, err := os.Create("logs.json")
	// if err != nil {
	// 	log.Fatalf("failed to create log file: %v", err)
	// }
	// defer jsonFile.Close()

	opts := &slog.HandlerOptions{Level: level}

	switch format {
	case JSONFormat:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// // MultiHandler to handle both text and JSON
	// multiHandler := slog.NewMultiHandler(textHandler, jsonHandler)

	return &defaultLogger{
		logger: slog.New(handler),
	}
}

func (l *defaultLogger) addSource(keysAndValues []interface{}) []interface{} {
	_, file, line, ok := runtime.Caller(2) // Skip two frames to get to the caller
	if ok {
		keysAndValues = append(keysAndValues, "source", file+":"+fmt.Sprintf("%d", line))
	}
	return keysAndValues
}

func (l *defaultLogger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.logger.DebugContext(ctx, msg, l.addSource(keysAndValues)...)
}

func (l *defaultLogger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.logger.InfoContext(ctx, msg, l.addSource(keysAndValues)...)
}

func (l *defaultLogger) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.logger.WarnContext(ctx, msg, l.addSource(keysAndValues)...)
}

func (l *defaultLogger) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.logger.ErrorContext(ctx, msg, l.addSource(keysAndValues)...)
}

func (l *defaultLogger) WithFields(fields map[string]interface{}) Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &defaultLogger{logger: l.logger.With(args...)}
}
