package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Sn0wo2/caelum/internal/utils"
)

func levelFromString(value string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "error":
		return LevelError, nil
	case "warn":
		return LevelWarn, nil
	case "info":
		return LevelInfo, nil
	case "debug":
		return LevelDebug, nil
	case "trace":
		return LevelTrace, nil
	default:
		return 0, fmt.Errorf("invalid log level %q", value)
	}
}

func levelName(level Level) (string, bool) {
	switch level {
	case LevelError:
		return "error", true
	case LevelWarn:
		return "warn", true
	case LevelInfo:
		return "info", true
	case LevelDebug:
		return "debug", true
	case LevelTrace:
		return "trace", true
	default:
		return "", false
	}
}

func decodeLevelSlice(data []byte, nullable bool) ([]Level, error) {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		if nullable {
			return nil, nil
		}

		return nil, errors.New("expected an array of log levels")
	}

	var values []json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, err
	}

	if values == nil {
		return nil, errors.New("expected an array of log levels")
	}

	levels := make([]Level, len(values))
	for i, raw := range values {
		value, err := utils.DecodeString(raw)
		if err != nil {
			return nil, fmt.Errorf("level at index %d: %w", i, err)
		}

		level, err := levelFromString(value)
		if err != nil {
			return nil, fmt.Errorf("level at index %d: %w", i, err)
		}

		levels[i] = level
	}

	return levels, nil
}

func encodeRustLevels(levels []Level) ([]string, error) {
	values := make([]string, len(levels))
	for i, level := range levels {
		name, ok := levelName(level)
		if !ok {
			return nil, fmt.Errorf("invalid log level %d", level)
		}

		values[i] = strings.ToUpper(name)
	}

	return values, nil
}

func (level Level) String() string {
	name, ok := levelName(level)
	if !ok {
		return fmt.Sprintf("level(%d)", level)
	}

	return name
}

func (level Level) MarshalJSON() ([]byte, error) {
	name, ok := levelName(level)
	if !ok {
		return nil, fmt.Errorf("invalid log level %d", level)
	}

	return json.Marshal(name)
}

func (level *Level) UnmarshalJSON(data []byte) error {
	value, err := utils.DecodeString(data)
	if err != nil {
		return err
	}

	parsed, err := levelFromString(value)
	if err != nil {
		return err
	}

	*level = parsed

	return nil
}

func (depth ColorDepth) MarshalJSON() ([]byte, error) {
	var name string

	switch depth {
	case ColorDepthTrueColor:
		name = "TrueColor"
	case ColorDepthAnsi256:
		name = "Ansi256"
	case ColorDepthAnsi16:
		name = "Ansi16"
	case ColorDepthNoColor:
		name = "NoColor"
	default:
		return nil, fmt.Errorf("invalid color depth %d", depth)
	}

	return json.Marshal(name)
}

func (depth *ColorDepth) UnmarshalJSON(data []byte) error {
	name, err := utils.DecodeString(data)
	if err != nil {
		return err
	}

	switch name {
	case "TrueColor":
		*depth = ColorDepthTrueColor
	case "Ansi256":
		*depth = ColorDepthAnsi256
	case "Ansi16":
		*depth = ColorDepthAnsi16
	case "NoColor":
		*depth = ColorDepthNoColor
	default:
		return fmt.Errorf("invalid color depth %q", name)
	}

	return nil
}

func (alignment LevelAlignment) MarshalJSON() ([]byte, error) {
	if alignment != LevelAlignmentLeft && alignment != LevelAlignmentRight {
		return nil, fmt.Errorf("invalid level alignment %q", alignment)
	}

	return json.Marshal(string(alignment))
}

func (alignment *LevelAlignment) UnmarshalJSON(data []byte) error {
	name, err := utils.DecodeString(data)
	if err != nil {
		return err
	}

	parsed := LevelAlignment(name)
	if parsed != LevelAlignmentLeft && parsed != LevelAlignmentRight {
		return fmt.Errorf("invalid level alignment %q", name)
	}

	*alignment = parsed

	return nil
}

func (rotation Rotation) MarshalJSON() ([]byte, error) {
	if rotation == "" {
		rotation = RotationNone
	}

	if rotation != RotationNone && rotation != RotationRename && rotation != RotationCompress {
		return nil, fmt.Errorf("invalid rotation %q", rotation)
	}

	return json.Marshal(string(rotation))
}

