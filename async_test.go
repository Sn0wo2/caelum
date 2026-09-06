package caelum

import (
	"bytes"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// syncBuffer is a goroutine-safe buffer for asserting on async output.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// TestAsyncDeliversWhenSinkKeepsUp checks that with a fast sink and ample
// buffer, async writes reach the underlying writer. Delivery is asynchronous,
// so we poll briefly before asserting.
func TestAsyncDeliversWhenSinkKeepsUp(t *testing.T) {
	var sink syncBuffer
	aw := Async(&sink, AsyncConfig{BufferSize: 4096, PollInterval: time.Millisecond})

	log := New(Config{
		Level:   LevelInfo,
		Targets: []Target{{Writer: aw, Color: NoColor}},
	})

	for i := 0; i < 50; i++ {
		log.Info("async line", "i", i)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Count(sink.String(), "async line") == 50 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if got := strings.Count(sink.String(), "async line"); got != 50 {
		t.Errorf("expected 50 delivered lines, got %d", got)
	}
	if err := log.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if aw.Dropped() != 0 {
		t.Errorf("no drops expected with a fast sink, got %d", aw.Dropped())
	}
}

// TestAsyncNeverBlocks verifies the core guarantee: a wedged sink must not
// block the producer. Writes return promptly even though the worker is stuck.
func TestAsyncNeverBlocks(t *testing.T) {
	block := make(chan struct{})
	bw := blockingWriter{release: block}
	aw := Async(bw, AsyncConfig{BufferSize: 1})
	defer func() {
		close(block)
		_ = aw.Close()
	}()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			_, _ = aw.Write([]byte("y\n"))
		}
		close(done)
	}()

	select {
	case <-done:
		// returned promptly — producer was never blocked by the stuck sink
	case <-time.After(2 * time.Second):
		t.Fatal("Write blocked on a wedged sink; async must be non-blocking")
	}
}

// TestAsyncDropsAndAlerts forces overflow against a wedged sink and asserts the
// drop-oldest backpressure surfaces through Dropped and the Alerter callback.
func TestAsyncDropsAndAlerts(t *testing.T) {
	var alerted uint64
	block := make(chan struct{})
	bw := blockingWriter{release: block}
	aw := Async(bw, AsyncConfig{BufferSize: 4, Alerter: func(missed int) {
		atomic.AddUint64(&alerted, uint64(missed))
	}})

	for i := 0; i < 1000; i++ {
		_, _ = aw.Write([]byte("z\n"))
	}

	close(block)
	_ = aw.Close()

	if aw.Dropped() == 0 {
		t.Errorf("expected drops against a wedged sink with a depth-4 buffer")
	}
	if atomic.LoadUint64(&alerted) == 0 {
		t.Errorf("expected the Alerter callback to fire on overflow")
	}
}

type blockingWriter struct{ release chan struct{} }

func (b blockingWriter) Write(p []byte) (int, error) {
	<-b.release
	return len(p), nil
}

// TestAsyncUnwrapForColorDetection ensures detection sees through the async
// layer: a wrapped buffer is not a terminal, so it must detect as NoColor.
func TestAsyncUnwrapForColorDetection(t *testing.T) {
	var sink bytes.Buffer
	aw := Async(&sink, AsyncConfig{})
	defer func() { _ = aw.Close() }()
	if d := detectDepth(aw); d != NoColor {
		t.Errorf("async-wrapped buffer should detect NoColor, got %v", d)
	}
}
