package caelum

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helper: a compact logger writing into buf with forced color depth.
func compactLogger(buf *bytes.Buffer, depth ColorDepth, level slog.Level) *Logger {
	return New(Config{
		Level:   level,
		Targets: []Target{{Writer: buf, Color: depth}},
	})
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

// Each compact target keeps its own style; partial styles fall back to
// defaults for zero-value sections.
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

// Attrs added before a WithGroup must not get that group's prefix.
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

// Inline slog.Group attrs flatten with the same dotted prefix as WithGroup.
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

// ShowSource on one target must not leak a source field into other targets.
func TestJSONSourceIsPerTarget(t *testing.T) {
	var console, file bytes.Buffer
	log := New(Config{
		Level: LevelInfo,
		Targets: []Target{
			{Writer: &console, Color: NoColor, ShowSource: true},
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
	log.Info("dual", "n", 1)

	if !strings.Contains(console.String(), "dual") {
		t.Errorf("console target missing output: %q", console.String())
	}
	var m map[string]any
	if err := json.Unmarshal(file.Bytes(), &m); err != nil {
		t.Fatalf("file target not valid JSON: %v", err)
	}
	if m["msg"] != "dual" {
		t.Errorf("file target wrong payload: %v", m)
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

func TestRGBTo256Grayscale(t *testing.T) {
	// Pure mid-gray should land in the grayscale ramp (232-255), not the cube.
	got := rgbTo256(RGB{128, 128, 128})
	if got < 232 || got > 255 {
		t.Errorf("mid-gray should map to grayscale ramp, got %d", got)
	}
}

func TestRGBTo16Nearest(t *testing.T) {
	cases := []struct {
		in   RGB
		want uint8
	}{
		{RGB{255, 0, 0}, 9},      // bright red
		{RGB{0, 0, 0}, 0},        // black
		{RGB{255, 255, 255}, 15}, // bright white
	}
	for _, c := range cases {
		if got := rgbTo16(c.in); got != c.want {
			t.Errorf("rgbTo16(%v) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestDimDarkens(t *testing.T) {
	c := RGB{200, 200, 200}.dim()
	if c.R != 50 || c.G != 50 || c.B != 50 {
		t.Errorf("dim should quarter each channel, got %v", c)
	}
}