func (rotation *Rotation) UnmarshalJSON(data []byte) error {
	name, err := utils.DecodeString(data)
	if err != nil {
		return err
	}

	parsed := Rotation(name)
	if parsed != RotationNone && parsed != RotationRename && parsed != RotationCompress {
		return fmt.Errorf("invalid rotation %q", name)
	}

	*rotation = parsed

	return nil
}

func (rgb *RGB) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("RGB color must be a three-element array")
	}

	var values []json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}

	if len(values) != 3 {
		return errors.New("RGB color must have exactly three channels")
	}

	var color RGB

	for i, raw := range values {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("RGB channel %d cannot be null", i)
		}

		if err := json.Unmarshal(raw, &color[i]); err != nil {
			return fmt.Errorf("RGB channel %d: %w", i, err)
		}
	}

	*rgb = color

	return nil
}

func (theme *Theme) UnmarshalJSON(data []byte) error {
	fields, err := utils.DecodeObject(data, "theme")
	if err != nil {
		return err
	}

	value := DefaultTheme()
	for _, item := range []struct {
		name string
		dest *RGB
	}{{"accent", &value.Accent}, {"secondary", &value.Secondary}, {"text", &value.Text}, {"error", &value.Error}, {"warn", &value.Warn}, {"info", &value.Info}, {"debug", &value.Debug}, {"trace", &value.Trace}} {
		if _, err := utils.DecodeField(fields, item.name, item.dest); err != nil {
			return err
		}
	}

	*theme = value

	return nil
}

func (icons *Icons) UnmarshalJSON(data []byte) error {
	fields, err := utils.DecodeObject(data, "icons")
	if err != nil {
		return err
	}

	value := DefaultIcons()
	for _, item := range []struct {
		name string
		dest *string
	}{{"name", &value.Name}, {"bracket_open", &value.BracketOpen}, {"bracket_close", &value.BracketClose}, {"time_bracket_open", &value.TimeBracketOpen}, {"time_bracket_close", &value.TimeBracketClose}, {"separator", &value.Separator}, {"arrow", &value.Arrow}, {"span_join", &value.SpanJoin}} {
		if _, err := utils.DecodeField(fields, item.name, item.dest); err != nil {
			return err
		}
	}

	*icons = value

	return nil
}

func (labels *LevelLabels) UnmarshalJSON(data []byte) error {
	value, err := decodeLevelLabels(data, DefaultLevelLabels())
	if err == nil {
		*labels = value
	}

	return err
}

func decodeLevelLabels(data []byte, defaults LevelLabels) (LevelLabels, error) {
	fields, err := utils.DecodeObject(data, "level labels")
	if err != nil {
		return LevelLabels{}, err
	}

	for _, item := range []struct {
		name string
		dest *string
	}{{"error", &defaults.Error}, {"warn", &defaults.Warn}, {"info", &defaults.Info}, {"debug", &defaults.Debug}, {"trace", &defaults.Trace}} {
		if _, err := utils.DecodeField(fields, item.name, item.dest); err != nil {
			return LevelLabels{}, err
		}
	}

	if defaults.Alignment, _, err = utils.DecodeOptionalField[LevelAlignment](fields, "alignment"); err != nil {
		return LevelLabels{}, err
	}

	return defaults, nil
}

func (style *Style) UnmarshalJSON(data []byte) error {
	fields, err := utils.DecodeObject(data, "style")
	if err != nil {
		return err
	}

	value := DefaultStyle()
	if _, err := utils.DecodeField(fields, "theme", &value.Theme); err != nil {
		return err
	}

	if _, err := utils.DecodeField(fields, "icons", &value.Icons); err != nil {
		return err
	}

	if raw, ok := fields["labels"]; ok {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return errors.New(`field "labels" cannot be null`)
		}

		labels, err := decodeLevelLabels(raw, LevelLabelsLong())
		if err != nil {
			return fmt.Errorf("field %q: %w", "labels", err)
		}

		value.Labels = labels
	}

	*style = value

	return nil
}

