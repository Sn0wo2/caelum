package logger

import (
	"fmt"
	"math"
	"runtime"
	"slices"
	"strings"
)

var skipPrefixes = []string{
	"runtime.",
	"github.com/Sn0wo2/caelum.",
	"github.com/Sn0wo2/caelum/internal/logger.",
}

type caller struct {
	ModulePath string
	File       string
	Line       int
}

func callerInfo(callerSkip int) (caller, bool) {
	var pcs [64]uintptr

	n := runtime.Callers(2, pcs[:])

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()

		if !slices.ContainsFunc(skipPrefixes, func(prefix string) bool { return strings.HasPrefix(frame.Function, prefix) }) {
			if callerSkip > 0 {
				callerSkip--
			} else {
				if frame.Line < 0 || uint64(frame.Line) > math.MaxUint32 {
					return caller{}, false
				}

				modulePath := frame.Function

				start := strings.LastIndexByte(modulePath, '/') + 1
				if dot := strings.IndexByte(modulePath[start:], '.'); dot >= 0 {
					modulePath = modulePath[:start+dot]
				}

				return caller{
					ModulePath: modulePath,
					File:       frame.File,
					Line:       frame.Line,
				}, true
			}
		}

		if !more {
			return caller{}, false
		}
	}
}

func stacktrace(callerSkip int) []string {
	var pcs [128]uintptr

	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	stack := make([]string, 0, n)

	for {
		frame, more := frames.Next()

		if !slices.ContainsFunc(skipPrefixes, func(prefix string) bool { return strings.HasPrefix(frame.Function, prefix) }) {
			if callerSkip > 0 {
				callerSkip--
			} else {
				stack = append(stack, fmt.Sprintf("%s (%s:%d)", frame.Function, frame.File, frame.Line))
			}
		}

		if !more {
			return stack
		}
	}
}
