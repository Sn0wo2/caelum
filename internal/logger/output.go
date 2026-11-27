package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"runtime"
	"unsafe"

	"github.com/Sn0wo2/caelum/internal/config"
	"github.com/Sn0wo2/caelum/internal/native"
)

func (l *Logger) Dropped() (uint64, error) {
	if err := l.begin(); err != nil {
		return 0, err
	}
	defer l.mu.Unlock()

	var (
		written uintptr
		dropped uint64
	)

	err := callNative(l.api, "dropped", func() int32 {
		return l.api.Drain(l.handle, nil, 0, unsafe.Pointer(&written), unsafe.Pointer(&dropped))
	})

	runtime.KeepAlive(&written)
	runtime.KeepAlive(&dropped)

	if err != nil {
		return 0, err
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

	if l.asyncOutput {
		return fmt.Errorf("caelum: native file/async outputs require Close; acta ABI %d cannot flush them: %w", native.ABIVersion, errors.ErrUnsupported)
	}

	return nil
}

func inspectOutputs(data []byte) (buffered, async bool) {
	var cfg struct {
		Writers []struct {
			Target any `json:"target"`
		} `json:"writers"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return true, false
	}

	if cfg.Writers == nil {
		return true, false
	}

	for _, writer := range cfg.Writers {
		if target, ok := writer.Target.(string); ok && target == string(config.WriterTargetBuffered) {
			buffered = true
		}

		target, ok := writer.Target.(map[string]any)
		if !ok {
			continue
		}

		_, file := target["file"]
		_, stdout := target["asyncstdout"]

		_, stderr := target["asyncstderr"]
		if file || stdout || stderr {
			async = true
		}
	}

	return buffered, async
}

func (l *Logger) drainLocked() error {
	var (
		written uintptr
		dropped uint64
	)

	err := callNative(l.api, "drain", func() int32 {
		return l.api.Drain(l.handle, unsafe.Pointer(unsafe.SliceData(l.buffer)), uintptr(len(l.buffer)), unsafe.Pointer(&written), unsafe.Pointer(&dropped))
	})
	runtime.KeepAlive(l.buffer)
	runtime.KeepAlive(&written)
	runtime.KeepAlive(&dropped)

	if err != nil {
		return err
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
	locked := true

	defer func() {
		if !locked {
			l.mu.Lock()
		}

		l.pumping = false
		l.mu.Unlock()
	}()

	for {
		if len(l.pending) == 0 {
			if err := l.drainLocked(); err != nil {
				return true, err
			}

			if len(l.pending) == 0 {
				if !flushWriter {
					return true, nil
				}

				locked = false
				l.mu.Unlock()

				var err error
				if writer, ok := l.writer.(interface{ Flush() error }); ok {
					err = writer.Flush()
				}

				l.mu.Lock()
				locked = true

				if err != nil {
					return true, err
				}

				if err := l.drainLocked(); err != nil {
					return true, err
				}

				if len(l.pending) == 0 {
					return true, nil
				}
			}
		}

		chunk := l.pending
		locked = false
		l.mu.Unlock()

		written, err := l.writer.Write(chunk)

		l.mu.Lock()
		locked = true

		if written < 0 || written > len(chunk) {
			err = fmt.Errorf("caelum: invalid write count %d", written)
			written = 0
		} else if err == nil && written != len(chunk) {
			err = io.ErrShortWrite
		}

		l.pending = l.pending[written:]
		if len(l.pending) == 0 {
			l.pending = nil
		}

		if err != nil {
			return true, err
		}
	}
}
