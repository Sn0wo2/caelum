package caelum

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func compactLogger(buf *bytes.Buffer, depth ColorDepth, level slog.Level) *Logger {
	return New(Config{
		Level:   level,
		Targets: []Target{{Writer: buf, Color: depth}},
	})
}

func TestTargetDefaults(t *testing.T) {
	var buf bytes.Buffer
	h := New(Config{Targets: []Target{{Writer: &buf}}}).Handler().(*Handler)
	if h.depth != NoColor || h.timeFormat != "15:04:05" || h.pathWidth != 40 || h.style != DefaultStyle() {
		t.Errorf("unexpected target defaults: %+v", h)
	}
}

func TestNoColorIsPlain(t *testing.T) {
	var buf bytes.Buffer
	log := compactLogger(&buf, NoColor, LevelInfo)
	log.Info("hello", "k", "v")

	out := buf.String()
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("NoColor output must not contain escape sequences: %q", out)
	}
	for _, want := range []string{"INFO", "hello", "k", "=", "v"} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q missing %q", out, want)
		}
	}
}

func TestTrueColorHasEscapes(t *testing.T) {
	var buf bytes.Buffer
	log := compactLogger(&buf, TrueColor, LevelInfo)
	log.Info("hi")
	out := buf.String()
	if !strings.Contains(out, "\x1b[38;2;") {
		t.Fatalf("TrueColor must emit 24-bit fg sequences, got %q", out)
	}
	if !strings.Contains(out, reset) {
		t.Errorf("missing reset sequence in %q", out)
	}
}

func TestAnsiProfilesHaveEscapes(t *testing.T) {
	for _, depth := range []ColorDepth{Ansi256, Ansi16} {
		var buf bytes.Buffer
		compactLogger(&buf, depth, LevelInfo).Info("hi")
		if !strings.Contains(buf.String(), "\x1b[") {
			t.Errorf("%v must emit an ANSI sequence, got %q", depth, buf.String())
		}
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	log := compactLogger(&buf, NoColor, LevelWarn)
	log.Info("dropped")
	log.Warn("kept")
	out := buf.String()
	if strings.Contains(out, "dropped") {
		t.Errorf("info should be filtered at warn level: %q", out)
	}
	if !strings.Contains(out, "kept") {
		t.Errorf("warn should pass: %q", out)
	}
}

func TestSetLevelRuntime(t *testing.T) {
	var buf bytes.Buffer
	log := compactLogger(&buf, NoColor, LevelInfo)
	log.SetLevel(LevelError)
	log.Warn("nope")
	log.Error("yes")
	out := buf.String()
	if strings.Contains(out, "nope") {
		t.Errorf("warn should be filtered after SetLevel(Error): %q", out)
	}
	if !strings.Contains(out, "yes") {
		t.Errorf("error should pass: %q", out)
	}
}

func TestTraceLevel(t *testing.T) {
	var buf bytes.Buffer
	log := compactLogger(&buf, NoColor, LevelTrace)
	log.Log(context.Background(), LevelTrace, "deep")
	if out := buf.String(); !strings.Contains(out, "TRACE") {
		t.Errorf("expected TRACE label, got %q", out)
	}
}

func TestPerTargetStyle(t *testing.T) {
	var a, b bytes.Buffer
	log := New(Config{
		Level: LevelInfo,
		Targets: []Target{
			{Writer: &a, Color: NoColor, Style: &Style{Labels: LabelsShort}},
			{Writer: &b, Color: NoColor, Style: &Style{Labels: LabelsMedium}},
		},
	})
	log.Info("x")
	if out := a.String(); !strings.Contains(out, "[I]") {
		t.Errorf("first target should use short labels: %q", out)
	}
	if out := b.String(); !strings.Contains(out, "[INF]") {
		t.Errorf("second target should use medium labels: %q", out)
	}
}

func TestGroupPrefix(t *testing.T) {
	var buf bytes.Buffer
	log := compactLogger(&buf, NoColor, LevelInfo)
	log.WithGroup("http").Info("served", "method", "GET")
	out := buf.String()
	if !strings.Contains(out, "http.method=GET") {
		t.Errorf("expected dotted group prefix, got %q", out)
	}
}

func TestGroupOnlyPrefixesLaterAttrs(t *testing.T) {
	var buf bytes.Buffer
	log := compactLogger(&buf, NoColor, LevelInfo)
	log.With("svc", "api").WithGroup("http").Info("served", "method", "GET")
	out := buf.String()
	if !strings.Contains(out, "svc=api") || strings.Contains(out, "http.svc") {
		t.Errorf("pre-group attr must stay unprefixed: %q", out)
	}
	if !strings.Contains(out, "http.method=GET") {
		t.Errorf("post-group attr must be prefixed: %q", out)
	}
}

func TestInlineGroupFlattens(t *testing.T) {
	var buf bytes.Buffer
	log := compactLogger(&buf, NoColor, LevelInfo)
	log.Info("q", slog.Group("req", slog.String("m", "GET"), slog.Int("code", 200)))
	out := buf.String()
	for _, want := range []string{"req.m=GET", "req.code=200"} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q missing %q", out, want)
		}
	}
}

func TestWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	log := compactLogger(&buf, NoColor, LevelInfo)
	log.With("svc", "api").Info("up", "port", 8080)
	out := buf.String()
	for _, want := range []string{"svc=api", "port=8080"} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q missing %q", out, want)
		}
	}
}

func TestJSONTarget(t *testing.T) {
	var buf bytes.Buffer
	log := New(Config{
		Level:   LevelInfo,
		Targets: []Target{{Writer: &buf, Format: JSON}},
	})
	log.Info("event", "user", "bob")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("JSON target did not emit valid JSON: %v (%q)", err, buf.String())
	}
	if m["msg"] != "event" || m["user"] != "bob" {
		t.Errorf("unexpected JSON payload: %v", m)
	}
}

func TestJSONSourceIsPerTarget(t *testing.T) {
	var console, file bytes.Buffer
	log := New(Config{
		Level: LevelInfo,
		Targets: []Target{
			{Writer: &console, Color: NoColor, AddSource: true},
			{Writer: &file, Format: JSON},
		},
	})
	log.Info("x")
	var m map[string]any
	if err := json.Unmarshal(file.Bytes(), &m); err != nil {
		t.Fatalf("file target not valid JSON: %v", err)
	}
	if _, ok := m["source"]; ok {
		t.Errorf("console ShowSource leaked into JSON output: %v", m)
	}
}

func TestFanoutMultipleTargets(t *testing.T) {
	var console, file bytes.Buffer
	log := New(Config{
		Level: LevelInfo,
		Targets: []Target{
			{Writer: &console, Color: NoColor, Format: Compact},
			{Writer: &file, Format: JSON},
		},
	})
	log.With("svc", "api").WithGroup("http").Info("dual", "n", 1)

	for _, want := range []string{"dual", "svc=api", "http.n=1"} {
		if !strings.Contains(console.String(), want) {
			t.Errorf("console target missing %q: %q", want, console.String())
		}
	}
	var m map[string]any
	if err := json.Unmarshal(file.Bytes(), &m); err != nil {
		t.Fatalf("file target not valid JSON: %v", err)
	}
	if m["msg"] != "dual" || m["svc"] != "api" {
		t.Errorf("file target wrong payload: %v", m)
	}
	if group, ok := m["http"].(map[string]any); !ok || group["n"] != float64(1) {
		t.Errorf("file target missing grouped attr: %v", m)
	}
}

func TestFanoutReturnsAllErrors(t *testing.T) {
	first, second := errors.New("first"), errors.New("second")
	log := New(Config{Targets: []Target{
		{Writer: errorWriter{first}, Format: Text},
		{Writer: errorWriter{second}, Format: JSON},
	}})
	err := log.Handler().Handle(context.Background(), slog.NewRecord(time.Time{}, LevelInfo, "x", 0))
	if !errors.Is(err, first) || !errors.Is(err, second) {
		t.Fatalf("fanout error = %v, want both target errors", err)
	}
}

func TestFormatPathAligns(t *testing.T) {
	got := formatPath("/home/me/proj/src/http/server.go", 42, 30, "")
	if len(got) != 30 || !strings.HasSuffix(got, "http/server.go:42") {
		t.Errorf("short path should right-align to width, got %q (len %d)", got, len(got))
	}

	long := formatPath("/very/long/prefix/pkg/deeply/nested/dir/handler.go", 7, 20, "")
	if len(long) != 20 || !strings.HasSuffix(long, "dir/handler.go:7") {
		t.Errorf("over-width path should keep dir tail + filename, got %q (len %d)", long, len(long))
	}

	backslash := formatPath(`D:\proj\src\fmt\mod.go`, 3, 15, "")
	if !strings.HasSuffix(backslash, "fmt/mod.go:3") {
		t.Errorf("backslashes should normalize to slashes, got %q", backslash)
	}

	root := t.TempDir()
	rootFile := formatPath(filepath.Join(root, "main.go"), 33, 20, root)
	if !strings.HasSuffix(rootFile, "main.go:33") || strings.Contains(rootFile, filepath.Base(root)) {
		t.Errorf("project-root file should omit the root path, got %q", rootFile)
	}

	nested := formatPath(filepath.Join(root, "pkg", "cli", "cli.go"), 12, 30, root)
	if !strings.HasSuffix(nested, "pkg/cli/cli.go:12") {
		t.Errorf("nested file should be relative to the project root, got %q", nested)
	}
}

func TestHandlerFindsModuleRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/app\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	dist := filepath.Join(root, "dist")
	if err := os.Mkdir(dist, 0o750); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dist)

	target := Target{Writer: &bytes.Buffer{}, Color: NoColor}.resolve()
	h := newHandler(target, LevelInfo)
	if h.sourceRoot != root {
		t.Fatalf("source root = %q, want module root %q", h.sourceRoot, root)
	}
}

func TestDimDarkens(t *testing.T) {
	c := RGB{200, 200, 200}.dim()
	if c.R != 50 || c.G != 50 || c.B != 50 {
		t.Errorf("dim should quarter each channel, got %v", c)
	}
}

func TestNewHandlerUsesSlogOptions(t *testing.T) {
	var buf bytes.Buffer
	var level slog.LevelVar
	level.Set(slog.LevelInfo)

	h := NewHandler(&buf, &HandlerOptions{
		HandlerOptions: slog.HandlerOptions{
			Level: &level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == "password" {
					return slog.String(a.Key, "redacted")
				}
				return a
			},
		},
		Color: NoColor,
	})
	log := slog.New(h)
	log.Info("login", "password", "secret")

	out := buf.String()
	if strings.Contains(out, "secret") || !strings.Contains(out, "password=redacted") {
		t.Fatalf("ReplaceAttr was not applied: %q", out)
	}

	level.Set(slog.LevelError)
	log.Info("dropped")
	log.Error("kept")
	if strings.Contains(buf.String(), "dropped") || !strings.Contains(buf.String(), "kept") {
		t.Fatalf("LevelVar was not honored: %q", buf.String())
	}
}

func TestCompactValueFormatting(t *testing.T) {
	for _, tc := range []struct {
		value slog.Value
		want  string
	}{
		{slog.StringValue(""), `""`},
		{slog.StringValue("a b\n"), `"a b\n"`},
		{slog.StringValue(`a="b"\c`), `"a=\"b\"\\c"`},
		{slog.Int64Value(-42), "-42"},
		{slog.Uint64Value(^uint64(0)), "18446744073709551615"},
		{slog.Float64Value(1.25), "1.25"},
		{slog.BoolValue(true), "true"},
		{slog.DurationValue(1500 * time.Microsecond), "1.5ms"},
		{slog.TimeValue(time.Date(2026, 9, 6, 1, 2, 3, 123456789, time.UTC)), "2026-09-06T01:02:03.123456789Z"},
		{slog.AnyValue(errors.New("a b")), `"a b"`},
		{slog.AnyValue(nil), "<nil>"},
		{slog.AnyValue(userValue{}), `"[name=alice]"`},
	} {
		if got := compactValue(tc.value); got != tc.want {
			t.Errorf("compactValue(%v) = %q, want %q", tc.value, got, tc.want)
		}
	}
}

func TestCompactHandlerUsesLogValuerAndQuotesUnsafeValues(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(NewHandler(&buf, &HandlerOptions{Color: NoColor}))
	log.Info("hello\nworld", "user", userValue{}, "note", "a b")

	out := buf.String()
	if strings.Count(out, "\n") != 1 {
		t.Fatalf("compact output must stay on one line: %q", out)
	}
	for _, want := range []string{`"hello\nworld"`, "user.name=alice", `note="a b"`} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q missing %q", out, want)
		}
	}
}

func TestCompactHandlerOmitsZeroRecordTime(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, &HandlerOptions{Color: NoColor})
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "zero time", 0)
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if strings.Contains(out, "┇") || strings.Contains(out, "｣") {
		t.Errorf("zero record time should be omitted: %q", out)
	}
	if !strings.Contains(out, "zero time") {
		t.Errorf("message missing from output: %q", out)
	}
}

type userValue struct{}

func (userValue) LogValue() slog.Value {
	return slog.GroupValue(slog.String("name", "alice"))
}

type errorWriter struct{ err error }

func (w errorWriter) Write([]byte) (int, error) { return 0, w.err }
