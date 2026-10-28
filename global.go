package caelum

import (
	"context"
	"errors"
	"sync/atomic"
)

var defaultLogger atomic.Pointer[Logger]

var ErrNoDefault = errors.New("caelum: no default logger is set")

func Init(cfg Config) (*Logger, error) {
	logger, err := New(cfg)
	if err != nil {
		return nil, err
	}

	defaultLogger.Store(logger)

	return logger, nil
}

func SetDefault(logger *Logger) {
	defaultLogger.Store(logger)
}

func Default() *Logger {
	return defaultLogger.Load()
}

func fromDefault() (*Logger, error) {
	if logger := defaultLogger.Load(); logger != nil {
		return logger, nil
	}

	return nil, ErrNoDefault
}

func Log(level Level, message string, fields ...Field) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.Log(level, message, fields...)
}

func LogContext(ctx context.Context, level Level, message string, fields ...Field) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.LogContext(ctx, level, message, fields...)
}

func Error(message string, fields ...Field) error {
	return Log(LevelError, message, fields...)
}

func Warn(message string, fields ...Field) error {
	return Log(LevelWarn, message, fields...)
}

func Info(message string, fields ...Field) error {
	return Log(LevelInfo, message, fields...)
}

func Debug(message string, fields ...Field) error {
	return Log(LevelDebug, message, fields...)
}

func Trace(message string, fields ...Field) error {
	return Log(LevelTrace, message, fields...)
}

func ErrorContext(ctx context.Context, message string, fields ...Field) error {
	return LogContext(ctx, LevelError, message, fields...)
}

func WarnContext(ctx context.Context, message string, fields ...Field) error {
	return LogContext(ctx, LevelWarn, message, fields...)
}

func InfoContext(ctx context.Context, message string, fields ...Field) error {
	return LogContext(ctx, LevelInfo, message, fields...)
}

func DebugContext(ctx context.Context, message string, fields ...Field) error {
	return LogContext(ctx, LevelDebug, message, fields...)
}

func TraceContext(ctx context.Context, message string, fields ...Field) error {
	return LogContext(ctx, LevelTrace, message, fields...)
}

func LogEvent(value Event) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.Event(value)
}

func LogEventContext(ctx context.Context, value Event) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.EventContext(ctx, value)
}

func CreateSpan(value Event) (*Span, error) {
	logger, err := fromDefault()
	if err != nil {
		return nil, err
	}

	return logger.CreateSpan(value)
}

func CreateSpanContext(ctx context.Context, value Event) (*Span, error) {
	logger, err := fromDefault()
	if err != nil {
		return nil, err
	}

	return logger.CreateSpanContext(ctx, value)
}

func SetFilter(filter string) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.SetFilter(filter)
}

func Dropped() (uint64, error) {
	logger, err := fromDefault()
	if err != nil {
		return 0, err
	}

	return logger.Dropped()
}

func Flush() error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.Flush()
}

func Close() error {
	logger := defaultLogger.Load()
	if logger == nil {
		return ErrNoDefault
	}

	if err := logger.Close(); err != nil {
		return err
	}

	defaultLogger.CompareAndSwap(logger, nil)

	return nil
}
