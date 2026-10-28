package caelum

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
)

func testLogger(t *testing.T, cfg Config) *Logger {
	t.Helper()

	logger, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := logger.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})

	return logger
}

func TestSynchronousLoggingAndFilter(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	logger := testLogger(t, Config{Writer: &output, Filter: "warn"})

	if err := logger.Info("filtered"); err != nil {
		t.Fatal(err)
	}

	if err := logger.Warn("visible"); err != nil {
		t.Fatal(err)
	}

	if got := output.String(); strings.Contains(got, "filtered") || !strings.Contains(got, "visible") {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestSetFilter(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	logger := testLogger(t, Config{Writer: &output, Filter: "info"})

	if err := logger.SetFilter("error"); err != nil {
		t.Fatal(err)
	}

	if err := logger.Warn("filtered"); err != nil {
		t.Fatal(err)
	}

	if err := logger.Error("visible"); err != nil {
		t.Fatal(err)
	}

	if got := output.String(); strings.Contains(got, "filtered") || !strings.Contains(got, "visible") {
		t.Fatalf("unexpected output after changing filter: %q", got)
	}
}

func TestJSONFormat(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	logger := testLogger(t, Config{Writer: &output, Writers: []Writer{{
		Format: Format{Json: &JsonOptions{FlattenEvent: true}},
		Target: WriterTarget{Buffered: true},
	}}})

	if err := logger.Info("json event"); err != nil {
		t.Fatal(err)
	}

	if got := output.String(); !strings.HasPrefix(got, "{") || !strings.Contains(got, "json event") {
		t.Fatalf("expected JSON event, got %q", got)
	}
}

func TestCapacityCountsDroppedEvents(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	logger := testLogger(t, Config{Writer: &output, Capacity: new(uint(1))})

	if err := logger.Info("too large for the queue"); err != nil {
		t.Fatal(err)
	}

	dropped, err := logger.Dropped()
	if err != nil {
		t.Fatal(err)
	}

	if dropped != 1 || output.Len() != 0 {
		t.Fatalf("dropped = %d, output = %q", dropped, output.String())
	}
}

func TestPartialWriteCanBeRetried(t *testing.T) {
	t.Parallel()

	output := &retryWriter{fail: true}
	logger := testLogger(t, Config{Writer: output})

	var nestedErr error

	output.duringWrite = func() {
		nestedErr = logger.Info("later event")
	}

	if err := logger.Info("retry me"); !errors.Is(err, errTemporary) {
		t.Fatalf("Log error = %v, want temporary writer error", err)
	}

	if nestedErr != nil {
		t.Fatalf("reentrant Log: %v", nestedErr)
	}

	if err := logger.Flush(); err != nil {
		t.Fatal(err)
	}

	got := output.String()
	if strings.Count(got, "retry me") != 1 || strings.Index(got, "retry me") > strings.Index(got, "later event") {
		t.Fatalf("partial write was not retained ahead of newer logs: %q", got)
	}
}

func TestNativeQueueBoundsLogsDuringWriterWrite(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})

	var releaseOnce sync.Once

	output := &blockingWriter{started: make(chan struct{}), release: release}
	logger := testLogger(t, Config{Writer: output, Capacity: new(uint(8192))})

	defer releaseOnce.Do(func() { close(release) })

	firstDone := make(chan error, 1)
	go func() { firstDone <- logger.Info("in flight") }()

	<-output.started

	for range 100 {
		if err := logger.Info("queued event"); err != nil {
			t.Fatal(err)
		}
	}

	dropped, err := logger.Dropped()
	if err != nil {
		t.Fatal(err)
	}

	if dropped == 0 {
		t.Fatal("expected the bounded native queue to drop events while the writer is blocked")
	}

	releaseOnce.Do(func() { close(release) })

	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}

	if err := logger.Flush(); err != nil {
		t.Fatal(err)
	}

	if got := output.String(); !strings.Contains(got, "in flight") {
		t.Fatalf("in-flight event was not delivered: %q", got)
	}
}

