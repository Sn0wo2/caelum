package caelum

import (
	"io"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/diode"
)

// AsyncWriter is a non-blocking decorator around any io.Writer, backed by
// rs/zerolog's lock-free diode: writes land on a ring buffer drained by a
// background goroutine, so a slow sink never blocks the logging goroutine.
//
// Backpressure is drop-oldest — when producers outrun the sink the oldest
// queued records are discarded rather than blocking the caller; observe loss
// via Dropped or WithAlerter. Records still queued at Close are not guaranteed
// to flush (a diode property), so drain-critical callers must pause logging
// first. Implements io.WriteCloser; Close stops the worker and closes the
// underlying writer.
type AsyncWriter struct {
	w       io.Writer
	d       diode.Writer
	dropped uint64
}

// AsyncConfig configures an AsyncWriter.
type AsyncConfig struct {
	BufferSize   int
	PollInterval time.Duration
	Alerter      func(missed int)
}

// Async wraps w so writes are performed asynchronously by a background worker.
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

// Write enqueues p on the diode and returns immediately; it never blocks and
// never errors. Records are dropped (oldest first) if the worker falls behind.
func (a *AsyncWriter) Write(p []byte) (int, error) {
	return a.d.Write(p)
}

// Close stops the worker and closes the underlying writer if it implements
// io.Closer. Note: records still queued at Close time may not be flushed (a
// diode property); drain-critical callers should pause logging beforehand.
func (a *AsyncWriter) Close() error {
	return a.d.Close()
}

// Dropped reports the total number of records discarded due to buffer overflow.
func (a *AsyncWriter) Dropped() uint64 {
	return atomic.LoadUint64(&a.dropped)
}

// Unwrap exposes the underlying writer so color detection can see through the
// async layer to the real terminal.
func (a *AsyncWriter) Unwrap() io.Writer { return a.w }
