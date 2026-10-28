package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/Sn0wo2/caelum"
)

var iconSets = [...]caelum.Icons{
	caelum.DefaultIcons(),
	{Name: "no-icons"},
	caelum.IconsNerd(),
}

func main() {
	demoFormatIcons()
	demoThemes()
	demoLevels()
	demoLabels()
	demoPathSpans()
	demoTimeFormat()
	demoTargets()
	demoSpans()
	demoInfra()
	demoReload()
}

func demoFormatIcons() {
	section("FORMAT × ICON")

	for _, format := range []caelum.Format{caelum.FormatCompact, caelum.FormatPretty, caelum.FormatJSON} {
		for _, icons := range iconSets {
			writer := caelum.Writer{Format: format}
			if format == caelum.FormatCompact {
				formatter := compactFormatter()
				formatter.Style.Icons = icons
				writer.Compact = &formatter
			}

			fmt.Printf("[-] %s + %s\n", format, icons.Name)
			runWith(writer, emitAllLevels)
		}
	}
}

func demoThemes() {
	section("THEMES")

	themes := []struct {
		name  string
		theme caelum.Theme
	}{
		{"acta", caelum.ThemeActa()},
		{"monokai", caelum.ThemeMonokai()},
		{"dracula", caelum.ThemeDracula()},
		{"nord", caelum.ThemeNord()},
		{"catppuccin_mocha", caelum.ThemeCatppuccinMocha()},
		{"gruvbox", caelum.ThemeGruvbox()},
		{"one_dark", caelum.ThemeOneDark()},
		{"tokyo_night", caelum.ThemeTokyoNight()},
	}

	fmt.Println("[-] Palette overview")

	for _, entry := range themes {
		fmt.Printf("    %-22s accent=%v  secondary=%v  text=%v\n",
			entry.name, entry.theme.Accent, entry.theme.Secondary, entry.theme.Text)
	}

	fmt.Println("[-] Live preview per theme")

	for _, entry := range themes {
		for _, icons := range iconSets {
			fmt.Printf("  [%s / %s]\n", entry.name, icons.Name)

			formatter := compactFormatter()
			formatter.Style.Theme = entry.theme
			formatter.Style.Icons = icons

			runWith(caelum.Writer{Format: caelum.FormatCompact, Compact: &formatter}, func(log *caelum.Logger) {
				emitDemo(log, entry.name)
			})
		}
	}
}

func demoLevels() {
	section("ALL LEVELS")

	for _, icons := range iconSets {
		for _, depth := range []caelum.ColorDepth{
			caelum.ColorDepthTrueColor,
			caelum.ColorDepthANSI256,
			caelum.ColorDepthANSI16,
			caelum.ColorDepthNoColor,
		} {
			fmt.Printf("[-] %s / %s\n", icons.Name, depth)

			formatter := compactFormatter()
			formatter.Style.Icons = icons
			formatter.Style.Theme = caelum.ThemeOneDark()

			runWith(caelum.Writer{
				Format:     caelum.FormatCompact,
				Compact:    &formatter,
				ColorDepth: depth,
			}, emitAllLevels)
		}
	}
}

func demoLabels() {
	section("LABELS")

	for _, entry := range []struct {
		name   string
		labels caelum.LevelLabels
	}{
		{"short", caelum.LevelLabelsShort()},
		{"medium", caelum.LevelLabelsMedium()},
		{"long", caelum.LevelLabelsLong()},
	} {
		fmt.Printf("[-] %s\n", entry.name)

		formatter := compactFormatter()
		formatter.Style.Labels = entry.labels

		runWith(caelum.Writer{Format: caelum.FormatCompact, Compact: &formatter}, func(log *caelum.Logger) {
			emitDemo(log, entry.name)
		})
	}
}

func demoPathSpans() {
	section("PATH & SPANS")

	for _, entry := range []struct {
		showPath    bool
		showSpans   bool
		description string
	}{
		{true, true, "path=on, spans=on"},
		{false, false, "path=off, spans=off"},
	} {
		fmt.Printf("[-] %s\n", entry.description)

		formatter := caelum.DefaultFormatter()
		formatter.ShowPath = entry.showPath
		formatter.ShowSpans = entry.showSpans

		runWith(caelum.Writer{Format: caelum.FormatCompact, Compact: &formatter}, func(log *caelum.Logger) {
			emitDemo(log, entry.description)
		})
	}
}

func demoTimeFormat() {
	section("TIME FORMAT")

	for _, entry := range []struct {
		layout      string
		description string
	}{
		{"", "default: 15:04:05"},
		{"2006-01-02 15:04:05.000", "custom: 2006-01-02 15:04:05.000"},
	} {
		fmt.Printf("[-] %s\n", entry.description)

		formatter := compactFormatter()
		formatter.TimeFormat = entry.layout

		runWith(caelum.Writer{Format: caelum.FormatCompact, Compact: &formatter}, func(log *caelum.Logger) {
			emitDemo(log, entry.description)
		})
	}
}

func demoTargets() {
	section("TARGETS")

	for _, target := range []caelum.WriterTarget{caelum.WriterTargetStdout, caelum.WriterTargetStderr} {
		fmt.Printf("[-] %s\n", target)

		if target == caelum.WriterTargetStderr {
			fmt.Fprintln(os.Stderr, "    (output to stderr)")
		}

		formatter := compactFormatter()
		runWith(caelum.Writer{
			Target:  target,
			Format:  caelum.FormatCompact,
			Compact: &formatter,
		}, func(log *caelum.Logger) {
			emitDemo(log, string(target))
		})
	}
}

