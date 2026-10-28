package caelum

import (
	"github.com/Sn0wo2/caelum/internal/logger"
)

type (
	Logger      = logger.Logger
	Field       = logger.Field
	Status      = logger.Status
	StatusError = logger.StatusError
	Event       = logger.Event
	SpanID      = logger.SpanID
)

var (
	ErrClosed          = logger.ErrClosed
	ErrBusy            = logger.ErrBusy
	ErrInvalidSpan     = logger.ErrInvalidSpan
	ErrSpanClosed      = logger.ErrSpanClosed
	ErrInvalidArgument = logger.ErrInvalidArgument
	ErrInvalidHandle   = logger.ErrInvalidHandle
	ErrNative          = logger.ErrNative
)

const (
	StatusOK              = logger.StatusOK
	StatusInvalidArgument = logger.StatusInvalidArgument
	StatusInvalidHandle   = logger.StatusInvalidHandle
	StatusLoggerError     = logger.StatusLoggerError
	StatusInternalError   = logger.StatusInternalError
)

func New(cfg Config) (*Logger, error) {
	return logger.New(cfg)
}
