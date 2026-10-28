package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"slices"
	"sync"
	"unsafe"

	"github.com/Sn0wo2/caelum/internal/config"
	"github.com/Sn0wo2/caelum/internal/native"
)

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
	asyncOutput   bool
}

func New(cfg config.Config) (*Logger, error) {
	if cfg.CallerSkip < 0 {
		return nil, fmt.Errorf("%w: caller skip cannot be negative", ErrInvalidArgument)
	}

	if cfg.Capacity < 0 {
		return nil, fmt.Errorf("%w: capacity cannot be negative", ErrInvalidArgument)
	}

	for _, level := range cfg.Stacktrace {
		if !level.Valid() {
			return nil, fmt.Errorf("%w: invalid stacktrace level %q", ErrInvalidArgument, level)
		}
	}

	capacity := cfg.Capacity
	if capacity == 0 {
		capacity = 4 << 20
	}

	var nativeConfig []byte

	if cfg.NativeConfig != nil {
		if !json.Valid(cfg.NativeConfig) {
			return nil, fmt.Errorf("%w: native config is not valid JSON", ErrInvalidArgument)
		}

		nativeConfig = slices.Clone(cfg.NativeConfig)
	} else {
		if cfg.Writers == nil {
			cfg.Writers = []config.Writer{config.DefaultWriter()}
		} else {
			cfg.Writers = slices.Clone(cfg.Writers)
		}

		for i := range cfg.Writers {
			if cfg.Writers[i].Target == "" {
				cfg.Writers[i].Target = config.WriterTargetBuffered
			}

			if cfg.Writers[i].Stacktrace == nil {
				cfg.Writers[i].Stacktrace = []config.Level{}
			}
		}

		var err error

		nativeConfig, err = json.Marshal(cfg)
		if err != nil {
			return nil, fmt.Errorf("%w: encode native config: %w", ErrInvalidArgument, err)
		}
	}

	buffered, asyncOutput := inspectOutputs(nativeConfig)
	if cfg.Writer != nil && !buffered {
		return nil, fmt.Errorf("%w: Go writer requires a buffered writer target", ErrInvalidArgument)
	}

	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}

	api, err := native.Load(cfg.Loader)
	if err != nil {
		return nil, err
	}

	var handle uint64

	err = callNative(api, "create", func() int32 {
		return api.Create(
			unsafe.Pointer(unsafe.SliceData(nativeConfig)),
			uintptr(len(nativeConfig)),
			uintptr(capacity),
			unsafe.Pointer(&handle),
		)
	})

	runtime.KeepAlive(nativeConfig)
	runtime.KeepAlive(&handle)

	if err != nil {
		return nil, err
	}

	return &Logger{
		api:           api,
		handle:        handle,
		writer:        cfg.Writer,
		buffer:        make([]byte, 32*1024),
		disableCaller: cfg.DisableCaller,
		callerSkip:    cfg.CallerSkip,
		stacktrace:    slices.Clone(cfg.Stacktrace),
		embedded:      cfg.Loader == nil,
		asyncOutput:   asyncOutput,
	}, nil
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

	err = callNative(l.api, "destroy", func() int32 { return l.api.Destroy(l.handle) })
	if err != nil {
		l.closing = false
		l.mu.Unlock()

		return err
	}

	l.closed = true
	l.closing = false
	l.pending = nil
	l.buffer = nil
	l.mu.Unlock()

	return nil
}

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
