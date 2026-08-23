# caelum

> Themed, slog-native logging for Go. A pragmatic Go take on [acta](https://github.com/Sn0wo2/acta).

caelum is a thin layer over the standard library's `log/slog`. It gives you
acta's colorized, themed console output — 8 built-in palettes, icon sets, level
chips, dimmed source paths — plus JSON/text formats and rotating file logs,
without leaving the slog ecosystem.

It deliberately **does not** reimplement tracing-style span trees or per-package
dynamic filter directives. Those don't map onto slog cleanly, and chasing them
is where a Go rewrite stops being worth it. Everything here is either provided
by slog/stdlib or is a few dozen lines of plain, deterministic code.

## Install

```bash
go get github.com/Sn0wo2/caelum
```

Depends on `github.com/muesli/termenv` (terminal capability detection),
`github.com/rs/zerolog` (the diode ring buffer behind async writers), and
`gopkg.in/natefinch/lumberjack.v2` (file rotation).

## Quick start

```go
package main

import "github.com/Sn0wo2/caelum"

func main() {
	log := caelum.New(caelum.Config{Level: caelum.LevelDebug})

	log.Info("Hello, caelum!")
	log.Debug("User logged in", "user", "alice", "id", 42)
}
```

`log` embeds an `*slog.Logger`, so all the usual methods (`Info`, `Warn`,
`Error`, `With`, `WithGroup`, …) are available. Use `caelum.Init` instead of
`New` to also install it as the slog default for the whole program.

## Configuration

```go
log := caelum.New(caelum.Config{
	Level: caelum.LevelInfo,
	Targets: []caelum.Target{
		{                                    // colorized console
			Format:     caelum.Compact,
			Color:      caelum.TrueColor,    // omit to auto-detect from the TTY
			ShowSource: true,                // right-aligned path:line column
			PathWidth:  40,                  // source column width (default 40)
			TimeFormat: "15:04:05",          // UTC timestamps, matching acta
			Style:      &caelum.Style{Theme: caelum.ThemeTokyoNight},
		},
		caelum.FileTarget(caelum.FileOptions{  // rotating JSON file
			Path:     "logs/app.log",
			Compress: true,
		}),
	},
})
```

Multiple targets fan out automatically: one record, written to each
destination in its own format. Each target keeps its own style; a
partially-filled `Style` keeps defaults for its zero-value sections, so
`&caelum.Style{Theme: caelum.ThemeNord}` just works.

## Formats

| Format          | Description                                       |
| --------------- | ------------------------------------------------- |
| `caelum.Compact`| Themed, colorized single line (default).          |
| `caelum.JSON`   | `slog.JSONHandler` — one JSON object per line.    |
| `caelum.Text`   | `slog.TextHandler` — `key=value`.                 |

## Color depth

Auto-detected per target via `termenv`, degrading across all four tiers exactly
like acta's supports-color:

| Depth            | When                                                        |
| ---------------- | ----------------------------------------------------------- |
| `TrueColor`      | `COLORTERM=truecolor`/`24bit`, modern terminals.            |
| `Ansi256`        | `TERM=*-256color`.                                          |
| `Ansi16`         | A basic ANSI terminal.                                      |
| `NoColor`        | Not a terminal (file/pipe/buffer), or `NO_COLOR` is set.    |

Force a depth with `Target.Color`; the zero value `ColorAuto` auto-detects.
When set explicitly, RGB colors degrade to `Ansi256` (xterm cube + grayscale
ramp) or `Ansi16` (nearest of the 16 basic colors). Detection sees *through*
async wrappers (via `Unwrap`) to the real terminal.

## Everything is a writer

A `Target.Writer` is a plain `io.Writer`, so anything that writes bytes is a
valid sink, and decorators compose freely — there is no special "async target"
type the way acta needs `WriterTarget::AsyncStdout`.

### Async (non-blocking) writers

`caelum.Async` wraps any writer so a slow sink never blocks the logging
goroutine. It is backed by `rs/zerolog/diode`, a lock-free ring buffer drained
by a background worker.

```go
out := caelum.Async(os.Stdout)                 // wrap a terminal
out := caelum.Async(file, caelum.WithBuffer(4096))
out := caelum.Async(file, caelum.WithAlerter(func(missed int) { /* count drops */ }))

log := caelum.New(caelum.Config{Targets: []caelum.Target{{Writer: out}}})
defer log.Close()  // flush queues + sync files before exit
```

`Async` returns an `*AsyncWriter` (`io.WriteCloser`). Backpressure is
**drop-oldest**: when producers outrun the sink, the oldest queued records are
discarded rather than blocking the caller — `Dropped()` reports the count, and
`WithAlerter` fires on each overflow. If you need lossless logging, write
synchronously instead. `log.Close()` flushes and closes every target writer that
implements `io.Closer` (async queues and rotating files alike), the caelum
equivalent of dropping acta's `TracingGuard`.

## Themes, icons, labels

Eight palettes: `ThemeCaelum` (default), `ThemeMonokai`, `ThemeDracula`,
`ThemeNord`, `ThemeCatppuccinMocha`, `ThemeGruvbox`, `ThemeOneDark`,
`ThemeTokyoNight`. Icon sets `IconsUnicode` (default) and `IconsNerd`. Label
sets `LabelsLong` (default), `LabelsMedium`, `LabelsShort`. All fields are plain
structs — build your own freely.

## Runtime reconfiguration

`SetLevel` is backed by `slog.LevelVar` — safe to call from any goroutine,
takes effect immediately for every target:

```go
log.SetLevel(caelum.LevelWarn)
```

Styles are fixed per target at construction. Like acta, caelum deliberately
has no style hot-swap: build a new logger if you need a different look.

## File logging

`caelum.FileTarget` wraps lumberjack for size-based rotation, backup retention,
age expiry, and optional gzip compression. File output is JSON and never
colored.
