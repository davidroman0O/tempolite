If people want to implement their own logs, they will be able to easily:

```go
package main

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(logger *zap.Logger) Logger {
	return &zapLogger{logger: logger}
}

func (l *zapLogger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.logger.Debug(msg, l.zapFields(ctx, keysAndValues...)...)
}

func (l *zapLogger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, l.zapFields(ctx, keysAndValues...)...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.logger.Warn(msg, l.zapFields(ctx, keysAndValues...)...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.logger.Error(msg, l.zapFields(ctx, keysAndValues...)...)
}

func (l *zapLogger) WithFields(fields map[string]interface{}) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &zapLogger{logger: l.logger.With(zapFields...)}
}

func (l *zapLogger) zapFields(ctx context.Context, keysAndValues ...interface{}) []zapcore.Field {
	fields := make([]zapcore.Field, 0, len(keysAndValues)/2+1)
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		value := keysAndValues[i+1]
		fields = append(fields, zap.Any(key, value))
	}
	return fields
}
```

Or even

```go
package main

import (
	"context"

	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	logger *logrus.Logger
}

func NewLogrusLogger(logger *logrus.Logger) Logger {
	return &logrusLogger{logger: logger}
}

func (l *logrusLogger) log(ctx context.Context, level Level, msg string, keysAndValues ...interface{}) {
	fields := logrus.Fields{}
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		fields[key] = keysAndValues[i+1]
	}

	entry := l.logger.WithFields(fields)

	switch level {
	case LevelDebug:
		entry.Debug(msg)
	case LevelInfo:
		entry.Info(msg)
	case LevelWarn:
		entry.Warn(msg)
	case LevelError:
		entry.Error(msg)
	}
}

func (l *logrusLogger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log(ctx, LevelDebug, msg, keysAndValues...)
}

func (l *logrusLogger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log(ctx, LevelInfo, msg, keysAndValues...)
}

func (l *logrusLogger) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log(ctx, LevelWarn, msg, keysAndValues...)
}

func (l *logrusLogger) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log(ctx, LevelError, msg, keysAndValues...)
}

func (l *logrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &logrusLogger{logger: l.logger.WithFields(fields).Logger}
}
```
