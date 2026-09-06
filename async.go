package caelum

import (
	"io"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/diode"
)

type AsyncWriter struct {
	w       io.Writer
	d       diode.Writer
	dropped uint64
}

type AsyncConfig struct {
	BufferSize   int
	PollInterval time.Duration
	Alerter      func(missed int)
}

func Async(w io.Writer, cfg AsyncConfig) *AsyncWriter {
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 1024
	}

	a := &AsyncWriter{w: w}
	alert := func(missed int) {
		atomic.AddUint64(&a.dropped, uint64(missed))
		if cfg.Alerter != nil {
			cfg.Alerter(missed)
		}
	}
	a.d = diode.NewWriter(w, cfg.BufferSize, cfg.PollInterval, alert)
	return a
}

func (a *AsyncWriter) Write(p []byte) (int, error) {
	return a.d.Write(p)
}

func (a *AsyncWriter) Close() error {
	return a.d.Close()
}

func (a *AsyncWriter) Dropped() uint64 {
	return atomic.LoadUint64(&a.dropped)
}

func (a *AsyncWriter) Unwrap() io.Writer { return a.w }
