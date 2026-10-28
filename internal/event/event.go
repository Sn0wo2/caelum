package event

import (
	"fmt"
	"slices"

	"github.com/Sn0wo2/caelum/internal/config"
	"github.com/Sn0wo2/caelum/internal/utils"
)

var skipPrefixes = []string{
	"runtime.",
	"github.com/Sn0wo2/caelum.",
	"github.com/Sn0wo2/caelum/internal/builder.",
	"github.com/Sn0wo2/caelum/internal/event.",
}

type Event struct {
	Level      *config.Level  `json:"level,omitempty"`
	Message    string         `json:"message"`
	Target     *string        `json:"target,omitempty"`
	Name       *string        `json:"name,omitempty"`
	ModulePath *string        `json:"module_path,omitempty"`
	File       *string        `json:"file,omitempty"`
	Line       *uint32        `json:"line,omitempty"`
	Fields     map[string]any `json:"fields,omitempty"`
	Parent     *uint32        `json:"parent,omitempty"`
}

func Capture(event Event, disableCaller bool, callerSkip int, levels []config.Level) Event {
	if !disableCaller && (event.Target == nil || event.ModulePath == nil || event.File == nil || event.Line == nil) {
		if caller, ok := utils.CallerInfo(skipPrefixes, callerSkip); ok {
			if event.Target == nil {
				event.Target = &caller.ModulePath
			}

			if event.ModulePath == nil {
				event.ModulePath = &caller.ModulePath
			}

			if event.File == nil {
				event.File = &caller.File
			}

			if event.Line == nil {
				event.Line = &caller.Line
			}
		}
	}

	level := config.LevelInfo
	if event.Level != nil {
		level = *event.Level
	}

	if (levels != nil || level != config.LevelError) && !slices.Contains(levels, level) {
		return event
	}

	frames := utils.Stacktrace(skipPrefixes, callerSkip)
	if len(frames) == 0 {
		return event
	}

	name := "stacktrace"
	for suffix := 1; ; suffix++ {
		if _, exists := event.Fields[name]; !exists {
			break
		}

		name = fmt.Sprintf("go_stacktrace_%d", suffix)
	}

	fields := make(map[string]any, len(event.Fields)+1)
	for key, value := range event.Fields {
		fields[key] = value
	}

	fields[name] = frames
	event.Fields = fields

	return event
}
