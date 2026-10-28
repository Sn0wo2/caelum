package builder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"slices"
	"sync"
	"unsafe"

	"github.com/Sn0wo2/caelum/internal/config"
	"github.com/Sn0wo2/caelum/internal/event"
	"github.com/Sn0wo2/caelum/internal/native"
	"github.com/Sn0wo2/caelum/internal/span"
)

var (
	ErrClosed = errors.New("caelum: logger is closed")
	ErrBusy   = errors.New("caelum: logger is already draining")
)

const (
	StatusOK = iota
	StatusInvalidArgument
	StatusInvalidHandle
	StatusLoggerError
	StatusInternalError
)

var statusNames = [...]string{StatusOK: "ok", StatusInvalidArgument: "invalid argument", StatusInvalidHandle: "invalid handle", StatusLoggerError: "logger error", StatusInternalError: "internal error"}

type StatusError struct {
	Operation string
	Status    int32
	Detail    string
}

func (e *StatusError) Error() string {
	name := "unknown"
	if e.Status >= 0 && int(e.Status) < len(statusNames) {
		name = statusNames[e.Status]
	}

	if e.Detail == "" {
		return fmt.Sprintf("caelum: %s failed with status %d (%s)", e.Operation, e.Status, name)
	}

	return fmt.Sprintf("caelum: %s failed with status %d (%s): %s", e.Operation, e.Status, name, e.Detail)
}

type Logger struct {
	mu      sync.Mutex
	api     *native.API
	handle  uint64
	writer  io.Writer
	buffer  []byte
	pending []byte
	pumping bool
	closing bool
	closed  bool

	disableCaller bool
	callerSkip    int
	stacktrace    []config.Level
	embedded      bool
}

func New(cfg config.Config) (*Logger, error) {
	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}

	if cfg.CallerSkip < 0 {
		return nil, errors.New("caelum: caller skip cannot be negative")
	}

	for _, level := range cfg.Stacktrace {
		if level > config.LevelTrace {
			return nil, fmt.Errorf("caelum: invalid stacktrace level %d", level)
		}
	}

	capacity := uint(4 << 20)
	if cfg.Capacity != nil {
		capacity = *cfg.Capacity
	}

	bufferedDepth := config.ColorDepthTrueColor
	if os.Getenv("NO_COLOR") != "" {
		bufferedDepth = config.ColorDepthNoColor
	}

	var nativeConfig []byte

	if cfg.NativeConfig != nil {
		if !json.Valid(cfg.NativeConfig) {
			return nil, errors.New("caelum: native config is not valid JSON")
		}

		nativeConfig = slices.Clone(cfg.NativeConfig)
	} else {
		if cfg.Writers == nil {
			cfg.Writers = []config.Writer{{ColorDepth: &bufferedDepth, Stacktrace: []config.Level{}, Target: config.WriterTarget{Buffered: true}}}
		} else {
			cfg.Writers = slices.Clone(cfg.Writers)
			for i := range cfg.Writers {
				if cfg.Writers[i].Stacktrace == nil {
					cfg.Writers[i].Stacktrace = []config.Level{}
				}

				if cfg.Writers[i].ColorDepth == nil && cfg.Writers[i].Target.Buffered {
					cfg.Writers[i].ColorDepth = &bufferedDepth
				}
			}
		}

		var err error

		nativeConfig, err = json.Marshal(cfg)
		if err != nil {
			return nil, fmt.Errorf("caelum: encode native config: %w", err)
		}
	}

	api, err := native.Load(cfg.Loader)
	if err != nil {
		return nil, err
	}

	var handle uint64

	status := api.Create(unsafe.Pointer(unsafe.SliceData(nativeConfig)), uintptr(len(nativeConfig)), uintptr(capacity), unsafe.Pointer(&handle))
	runtime.KeepAlive(nativeConfig)
	runtime.KeepAlive(&handle)

	if status != 0 {
		return nil, &StatusError{Operation: "create", Status: status, Detail: api.LastError()}
	}

	return &Logger{
		api: api, handle: handle, writer: cfg.Writer, buffer: make([]byte, 32*1024),
		disableCaller: cfg.DisableCaller, callerSkip: cfg.CallerSkip,
		stacktrace: slices.Clone(cfg.Stacktrace),
		embedded:   cfg.Loader == nil,
	}, nil
}

// ActaVersion reports the version of the acta runtime backing this logger. It
// queries the loaded library when it exports acta_version and falls back to
// the embedded build stamp; "" means the version is unknown.
func (l *Logger) ActaVersion() string {
	if version := l.api.ActaVersion(); version != "" {
		return version
	}

	if l.embedded {
		return native.EmbeddedActaVersion
	}

	return ""
}

