package utils

import (
	"fmt"
	"math"
	"runtime"
	"slices"
	"strings"
)

type Caller struct {
	ModulePath string
	File       string
	Line       uint32
}

func CallerInfo(skipPrefixes []string, callerSkip int) (Caller, bool) {
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
					return Caller{}, false
				}

				modulePath := frame.Function

				start := strings.LastIndexByte(modulePath, '/') + 1
				if dot := strings.IndexByte(modulePath[start:], '.'); dot >= 0 {
					modulePath = modulePath[:start+dot]
				}

				return Caller{
					ModulePath: modulePath,
					File:       frame.File,
					Line:       uint32(frame.Line),
				}, true
			}
		}

		if !more {
			return Caller{}, false
		}
	}
}

func Stacktrace(skipPrefixes []string, callerSkip int) []string {
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