func demoSpans() {
	section("SPANS")

	for _, icons := range iconSets {
		fmt.Printf("[-] %s\n", icons.Name)

		formatter := caelum.DefaultFormatter()
		formatter.Style.Icons = icons

		runWith(caelum.Writer{Format: caelum.FormatCompact, Compact: &formatter}, func(log *caelum.Logger) {
			must(log.Info("no span"))

			layer1, err := log.CreateSpanContext(context.Background(), caelum.Event{
				Level: caelum.LevelInfo,
				Name:  "layer1",
			})

			must(err)
			defer func() { must(layer1.Close()) }()

			ctx := layer1.Context(context.Background())
			must(log.InfoContext(ctx, "span [layer1]"))

			layer2, err := log.CreateSpanContext(ctx, caelum.Event{
				Level:  caelum.LevelInfo,
				Name:   "layer2",
				Fields: map[string]any{"depth": 2},
			})

			must(err)
			defer func() { must(layer2.Close()) }()

			ctx = layer2.Context(ctx)
			must(log.WarnContext(ctx, "span [layer1 > layer2{depth=2}]"))

			layer3, err := log.CreateSpanContext(ctx, caelum.Event{
				Level: caelum.LevelDebug,
				Name:  "layer3",
			})

			must(err)
			defer func() { must(layer3.Close()) }()

			ctx = layer3.Context(ctx)
			must(log.ErrorContext(ctx, "span [layer1 > layer2 > layer3]"))
		})
	}
}

func demoInfra() {
	section("INFRA")
	fmt.Println("[-] Level → directive")

	for _, level := range []caelum.Level{
		caelum.LevelError,
		caelum.LevelWarn,
		caelum.LevelInfo,
		caelum.LevelDebug,
		caelum.LevelTrace,
	} {
		fmt.Printf("[+] %s → %q\n", strings.ToUpper(level.String()), level.String())
	}

	fmt.Println("[+] Off → \"off\"")
	fmt.Println("[+] Filter → caelum.Config{Filter: \"info,my_crate=debug\"}")
}

func demoReload() {
	section("RELOAD via init")
	fmt.Println("[-] init + runtime reload")

	const dir = "data/logs/full"

	must(os.MkdirAll(dir, 0o750))

	formatter := compactFormatter()

	log, err := caelum.Init(caelum.Config{
		Filter:     "debug",
		Stacktrace: []caelum.Level{},
		Writers: []caelum.Writer{
			{
				Target:     caelum.WriterTargetStdout,
				Format:     caelum.FormatCompact,
				Compact:    &formatter,
				ColorDepth: caelum.ColorDepthTrueColor,
				Stacktrace: []caelum.Level{},
			},
			{
				Target:     caelum.WriterTargetFile,
				Format:     caelum.FormatJSON,
				ColorDepth: caelum.ColorDepthNoColor,
				Stacktrace: []caelum.Level{},
				File:       &caelum.FileConfig{Path: filepath.Join(dir, "app.log")},
			},
		},
	})
	if err != nil {
		fmt.Printf("[X] init: %v\n", err)

		return
	}

	defer func() { must(caelum.Close()) }()

	fmt.Println("[√] init")
	must(caelum.Info("console + file",
		caelum.Bool("init", true),
		caelum.String("caelum_version", log.CaelumVersion()),
		caelum.String("acta_version", log.ActaVersion()),
	))
	must(caelum.SetFilter("warn,demo=trace"))
	fmt.Println("[+] → set_filter(\"warn,demo=trace\")")
	must(caelum.Info("info suppressed"))
	must(caelum.Warn("warn passes"))
	must(caelum.LogEvent(caelum.Event{
		Level:   caelum.LevelTrace,
		Target:  "demo",
		Message: "demo trace passes",
	}))
}

func section(title string) {
	const width = 64

	border := strings.Repeat("─", width)
	padding := max(width-utf8.RuneCountInString(title), 0)
	fmt.Printf("\n┌%s┐\n|%*s%s%*s|\n└%s┘\n",
		border, padding/2, "", title, padding-padding/2, "", border)
}

func compactFormatter() caelum.Formatter {
	formatter := caelum.DefaultFormatter()
	formatter.ShowPath = false
	formatter.ShowSpans = false

	return formatter
}

func runWith(writer caelum.Writer, emit func(*caelum.Logger)) {
	if writer.Target == "" {
		writer.Target = caelum.WriterTargetStdout
	}

	if writer.ColorDepth == "" {
		writer.ColorDepth = caelum.ColorDepthTrueColor
	}

	if writer.Format == caelum.FormatJSON {
		writer.ColorDepth = caelum.ColorDepthNoColor
	}

	if writer.Compact != nil {
		formatter := *writer.Compact
		formatter.ColorDepth = writer.ColorDepth
		writer.Compact = &formatter
	}

	writer.Stacktrace = []caelum.Level{}
	log, err := caelum.New(caelum.Config{
		Filter:     "trace",
		Stacktrace: []caelum.Level{},
		Writers:    []caelum.Writer{writer},
	})

	must(err)
	defer func() { must(log.Close()) }()

	emit(log)
}

func emitDemo(log *caelum.Logger, label string) {
	must(log.Info(label + ": info"))
	must(log.Warn(label+": warn", caelum.String("user", "alice"), caelum.Int("count", 42)))
	must(log.Error(label+": error", caelum.Int("code", 500)))
}

func emitAllLevels(log *caelum.Logger) {
	must(log.Error("error: crash", caelum.Int("code", 500)))
	must(log.Warn("warn: resource limit", caelum.Float64("threshold", 0.9)))
	must(log.Info("info: normal", caelum.Int("users", 42)))
	must(log.Debug("debug: query", caelum.String("query", "SELECT *"), caelum.Int("took_ms", 3)))
	must(log.Trace("trace: idle", caelum.String("state", "idle")))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