func (formatter Formatter) MarshalJSON() ([]byte, error) {
	defaults := DefaultFormatter()
	if formatter.PathWidth == nil {
		formatter.PathWidth = defaults.PathWidth
	}

	if formatter.ShowPath == nil {
		formatter.ShowPath = defaults.ShowPath
	}

	if formatter.ShowSpans == nil {
		formatter.ShowSpans = defaults.ShowSpans
	}

	if formatter.ColorDepth == nil {
		formatter.ColorDepth = defaults.ColorDepth
	}

	style := DefaultStyle()
	if formatter.Style != nil {
		style = *formatter.Style
	}

	labels := map[string]any{
		"error": style.Labels.Error, "warn": style.Labels.Warn, "info": style.Labels.Info,
		"debug": style.Labels.Debug, "trace": style.Labels.Trace,
	}
	if style.Labels.Alignment != nil {
		labels["alignment"] = *style.Labels.Alignment
	}

	serializedStyle, err := json.Marshal(map[string]any{"theme": style.Theme, "icons": style.Icons, "labels": labels})
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		TimeFormat *string         `json:"time_format"`
		PathWidth  *uint           `json:"path_width"`
		ShowPath   *bool           `json:"show_path"`
		ShowSpans  *bool           `json:"show_spans"`
		ColorDepth *ColorDepth     `json:"color_depth"`
		Style      json.RawMessage `json:"style"`
	}{TimeFormat: formatter.TimeFormat, PathWidth: formatter.PathWidth, ShowPath: formatter.ShowPath, ShowSpans: formatter.ShowSpans, ColorDepth: formatter.ColorDepth, Style: serializedStyle})
}

func (formatter *Formatter) UnmarshalJSON(data []byte) error {
	fields, err := utils.DecodeObject(data, "formatter")
	if err != nil {
		return err
	}

	value := Formatter{}
	if value.TimeFormat, _, err = utils.DecodeOptionalField[string](fields, "time_format"); err != nil {
		return err
	}

	if value.PathWidth, _, err = utils.DecodeOptionalField[uint](fields, "path_width"); err != nil {
		return err
	}

	if value.ShowPath, _, err = utils.DecodeOptionalField[bool](fields, "show_path"); err != nil {
		return err
	}

	if value.ShowSpans, _, err = utils.DecodeOptionalField[bool](fields, "show_spans"); err != nil {
		return err
	}

	if value.ColorDepth, _, err = utils.DecodeOptionalField[ColorDepth](fields, "color_depth"); err != nil {
		return err
	}

	if value.Style, _, err = utils.DecodeOptionalField[Style](fields, "style"); err != nil {
		return err
	}

	*formatter = value

	return nil
}

func (options *PrettyOptions) UnmarshalJSON(data []byte) error {
	fields, err := utils.DecodeObject(data, "pretty options")
	if err != nil {
		return err
	}

	value := DefaultPrettyOptions()
	for _, item := range []struct {
		name string
		dest *bool
	}{{"target", &value.Target}, {"file", &value.File}, {"line_number", &value.LineNumber}} {
		if _, err := utils.DecodeField(fields, item.name, item.dest); err != nil {
			return err
		}
	}

	*options = value

	return nil
}

func (options *JsonOptions) UnmarshalJSON(data []byte) error {
	fields, err := utils.DecodeObject(data, "JSON options")
	if err != nil {
		return err
	}

	value := DefaultJsonOptions()
	for _, item := range []struct {
		name string
		dest *bool
	}{{"target", &value.Target}, {"file", &value.File}, {"line_number", &value.LineNumber}, {"current_span", &value.CurrentSpan}, {"span_list", &value.SpanList}, {"flatten_event", &value.FlattenEvent}} {
		if _, err := utils.DecodeField(fields, item.name, item.dest); err != nil {
			return err
		}
	}

	*options = value

	return nil
}

func (format Format) MarshalJSON() ([]byte, error) {
	variants := 0
	if format.Compact != nil {
		variants++
	}

	if format.Pretty != nil {
		variants++
	}

	if format.Json != nil {
		variants++
	}

	if variants > 1 {
		return nil, errors.New("format must contain exactly one variant")
	}

	if format.Pretty != nil {
		return json.Marshal(map[string]PrettyOptions{"pretty": *format.Pretty})
	}

	if format.Json != nil {
		return json.Marshal(map[string]JsonOptions{"json": *format.Json})
	}

	formatter := DefaultFormatter()
	if format.Compact != nil {
		formatter = *format.Compact
	}

	return json.Marshal(map[string]Formatter{"compact": formatter})
}

