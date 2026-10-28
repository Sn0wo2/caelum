package caelum

import (
	"io"
	"testing"
)

func benchLogger(b *testing.B, disableCaller bool) *Logger {
	b.Helper()

	logger, err := New(Config{Writer: io.Discard, Filter: "info", DisableCaller: disableCaller})
	if err != nil {
		b.Fatal(err)
	}

	b.Cleanup(func() {
		if err := logger.Close(); err != nil {
			b.Fatal(err)
		}
	})

	return logger
}

func BenchmarkSimple(b *testing.B) {
	logger := benchLogger(b, true)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		if err := logger.Info("hello benchmark message"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSimpleCaller(b *testing.B) {
	logger := benchLogger(b, false)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		if err := logger.Info("hello benchmark message"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFields(b *testing.B) {
	logger := benchLogger(b, true)
	level := LevelInfo

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		if err := logger.Event(Event{Level: &level, Message: "hello benchmark message", Fields: map[string]any{"user": "Me0wo", "count": 42, "ratio": 3.14, "ok": true, "path": "/usr/local/bin/caelum"}}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParallelSimple(b *testing.B) {
	logger := benchLogger(b, true)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var err error

		for pb.Next() {
			err = logger.Info("hello benchmark message")
		}

		if err != nil {
			b.Error(err)
		}
	})
}

func BenchmarkStartup(b *testing.B) {
	b.ReportAllocs()

	for range b.N {
		logger, err := New(Config{Writer: io.Discard, Filter: "info", DisableCaller: true})
		if err != nil {
			b.Fatal(err)
		}

		if err := logger.Info("warm"); err != nil {
			b.Fatal(err)
		}

		if err := logger.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
