package logger

import (
	"context"
	"fmt"
	"runtime"
	"slices"
	"unicode/utf8"
	"unsafe"

	"github.com/Sn0wo2/caelum/internal/config"
)

func (l *Logger) Log(level config.Level, message string, fields ...Field) error {
	return l.LogContext(context.Background(), level, message, fields...)
}

func (l *Logger) LogContext(ctx context.Context, level config.Level, message string, fields ...Field) error {
	_, err := l.logContext(ctx, level, message, fields...)

	return err
}

func (l *Logger) logContext(ctx context.Context, level config.Level, message string, fields ...Field) (bool, error) {
	code, err := levelCode(level)
	if err != nil {
		return false, err
	}

	var eventFields map[string]any
	if len(fields) != 0 {
		eventFields = make(map[string]any, len(fields))
		for _, field := range fields {
			eventFields[field.Key] = field.Value
		}
	}

	value := withParent(ctx, Event{Level: level, Message: message, Fields: eventFields})
	if len(fields) != 0 || !l.disableCaller || value.Parent != nil ||
		l.stacktrace == nil && level == config.LevelError ||
		slices.Contains(l.stacktrace, level) || !utf8.ValidString(message) {
		return l.event(value)
	}

	if err := l.begin(); err != nil {
		return false, err
	}

	err = callNative(l.api, "log", func() int32 {
		return l.api.Log(l.handle, code, unsafe.Pointer(unsafe.StringData(message)), uintptr(len(message)))
	})

	runtime.KeepAlive(message)
	l.mu.Unlock()

	if err != nil {
		return false, err
	}

	_, err = l.pump(false)

	return true, err
}

func (l *Logger) Error(message string, fields ...Field) error {
	return l.Log(config.LevelError, message, fields...)
}

func (l *Logger) Warn(message string, fields ...Field) error {
	return l.Log(config.LevelWarn, message, fields...)
}

func (l *Logger) Info(message string, fields ...Field) error {
	return l.Log(config.LevelInfo, message, fields...)
}

func (l *Logger) Debug(message string, fields ...Field) error {
	return l.Log(config.LevelDebug, message, fields...)
}

func (l *Logger) Trace(message string, fields ...Field) error {
	return l.Log(config.LevelTrace, message, fields...)
}

func (l *Logger) Event(value Event) error {
	_, err := l.event(value)

	return err
}

func (l *Logger) event(value Event) (bool, error) {
	if err := value.Parent.check(l); err != nil {
		return false, err
	}

	value = captureEvent(value, l.disableCaller, l.callerSkip, l.stacktrace)

	encoded, err := marshalEvent(value)
	if err != nil {
		return false, fmt.Errorf("%w: encode event: %w", ErrInvalidArgument, err)
	}

	if err := value.Parent.lock(l); err != nil {
		return false, err
	}

	if err := l.begin(); err != nil {
		value.Parent.unlock()

		return false, err
	}

	err = callNative(l.api, "event", func() int32 {
		return l.api.Event(l.handle, unsafe.Pointer(unsafe.SliceData(encoded)), uintptr(len(encoded)))
	})

	runtime.KeepAlive(encoded)
	l.mu.Unlock()
	value.Parent.unlock()

	if err != nil {
		return false, err
	}

	_, err = l.pump(false)

	return true, err
}

func (l *Logger) SetFilter(filter string) error {
	if err := l.begin(); err != nil {
		return err
	}
	defer l.mu.Unlock()

	err := callNative(l.api, "set filter", func() int32 {
		return l.api.SetFilter(l.handle, unsafe.Pointer(unsafe.StringData(filter)), uintptr(len(filter)))
	})

	runtime.KeepAlive(filter)

	return err
}

func (l *Logger) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	accepted, err := l.logContext(context.Background(), config.LevelInfo, string(p))
	if accepted {
		return len(p), err
	}

	return 0, err
}

func levelCode(level config.Level) (uint32, error) {
	switch level {
	case config.LevelError:
		return 0, nil
	case config.LevelWarn:
		return 1, nil
	case config.LevelInfo:
		return 2, nil
	case config.LevelDebug:
		return 3, nil
	case config.LevelTrace:
		return 4, nil
	default:
		return 0, fmt.Errorf("%w: invalid log level %q", ErrInvalidArgument, level)
	}
}