func (l *Logger) Log(level config.Level, message string) error {
	return l.LogContext(context.Background(), level, message)
}

func (l *Logger) LogContext(ctx context.Context, level config.Level, message string) error {
	if level > config.LevelTrace {
		return fmt.Errorf("caelum: invalid log level %d", level)
	}

	value, err := span.WithParent(ctx, l, event.Event{Level: &level, Message: message})
	if err != nil {
		return err
	}

	if !l.disableCaller || value.Parent != nil || l.stacktrace == nil && level == config.LevelError || slices.Contains(l.stacktrace, level) {
		return l.Event(value)
	}

	if err := l.begin(); err != nil {
		return err
	}

	status := l.api.Log(l.handle, uint32(level), unsafe.Pointer(unsafe.StringData(message)), uintptr(len(message)))
	runtime.KeepAlive(message)

	if status != 0 {
		l.mu.Unlock()

		return l.fail("log", status)
	}
	l.mu.Unlock()

	_, err = l.pump(false)

	return err
}

func (l *Logger) Error(message string) error { return l.Log(config.LevelError, message) }
func (l *Logger) Warn(message string) error  { return l.Log(config.LevelWarn, message) }
func (l *Logger) Info(message string) error  { return l.Log(config.LevelInfo, message) }
func (l *Logger) Debug(message string) error { return l.Log(config.LevelDebug, message) }
func (l *Logger) Trace(message string) error { return l.Log(config.LevelTrace, message) }

func (l *Logger) Event(value event.Event) error {
	value = event.Capture(value, l.disableCaller, l.callerSkip, l.stacktrace)

	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("caelum: encode event: %w", err)
	}

	if err := l.begin(); err != nil {
		return err
	}

	status := l.api.Event(l.handle, unsafe.Pointer(unsafe.SliceData(encoded)), uintptr(len(encoded)))
	runtime.KeepAlive(encoded)

	if status != 0 {
		l.mu.Unlock()

		return l.fail("event", status)
	}
	l.mu.Unlock()

	_, err = l.pump(false)

	return err
}

func (l *Logger) CreateSpan(value event.Event) (uint32, error) {
	value = event.Capture(value, l.disableCaller, l.callerSkip, []config.Level{})

	encoded, err := json.Marshal(value)
	if err != nil {
		return 0, fmt.Errorf("caelum: encode event: %w", err)
	}

	if err := l.begin(); err != nil {
		return 0, err
	}

	var id uint32

	status := l.api.CreateSpan(l.handle, unsafe.Pointer(unsafe.SliceData(encoded)), uintptr(len(encoded)), unsafe.Pointer(&id))
	runtime.KeepAlive(encoded)
	runtime.KeepAlive(&id)

	if status != 0 {
		l.mu.Unlock()

		return 0, l.fail("create span", status)
	}
	l.mu.Unlock()

	return id, nil
}

func (l *Logger) RecordSpan(id uint32, fields map[string]any) error {
	if fields == nil {
		fields = map[string]any{}
	}

	encoded, err := json.Marshal(fields)
	if err != nil {
		return fmt.Errorf("caelum: encode span fields: %w", err)
	}

	if err := l.begin(); err != nil {
		return err
	}

	status := l.api.RecordSpan(l.handle, id, unsafe.Pointer(unsafe.SliceData(encoded)), uintptr(len(encoded)))
	runtime.KeepAlive(encoded)

	if status != 0 {
		l.mu.Unlock()

		return l.fail("record span", status)
	}
	l.mu.Unlock()

	_, err = l.pump(false)

	return err
}

func (l *Logger) CloseSpan(id uint32) error {
	if err := l.begin(); err != nil {
		return err
	}
	defer l.mu.Unlock()

	status := l.api.CloseSpan(l.handle, id)
	if status != 0 {
		return l.fail("close span", status)
	}

	return nil
}

func (l *Logger) SetFilter(filter string) error {
	if err := l.begin(); err != nil {
		return err
	}
	defer l.mu.Unlock()

	status := l.api.SetFilter(l.handle, unsafe.Pointer(unsafe.StringData(filter)), uintptr(len(filter)))
	runtime.KeepAlive(filter)

	if status != 0 {
		return l.fail("set filter", status)
	}

	return nil
}

