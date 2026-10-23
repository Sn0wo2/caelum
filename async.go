package caelum

import (
	"errors"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type AsyncWriter struct {
	w            io.Writer
	queue        chan asyncRecord
	wake         chan struct{}
	mu           sync.Mutex
	requests     []*asyncRequest
	accepted     uint64
	writeErr     error
	closed       bool
	closeErr     error
	done         chan struct{}
	dropped      atomic.Uint64
	pollInterval time.Duration
	alerter      func(missed int)
}

type asyncRecord struct {
	seq  uint64
	data []byte
}

type asyncRequest struct {
	target  uint64
	closing bool
	result  chan error
}

type AsyncConfig struct {
	BufferSize   int
	PollInterval time.Duration
	Alerter      func(missed int)
}

func Async(w io.Writer, cfg AsyncConfig) *AsyncWriter {
	if w == nil {
		w = os.Stdout
	}

	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 1024
	}

	a := &AsyncWriter{
		w:            w,
		queue:        make(chan asyncRecord, cfg.BufferSize),
		wake:         make(chan struct{}, 1),
		done:         make(chan struct{}),
		pollInterval: cfg.PollInterval,
		alerter:      cfg.Alerter,
	}
	go a.run()

	return a
}

func (a *AsyncWriter) Write(p []byte) (int, error) {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()

		return 0, io.ErrClosedPipe
	}

	if len(a.queue) == cap(a.queue) {
		a.dropped.Add(1)
		a.mu.Unlock()

		if a.alerter != nil {
			a.alerter(1)
		}

		return len(p), nil
	}

	data := append([]byte(nil), p...)

	a.accepted++
	a.queue <- asyncRecord{seq: a.accepted, data: data}

	if a.pollInterval <= 0 {
		a.signalLocked()
	}
	a.mu.Unlock()

	return len(p), nil
}

func (a *AsyncWriter) Close() error {
	a.mu.Lock()
	if !a.closed {
		a.closed = true
		a.requests = append(a.requests, &asyncRequest{target: a.accepted, closing: true})
		a.signalLocked()
	}

	done := a.done
	a.mu.Unlock()

	<-done
	a.mu.Lock()
	err := a.closeErr
	a.mu.Unlock()

	return err
}

func (a *AsyncWriter) Flush() error {
	a.mu.Lock()
	if a.closed {
		done := a.done
		a.mu.Unlock()
		<-done
		a.mu.Lock()
		err := a.closeErr
		a.mu.Unlock()

		return err
	}

	req := &asyncRequest{target: a.accepted, result: make(chan error, 1)}
	a.requests = append(a.requests, req)
	a.signalLocked()
	a.mu.Unlock()

	return <-req.result
}

func (a *AsyncWriter) Dropped() uint64 {
	return a.dropped.Load()
}

func (a *AsyncWriter) Unwrap() io.Writer { return a.w }

func (a *AsyncWriter) signalLocked() {
	select {
	case a.wake <- struct{}{}:
	default:
	}
}

func (a *AsyncWriter) nextWork(allowRecord bool) (*asyncRequest, asyncRecord, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.requests) != 0 {
		req := a.requests[0]
		if len(a.requests) == 1 {
			a.requests = nil
		} else {
			a.requests[0] = nil
			a.requests = a.requests[1:]
		}

		return req, asyncRecord{}, false
	}

	if allowRecord {
		select {
		case record := <-a.queue:
			return nil, record, true
		default:
		}
	}

	return nil, asyncRecord{}, false
}

func (a *AsyncWriter) write(record asyncRecord) {
	n, err := a.w.Write(record.data)
	if err == nil && n != len(record.data) {
		err = io.ErrShortWrite
	}

	if a.writeErr == nil && err != nil {
		a.writeErr = err
	}
}

func (a *AsyncWriter) flush() error {
	var flushErr error
	if w, ok := a.w.(interface{ Flush() error }); ok {
		flushErr = w.Flush()
	}

	return errors.Join(a.writeErr, flushErr)
}

func (a *AsyncWriter) run() {
	defer close(a.done)

	var (
		ticker *time.Ticker
		tick   <-chan time.Time
	)

	polling := a.pollInterval > 0
	if polling {
		ticker = time.NewTicker(a.pollInterval)
		tick = ticker.C

		defer ticker.Stop()
	}

	var completed uint64

	pollReady := true
	for {
		req, record, haveRecord := a.nextWork(pollReady)
		if req != nil {
			for completed < req.target {
				record := <-a.queue
				a.write(record)
				completed = record.seq
			}

			err := a.flush()
			if req.closing {
				a.mu.Lock()
				a.closeErr = err
				a.mu.Unlock()

				return
			}

			req.result <- err

			continue
		}

		if haveRecord {
			a.write(record)
			completed = record.seq

			continue
		}

		if polling && pollReady {
			pollReady = false
		}

		if polling {
			select {
			case <-a.wake:
			case <-tick:
				pollReady = true
			}
		} else {
			<-a.wake
		}
	}
}
