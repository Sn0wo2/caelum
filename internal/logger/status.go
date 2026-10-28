package logger

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/Sn0wo2/caelum/internal/native"
)

var (
	ErrClosed          = errors.New("caelum: logger is closed")
	ErrBusy            = errors.New("caelum: logger is already draining")
	ErrInvalidArgument = errors.New("caelum: invalid argument")
	ErrInvalidHandle   = errors.New("caelum: invalid logger handle")
	ErrNative          = errors.New("caelum: native runtime error")
)

type Status int32

const (
	StatusOK Status = iota
	StatusInvalidArgument
	StatusInvalidHandle
	StatusLoggerError
	StatusInternalError
)

func (status Status) String() string {
	switch status {
	case StatusOK:
		return "ok"
	case StatusInvalidArgument:
		return "invalid argument"
	case StatusInvalidHandle:
		return "invalid handle"
	case StatusLoggerError:
		return "logger error"
	case StatusInternalError:
		return "internal error"
	default:
		return fmt.Sprintf("status(%d)", int32(status))
	}
}

type StatusError struct {
	Operation string
	Status    Status
	Detail    string
}

func (e *StatusError) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("caelum: %s failed with status %d (%s)", e.Operation, e.Status, e.Status.String())
	}

	return fmt.Sprintf("caelum: %s failed with status %d (%s): %s", e.Operation, e.Status, e.Status.String(), e.Detail)
}

func (e *StatusError) Unwrap() error {
	switch e.Status {
	case StatusOK:
		return nil
	case StatusInvalidArgument:
		return ErrInvalidArgument
	case StatusInvalidHandle:
		return ErrInvalidHandle
	case StatusLoggerError, StatusInternalError:
		return ErrNative
	default:
		return ErrNative
	}
}

func callNative(api *native.API, operation string, invoke func() int32) error {
	runtime.LockOSThread()

	defer runtime.UnlockOSThread()

	status := Status(invoke())
	if status != StatusOK {
		return &StatusError{Operation: operation, Status: status, Detail: api.LastError()}
	}

	return nil
}