func (l *Logger) Dropped() (uint64, error) {
	if err := l.begin(); err != nil {
		return 0, err
	}
	defer l.mu.Unlock()

	var (
		written uintptr
		dropped uint64
	)

	status := l.api.Drain(l.handle, nil, 0, unsafe.Pointer(&written), unsafe.Pointer(&dropped))
	runtime.KeepAlive(&written)
	runtime.KeepAlive(&dropped)

	if status != 0 {
		return 0, l.fail("dropped", status)
	}

	return dropped, nil
}

func (l *Logger) Flush() error {
	if err := l.begin(); err != nil {
		return err
	}
	l.mu.Unlock()

	owned, err := l.pump(true)
	if err != nil {
		return err
	}

	if !owned {
		return ErrBusy
	}

	return nil
}

func (l *Logger) Close() error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()

		return nil
	}

	if l.closing || l.pumping {
		l.mu.Unlock()

		return ErrBusy
	}

	l.closing = true
	l.mu.Unlock()

	owned, err := l.pump(true)
	if err != nil || !owned {
		l.mu.Lock()
		l.closing = false
		l.mu.Unlock()

		if err != nil {
			return err
		}

		return ErrBusy
	}

	l.mu.Lock()

	status := l.api.Destroy(l.handle)
	if status != 0 {
		l.closing = false
		l.mu.Unlock()

		return l.fail("destroy", status)
	}

	l.closed = true
	l.closing = false
	l.pending = nil
	l.buffer = nil
	l.mu.Unlock()

	return nil
}

func (l *Logger) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	if err := l.Info(string(p)); err != nil {
		return 0, err
	}

	return len(p), nil
}

// begin locks the logger for a new operation and reports why the logger must
// not be used. On success the caller owns l.mu and must unlock it.
func (l *Logger) begin() error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()

		return ErrClosed
	}

	if l.closing {
		l.mu.Unlock()

		return ErrBusy
	}

	return nil
}

// fail builds the error for a failed acta call, capturing the detail message
// recorded for it on this thread.
func (l *Logger) fail(operation string, status int32) error {
	return &StatusError{Operation: operation, Status: status, Detail: l.api.LastError()}
}

func (l *Logger) drainLocked() error {
	var (
		written uintptr
		dropped uint64
	)

	status := l.api.Drain(l.handle, unsafe.Pointer(unsafe.SliceData(l.buffer)), uintptr(len(l.buffer)), unsafe.Pointer(&written), unsafe.Pointer(&dropped))
	runtime.KeepAlive(l.buffer)
	runtime.KeepAlive(&written)

	if status != 0 {
		return l.fail("drain", status)
	}

	if written > uintptr(len(l.buffer)) {
		return fmt.Errorf("caelum: drain returned invalid byte count %d", written)
	}

	if written != 0 {
		l.pending = l.buffer[:int(written)]
	}

	return nil
}

func (l *Logger) pump(flushWriter bool) (bool, error) {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()

		return false, ErrClosed
	}

	if l.pumping {
		l.mu.Unlock()

		return false, nil
	}

	l.pumping = true
	l.mu.Unlock()

	return true, l.pumpLoop(flushWriter)
}

func (l *Logger) pumpLoop(flushWriter bool) error {
	defer func() {
		l.mu.Lock()
		l.pumping = false
		l.mu.Unlock()
	}()

	for {
		l.mu.Lock()
		if len(l.pending) == 0 {
			if err := l.drainLocked(); err != nil {
				l.mu.Unlock()

				return err
			}

			if len(l.pending) != 0 {
				l.mu.Unlock()

				continue
			}

			if !flushWriter {
				l.mu.Unlock()

				return nil
			}
			l.mu.Unlock()

			if writer, ok := l.writer.(interface{ Flush() error }); ok {
				if err := writer.Flush(); err != nil {
					return err
				}
			}

			l.mu.Lock()
			if err := l.drainLocked(); err != nil {
				l.mu.Unlock()

				return err
			}

			if len(l.pending) == 0 {
				l.mu.Unlock()

				return nil
			}
			l.mu.Unlock()

			continue
		}

		chunk := l.pending
		l.mu.Unlock()

		written, err := l.writer.Write(chunk)
		if written < 0 || written > len(chunk) {
			err = fmt.Errorf("caelum: invalid write count %d", written)
			written = 0
		} else if err == nil && written != len(chunk) {
			err = io.ErrShortWrite
		}

		l.mu.Lock()

		l.pending = l.pending[written:]
		if len(l.pending) == 0 {
			l.pending = nil
		}

		if err != nil {
			l.mu.Unlock()

			return err
		}
		l.mu.Unlock()
	}
}
