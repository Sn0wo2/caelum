package caelum

import (
	"github.com/Sn0wo2/caelum/internal/builder"
	"github.com/Sn0wo2/caelum/internal/event"
)

type (
	Logger      = builder.Logger
	StatusError = builder.StatusError
	Event       = event.Event
)

var (
	ErrClosed = builder.ErrClosed
	ErrBusy   = builder.ErrBusy
)

const (
	StatusOK              = builder.StatusOK
	StatusInvalidArgument = builder.StatusInvalidArgument
	StatusInvalidHandle   = builder.StatusInvalidHandle
	StatusLoggerError     = builder.StatusLoggerError
	StatusInternalError   = builder.StatusInternalError
)

func New(cfg Config) (*Logger, error) {
	return builder.New(cfg)
}