func (format *Format) UnmarshalJSON(data []byte) error {
	variants, err := utils.DecodeObject(data, "format")
	if err != nil {
		return err
	}

	if len(variants) != 1 {
		return errors.New("format must have exactly one variant")
	}

	for name, raw := range variants {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("format variant %q cannot be null", name)
		}

		switch name {
		case "compact":
			var formatter Formatter
			if err := json.Unmarshal(raw, &formatter); err != nil {
				return fmt.Errorf("compact format: %w", err)
			}

			*format = Format{Compact: &formatter}
		case "pretty":
			var options PrettyOptions
			if err := json.Unmarshal(raw, &options); err != nil {
				return fmt.Errorf("pretty format: %w", err)
			}

			*format = Format{Pretty: &options}
		case "json":
			var options JsonOptions
			if err := json.Unmarshal(raw, &options); err != nil {
				return fmt.Errorf("JSON format: %w", err)
			}

			*format = Format{Json: &options}
		default:
			return fmt.Errorf("unknown format variant %q", name)
		}
	}

	return nil
}

func (file FileConfig) MarshalJSON() ([]byte, error) {
	rotation := file.Rotation
	if rotation == "" {
		rotation = RotationNone
	}

	return json.Marshal(struct {
		Path     string   `json:"path"`
		Rotation Rotation `json:"rotation"`
	}{Path: file.Path, Rotation: rotation})
}

func (file *FileConfig) UnmarshalJSON(data []byte) error {
	fields, err := utils.DecodeObject(data, "file config")
	if err != nil {
		return err
	}

	value := FileConfig{Rotation: RotationNone}
	if _, err := utils.DecodeField(fields, "path", &value.Path); err != nil {
		return err
	}

	if _, ok := fields["path"]; !ok {
		return errors.New("file config requires path")
	}

	if _, err := utils.DecodeField(fields, "rotation", &value.Rotation); err != nil {
		return err
	}

	*file = value

	return nil
}

func (mode AsyncMode) MarshalJSON() ([]byte, error) {
	if !mode.Native {
		return nil, errors.New("async mode must be native")
	}

	return json.Marshal("native")
}

func (mode *AsyncMode) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '"' {
		name, err := utils.DecodeString(data)
		if err != nil {
			return err
		}

		if name != "native" {
			return fmt.Errorf("invalid async mode %q", name)
		}

		*mode = AsyncMode{Native: true}

		return nil
	}

	variants, err := utils.DecodeObject(data, "async mode")
	if err != nil {
		return err
	}

	if name := variants["custom"]; len(name) > 0 {
		return errors.New(`unsupported async mode "custom": the acta FFI runtime only supports native async`)
	}

	return errors.New("invalid async mode variant")
}

func (target WriterTarget) MarshalJSON() ([]byte, error) {
	variants := 0
	if target.Stdout {
		variants++
	}

	if target.Stderr {
		variants++
	}

	if target.Buffered {
		variants++
	}

	if target.File != nil {
		variants++
	}

	if target.AsyncStdout != nil {
		variants++
	}

	if target.AsyncStderr != nil {
		variants++
	}

	if variants > 1 {
		return nil, errors.New("writer target must contain exactly one variant")
	}

	if target.Stderr {
		return json.Marshal("stderr")
	}

	if target.Buffered {
		return json.Marshal("buffered")
	}

	if target.File != nil {
		return json.Marshal(map[string]FileConfig{"file": *target.File})
	}

	if target.AsyncStdout != nil {
		return json.Marshal(map[string]AsyncMode{"asyncstdout": *target.AsyncStdout})
	}

	if target.AsyncStderr != nil {
		return json.Marshal(map[string]AsyncMode{"asyncstderr": *target.AsyncStderr})
	}

	return json.Marshal("stdout")
}

