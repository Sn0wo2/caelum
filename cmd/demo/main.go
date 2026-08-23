package main

import (
	"os"

	"github.com/Sn0wo2/caelum"
)

func main() {
	// Force colors so the demo looks right even when piped into a file/pager.
	// "Everything is a writer": wrap stdout in a non-blocking async writer.
	out := caelum.Async(os.Stdout)

	log := caelum.New(caelum.Config{
		Level: caelum.LevelDebug,
		Targets: []caelum.Target{
			{Writer: out, Color: caelum.TrueColor, ShowSource: true},
		},
	})
	// Drain the async queue and release writers before exit.
	defer log.Close()

	log.Info("Hello, caelum!")
	log.Debug("User logged in", "user", "alice", "id", 42)
	log.Warn("disk almost full", "free_pct", 7)
	log.Error("request failed", "status", 500, "path", "/api/x")

	// Grouped attributes show the dotted prefix.
	log.WithGroup("http").Info("served", "method", "GET", "ms", 12)

	// Runtime level swap, the set_level equivalent.
	log.SetLevel(caelum.LevelWarn)
	log.Info("this is filtered out")
	log.Warn("only warnings and above now")

	// Styles are fixed per target: build a themed logger per palette. Partial
	// styles keep defaults for their zero-value sections.
	for name, th := range map[string]caelum.Theme{
		"dracula": caelum.ThemeDracula,
		"monokai": caelum.ThemeMonokai,
		"nord":    caelum.ThemeNord,
		"tokyo":   caelum.ThemeTokyoNight,
	} {
		themed := caelum.New(caelum.Config{Targets: []caelum.Target{{
			Writer: out,
			Color:  caelum.TrueColor,
			Style:  &caelum.Style{Theme: th, Labels: caelum.LabelsMedium},
		}}})
		themed.Warn("theme demo", "theme", name)
	}
}
