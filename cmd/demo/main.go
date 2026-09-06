package main

import (
	"os"

	"github.com/Sn0wo2/caelum"
)

func main() {
	out := caelum.Async(os.Stdout, caelum.AsyncConfig{})

	log := caelum.New(caelum.Config{
		Level: caelum.LevelDebug,
		Targets: []caelum.Target{
			{Writer: out, Color: caelum.TrueColor, AddSource: true},
		},
	})
	defer func() { _ = out.Close() }()

	log.Info("Hello, caelum!")
	log.Debug("User logged in", "user", "alice", "id", 42)
	log.Warn("disk almost full", "free_pct", 7)
	log.Error("request failed", "status", 500, "path", "/api/x")

	log.WithGroup("http").Info("served", "method", "GET", "ms", 12)

	log.SetLevel(caelum.LevelWarn)
	log.Info("this is filtered out")
	log.Warn("only warnings and above now")

	for _, item := range []struct {
		name  string
		theme caelum.Theme
	}{
		{name: "dracula", theme: caelum.ThemeDracula},
		{name: "monokai", theme: caelum.ThemeMonokai},
		{name: "nord", theme: caelum.ThemeNord},
		{name: "tokyo", theme: caelum.ThemeTokyoNight},
	} {
		themed := caelum.New(caelum.Config{Targets: []caelum.Target{{
			Writer: out,
			Color:  caelum.TrueColor,
			Style:  &caelum.Style{Theme: item.theme, Labels: caelum.LabelsMedium},
		}}})
		themed.Warn("theme demo", "theme", item.name)
	}
}