func (target *WriterTarget) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '"' {
		name, err := utils.DecodeString(data)
		if err != nil {
			return err
		}

		switch name {
		case "stdout":
			*target = WriterTarget{Stdout: true}
		case "stderr":
			*target = WriterTarget{Stderr: true}
		case "buffered":
			*target = WriterTarget{Buffered: true}
		default:
			return fmt.Errorf("invalid writer target %q", name)
		}

		return nil
	}

	variants, err := utils.DecodeObject(data, "writer target")
	if err != nil {
		return err
	}

	if len(variants) != 1 {
		return errors.New("writer target must have exactly one variant")
	}

	for name, raw := range variants {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("writer target variant %q cannot be null", name)
		}

		switch name {
		case "file":
			var file FileConfig
			if err := json.Unmarshal(raw, &file); err != nil {
				return fmt.Errorf("file writer target: %w", err)
			}

			*target = WriterTarget{File: &file}
		case "asyncstdout", "asyncstderr":
			var mode AsyncMode
			if err := json.Unmarshal(raw, &mode); err != nil {
				return fmt.Errorf("%s writer target: %w", name, err)
			}

			if name == "asyncstdout" {
				*target = WriterTarget{AsyncStdout: &mode}
			} else {
				*target = WriterTarget{AsyncStderr: &mode}
			}
		default:
			return fmt.Errorf("unknown writer target variant %q", name)
		}
	}

	return nil
}

func (writer Writer) MarshalJSON() ([]byte, error) {
	stacktrace := writer.Stacktrace
	if stacktrace == nil {
		stacktrace = []Level{LevelError}
	}

	stacktraceNames, err := encodeRustLevels(stacktrace)
	if err != nil {
		return nil, err
	}

	var levels any
	if writer.Levels != nil {
		levels, err = encodeRustLevels(writer.Levels)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(struct {
		Format     Format       `json:"format"`
		ColorDepth *ColorDepth  `json:"color_depth"`
		Stacktrace []string     `json:"stacktrace"`
		Levels     any          `json:"levels"`
		Target     WriterTarget `json:"target"`
	}{Format: writer.Format, ColorDepth: writer.ColorDepth, Stacktrace: stacktraceNames, Levels: levels, Target: writer.Target})
}

func (writer *Writer) UnmarshalJSON(data []byte) error {
	fields, err := utils.DecodeObject(data, "writer")
	if err != nil {
		return err
	}

	value := DefaultWriter()
	value.Stacktrace = []Level{LevelError}

	if raw, ok := fields["format"]; ok {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return errors.New(`field "format" cannot be null`)
		}

		if err := json.Unmarshal(raw, &value.Format); err != nil {
			return fmt.Errorf("field %q: %w", "format", err)
		}
	}

	if raw, ok := fields["color_depth"]; ok {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			value.ColorDepth = nil
		} else {
			var depth ColorDepth
			if err := json.Unmarshal(raw, &depth); err != nil {
				return fmt.Errorf("field %q: %w", "color_depth", err)
			}

			value.ColorDepth = &depth
		}
	}

	if raw, ok := fields["stacktrace"]; ok {
		value.Stacktrace, err = decodeLevelSlice(raw, false)
		if err != nil {
			return fmt.Errorf("field %q: %w", "stacktrace", err)
		}
	}

	if raw, ok := fields["levels"]; ok {
		value.Levels, err = decodeLevelSlice(raw, true)
		if err != nil {
			return fmt.Errorf("field %q: %w", "levels", err)
		}
	}

	if raw, ok := fields["target"]; ok {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return errors.New(`field "target" cannot be null`)
		}

		if err := json.Unmarshal(raw, &value.Target); err != nil {
			return fmt.Errorf("field %q: %w", "target", err)
		}
	}

	*writer = value

	return nil
}

func (config Config) MarshalJSON() ([]byte, error) {
	filter := config.Filter
	if filter == "" && !config.filterExplicit {
		filter = "info"
	}

	writers := config.Writers
	if writers == nil {
		writers = []Writer{DefaultWriter()}
	}

	return json.Marshal(struct {
		Filter  string   `json:"filter"`
		Writers []Writer `json:"writers"`
	}{Filter: filter, Writers: writers})
}

func (config *Config) UnmarshalJSON(data []byte) error {
	fields, err := utils.DecodeObject(data, "config")
	if err != nil {
		return err
	}

	value := DefaultConfig()
	if _, ok := fields["filter"]; ok {
		if _, err := utils.DecodeField(fields, "filter", &value.Filter); err != nil {
			return err
		}

		value.filterExplicit = true
	}

	if raw, ok := fields["writers"]; ok {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return errors.New(`field "writers" cannot be null`)
		}

		if err := json.Unmarshal(raw, &value.Writers); err != nil {
			return fmt.Errorf("field %q: %w", "writers", err)
		}

		if value.Writers == nil {
			value.Writers = []Writer{}
		}
	}

	*config = value

	return nil
}