func TestReentrantWriterDoesNotDeadlockOrReorder(t *testing.T) {
	t.Parallel()

	output := new(reentrantWriter)
	logger := testLogger(t, Config{Writer: output})
	output.logger = logger

	if err := logger.Info("outer"); err != nil {
		t.Fatal(err)
	}

	if err := logger.Flush(); err != nil {
		t.Fatal(err)
	}

	got := output.String()

	outer, nested := strings.Index(got, "outer"), strings.Index(got, "nested")
	if outer < 0 || nested < 0 || outer > nested {
		t.Fatalf("reentrant output is missing or reordered: %q", got)
	}
}

func TestWriterInterfaceAndClose(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	logger := testLogger(t, Config{Writer: &output})

	if n, err := logger.Write([]byte("writer message")); err != nil || n != len("writer message") {
		t.Fatalf("Write() = %d, %v", n, err)
	}

	if err := logger.Close(); err != nil {
		t.Fatal(err)
	}

	if err := logger.Info("after close"); !errors.Is(err, ErrClosed) {
		t.Fatalf("Log after Close() = %v, want ErrClosed", err)
	}

	if got := output.String(); !strings.Contains(got, "writer message") {
		t.Fatalf("writer message was not delivered: %q", got)
	}
}

func TestConcurrentLogging(t *testing.T) {
	t.Parallel()

	output := &lockedBuffer{}
	logger := testLogger(t, Config{Writer: output, Filter: "trace"})

	var (
		group    sync.WaitGroup
		errorsMu sync.Mutex
		errs     []error
	)

	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()

			for range 20 {
				if err := logger.Info("concurrent"); err != nil {
					errorsMu.Lock()

					errs = append(errs, err)
					errorsMu.Unlock()
				}
			}
		}()
	}

	group.Wait()

	for {
		err := logger.Flush()
		if !errors.Is(err, ErrBusy) {
			if err != nil {
				t.Fatal(err)
			}

			break
		}
	}

	if len(errs) != 0 {
		t.Fatalf("concurrent logging errors: %v", errs)
	}

	if got := strings.Count(output.String(), "concurrent"); got != 160 {
		t.Fatalf("got %d concurrent events, want 160", got)
	}
}

func TestInvalidConfig(t *testing.T) {
	t.Parallel()

	if _, err := New(Config{CallerSkip: -1}); err == nil {
		t.Fatal("negative capacity should be rejected")
	}

	if _, err := New(Config{Writers: []Writer{{Format: Format{Compact: &Formatter{}, Json: &JsonOptions{}}}}}); err == nil {
		t.Fatal("invalid format should be rejected")
	}
}

var errTemporary = errors.New("temporary write failure")

type retryWriter struct {
	bytes.Buffer
	fail        bool
	duringWrite func()
}

func (w *retryWriter) Write(p []byte) (int, error) {
	if w.fail {
		w.fail = false
		n := min(len(p), 4)

		_, _ = w.Buffer.Write(p[:n])
		if w.duringWrite != nil {
			w.duringWrite()
		}

		return n, errTemporary
	}

	return w.Buffer.Write(p)
}

type blockingWriter struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
	bytes.Buffer
}

func (w *blockingWriter) Write(p []byte) (int, error) {
	w.once.Do(func() {
		close(w.started)
		<-w.release
	})

	return w.Buffer.Write(p)
}

type reentrantWriter struct {
	bytes.Buffer
	logger *Logger
	once   sync.Once
}

func (w *reentrantWriter) Write(p []byte) (int, error) {
	n, err := w.Buffer.Write(p)
	if err != nil {
		return n, err
	}

	w.once.Do(func() {
		err = w.logger.Info("nested")
	})

	return n, err
}

type lockedBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

var _ io.Writer = (*Logger)(nil)
