package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"slices"

	"github.com/Sn0wo2/caelum/internal/config"
)

type Event struct {
	Level      config.Level   `json:"level,omitempty"`
	Message    string         `json:"message"`
	Target     string         `json:"target,omitempty"`
	Name       string         `json:"name,omitempty"`
	ModulePath string         `json:"module_path,omitempty"`
	File       string         `json:"file,omitempty"`
	Line       int            `json:"line,omitempty"`
	Fields     map[string]any `json:"fields,omitempty"`
	Parent     *Span          `json:"-"`
}

type eventJSON Event

func (value Event) MarshalJSON() ([]byte, error) {
	if err := value.Parent.check(nil); err != nil {
		return nil, err
	}

	return marshalEvent(value)
}

func (value *Event) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("%w: event must be an object", ErrInvalidArgument)
	}

	var fields struct {
		eventJSON
		Parent json.RawMessage `json:"parent"`
	}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	if len(fields.Parent) != 0 && !bytes.Equal(bytes.TrimSpace(fields.Parent), []byte("null")) {
		return fmt.Errorf("%w: parent span IDs cannot be decoded; set Event.Parent to a Span or use context propagation", ErrInvalidArgument)
	}

	if fields.Line < 0 || uint64(fields.Line) > math.MaxUint32 {
		return fmt.Errorf("%w: event line must fit in uint32", ErrInvalidArgument)
	}

	*value = Event(fields.eventJSON)

	return nil
}

func marshalEvent(value Event) ([]byte, error) {
	if value.Line < 0 || uint64(value.Line) > math.MaxUint32 {
		return nil, fmt.Errorf("%w: event line must fit in uint32", ErrInvalidArgument)
	}

	var parent SpanID
	if value.Parent != nil {
		parent = value.Parent.id
	}

	value.Fields = normalizeFields(value.Fields)

	return json.Marshal(struct {
		eventJSON
		Parent SpanID `json:"parent,omitempty"`
	}{eventJSON: eventJSON(value), Parent: parent})
}

func captureEvent(event Event, disableCaller bool, callerSkip int, levels []config.Level) Event {
	if !disableCaller && (event.Target == "" || event.ModulePath == "" || event.File == "" || event.Line == 0) {
		if caller, ok := callerInfo(callerSkip); ok {
			if event.Target == "" {
				event.Target = caller.ModulePath
			}

			if event.ModulePath == "" {
				event.ModulePath = caller.ModulePath
			}

			if event.File == "" {
				event.File = caller.File
			}

			if event.Line == 0 {
				event.Line = caller.Line
			}
		}
	}

	level := event.Level
	if level == "" {
		level = config.LevelInfo
	}

	if (levels != nil || level != config.LevelError) && !slices.Contains(levels, level) {
		return event
	}

	frames := stacktrace(callerSkip)
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
