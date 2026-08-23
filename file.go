package caelum

import "gopkg.in/natefinch/lumberjack.v2"

// FileOptions configures a rotating log file. It is a thin, well-named wrapper
// over lumberjack so callers don't import it directly for the common case.
type FileOptions struct {
	// Path is the active log file. Rotated files get a timestamp suffix.
	Path string
	// MaxSizeMB rotates the file once it exceeds this size. Defaults to 100.
	MaxSizeMB int
	// MaxBackups is the number of rotated files to keep (0 = keep all).
	MaxBackups int
	// MaxAgeDays removes rotated files older than this (0 = never by age).
	MaxAgeDays int
	// Compress gzips rotated files. This is the acta `Rotation::Compress`
	// equivalent, provided for free by lumberjack.
	Compress bool
}

// FileTarget returns a Target that writes rotating, JSON-formatted logs to a
// file. File output never uses ANSI colors. Use it inside Config.Targets:
//
//	caelum.New(caelum.Config{Targets: []caelum.Target{
//	    {},                                  // colored console
//	    caelum.FileTarget(caelum.FileOptions{Path: "logs/app.log", Compress: true}),
//	}})
//
// The returned target's Writer is a *lumberjack.Logger; keep the Logger alive
// for as long as you log. lumberjack flushes on every write, so there is no
// guard to drop (unlike acta's TracingGuard).
func FileTarget(opts FileOptions) Target {
	if opts.MaxSizeMB == 0 {
		opts.MaxSizeMB = 100
	}
	lj := &lumberjack.Logger{
		Filename:   opts.Path,
		MaxSize:    opts.MaxSizeMB,
		MaxBackups: opts.MaxBackups,
		MaxAge:     opts.MaxAgeDays,
		Compress:   opts.Compress,
	}
	return Target{
		Writer: lj,
		Format: JSON,
		Color:  NoColor,
	}
}
