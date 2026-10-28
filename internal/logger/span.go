package logger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/Sn0wo2/caelum/internal/config"
)

var (
	ErrInvalidSpan  = errors.New("caelum: invalid span")
	ErrSpanClosed   = errors.New("caelum: span is closed")
	ErrSpanMismatch = errors.New("caelum: span belongs to a different logger")
)

type SpanID uint32

type Span struct {
	mu     sync.Mutex
	logger *Logger
	id     SpanID
	closed bool
}

func (l *Logger) CreateSpan(value Event) (*Span, error) {
	return l.CreateSpanContext(context.Background(), value)
}

func (l *Logger) CreateSpanContext(ctx context.Context, value Event) (*Span, error) {
	value = withParent(ctx, value)
	if err := value.Parent.check(l); err != nil {
		return nil, err
	}

	value = captureEvent(value, l.disableCaller, l.callerSkip, []config.Level{})

	encoded, err := marshalEvent(value)
	if err != nil {
		return nil, fmt.Errorf("%w: encode span: %w", ErrInvalidArgument, err)
	}

	if err := value.Parent.lock(l); err != nil {
		return nil, err
	}

	if err := l.begin(); err != nil {
		value.Parent.unlock()

		return nil, err
	}

	var id uint32

	err = callNative(l.api, "create span", func() int32 {
		return l.api.CreateSpan(l.handle, unsafe.Pointer(unsafe.SliceData(encoded)), uintptr(len(encoded)), unsafe.Pointer(&id))
	})

	runtime.KeepAlive(encoded)
	runtime.KeepAlive(&id)
	l.mu.Unlock()
	value.Parent.unlock()

	if err != nil {
		return nil, err
	}

	return &Span{logger: l, id: SpanID(id)}, nil
}

func (s *Span) ID() SpanID {
	if s == nil {
		return 0
	}

	return s.id
}

func (s *Span) Context(ctx context.Context) context.Context {
	return ContextWithSpan(ctx, s)
}

func (s *Span) Record(fields map[string]any) error {
	if s == nil {
		return ErrInvalidSpan
	}

	if err := s.check(nil); err != nil {
		return err
	}

	if fields == nil {
		fields = map[string]any{}
	}

	encoded, err := json.Marshal(normalizeFields(fields))
	if err != nil {
		return fmt.Errorf("%w: encode span fields: %w", ErrInvalidArgument, err)
	}

	if err := s.lock(nil); err != nil {
		return err
	}

	l := s.logger
	if err := l.begin(); err != nil {
		s.unlock()

		return err
	}

	err = callNative(l.api, "record span", func() int32 {
		return l.api.RecordSpan(l.handle, uint32(s.id), unsafe.Pointer(unsafe.SliceData(encoded)), uintptr(len(encoded)))
	})

	runtime.KeepAlive(encoded)
	l.mu.Unlock()
	s.unlock()

	if err != nil {
		return err
	}

	_, err = l.pump(false)

	return err
}

func (s *Span) Close() error {
	if s == nil {
		return ErrInvalidSpan
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.logger == nil || s.id == 0 {
		return ErrInvalidSpan
	}

	if s.closed {
		return nil
	}

	l := s.logger
	if err := l.begin(); err != nil {
		if errors.Is(err, ErrClosed) {
			s.closed = true

			return nil
		}

		return err
	}

	err := callNative(l.api, "close span", func() int32 {
		return l.api.CloseSpan(l.handle, uint32(s.id))
	})
	l.mu.Unlock()

	if err == nil {
		s.closed = true
	}

	return err
}

func (s *Span) check(owner *Logger) error {
	if err := s.lock(owner); err != nil {
		return err
	}

	s.unlock()

	return nil
}

func (s *Span) lock(owner *Logger) error {
	if s == nil {
		return nil
	}

	s.mu.Lock()

	var err error

	switch {
	case s.logger == nil || s.id == 0:
		err = ErrInvalidSpan
	case owner != nil && s.logger != owner:
		err = ErrSpanMismatch
	case s.closed:
		err = ErrSpanClosed
	}

	if err != nil {
		s.mu.Unlock()
	}

	return err
}

func (s *Span) unlock() {
	if s != nil {
		s.mu.Unlock()
	}
}
