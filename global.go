package caelum

import (
	"context"
	"errors"
	"sync/atomic"
)

var defaultLogger atomic.Pointer[Logger]

var ErrNoDefault = errors.New("caelum: no default logger is set")

// Init creates a logger and installs it as the package default.
func Init(cfg Config) (*Logger, error) {
	logger, err := New(cfg)
	if err != nil {
		return nil, err
	}

	defaultLogger.Store(logger)

	return logger, nil
}

func SetDefault(logger *Logger) { defaultLogger.Store(logger) }

func Default() *Logger { return defaultLogger.Load() }

func fromDefault() (*Logger, error) {
	if logger := defaultLogger.Load(); logger != nil {
		return logger, nil
	}

	return nil, ErrNoDefault
}

func Log(level Level, message string) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.Log(level, message)
}

func LogContext(ctx context.Context, level Level, message string) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.LogContext(ctx, level, message)
}

func Error(message string) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.Error(message)
}

func Warn(message string) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.Warn(message)
}

func Info(message string) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.Info(message)
}

func Debug(message string) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.Debug(message)
}

func Trace(message string) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.Trace(message)
}

func LogEvent(value Event) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.Event(value)
}

func CreateSpan(value Event) (uint32, error) {
	logger, err := fromDefault()
	if err != nil {
		return 0, err
	}

	return logger.CreateSpan(value)
}

func RecordSpan(id uint32, fields map[string]any) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.RecordSpan(id, fields)
}

func CloseSpan(id uint32) error {
	logger, err := fromDefault()
	if err != nil {
		return err
	}

	return logger.CloseSpan(id)
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

// Close flushes and destroys the default logger, then clears it. A default
// installed afterwards by SetDefault is left untouched.
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
