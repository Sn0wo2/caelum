package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type Rotation string

const (
	RotationNone     Rotation = "none"
	RotationRename   Rotation = "rename"
	RotationCompress Rotation = "compress"
)

type WriterTarget string

const (
	WriterTargetStdout      WriterTarget = "stdout"
	WriterTargetStderr      WriterTarget = "stderr"
	WriterTargetBuffered    WriterTarget = "buffered"
	WriterTargetFile        WriterTarget = "file"
	WriterTargetAsyncStdout WriterTarget = "asyncstdout"
	WriterTargetAsyncStderr WriterTarget = "asyncstderr"
)

type AsyncMode string

const AsyncModeNative AsyncMode = "native"

type FileConfig struct {
	Path     string   `json:"path"`
	Rotation Rotation `json:"rotation"`
}

type Writer struct {
	Format     Format
	Compact    *Formatter
	Pretty     *PrettyOptions
	JSON       *JSONOptions
	ColorDepth ColorDepth
	Stacktrace []Level
	Levels     []Level
	Target     WriterTarget
	File       *FileConfig
	AsyncMode  AsyncMode
}

func DefaultWriter() Writer {
	return Writer{Format: FormatCompact, Target: WriterTargetBuffered}
}

func (file FileConfig) MarshalJSON() ([]byte, error) {
	if file.Path == "" {
		return nil, errors.New("file config requires a non-empty path")
	}

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
	fields, err := decodeObject(data, "file config")
	if err != nil {
		return err
	}

	value := FileConfig{Rotation: RotationNone}
	if _, err := decodeField(fields, "path", &value.Path); err != nil {
		return err
	}

	if _, ok := fields["path"]; !ok {
		return errors.New("file config requires path")
	}

	if value.Path == "" {
		return errors.New("file config requires a non-empty path")
	}

	if _, err := decodeField(fields, "rotation", &value.Rotation); err != nil {
		return err
	}

	*file = value

	return nil
}

func (writer Writer) MarshalJSON() ([]byte, error) {
	format, err := encodeFormat(writer)
	if err != nil {
		return nil, err
	}

	target, err := encodeTarget(writer)
	if err != nil {
		return nil, err
	}

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

	var colorDepth *ColorDepth

	depth := writer.ColorDepth
	if depth == "" && writer.Compact != nil {
		depth = writer.Compact.ColorDepth
	}

	if depth == "" && (writer.Target == "" || writer.Target == WriterTargetBuffered) {
		depth = ColorDepthNoColor
	}

	if depth != "" {
		colorDepth = &depth
	}

	return json.Marshal(struct {
		Format     json.RawMessage `json:"format"`
		ColorDepth *ColorDepth     `json:"color_depth"`
		Stacktrace []string        `json:"stacktrace"`
		Levels     any             `json:"levels"`
		Target     json.RawMessage `json:"target"`
	}{
		Format:     format,
		ColorDepth: colorDepth,
		Stacktrace: stacktraceNames,
		Levels:     levels,
		Target:     target,
	})
}

func (writer *Writer) UnmarshalJSON(data []byte) error {
	fields, err := decodeObject(data, "writer")
	if err != nil {
		return err
	}

	value := DefaultWriter()
	value.Stacktrace = []Level{LevelError}

	if raw, ok := fields["format"]; ok {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return errors.New(`field "format" cannot be null`)
		}

		if err := decodeFormat(raw, &value); err != nil {
			return fmt.Errorf("field %q: %w", "format", err)
		}
	}

	colorDepth, _, err := decodeOptionalField[ColorDepth](fields, "color_depth")
	if err != nil {
		return err
	}

	if colorDepth != nil {
		value.ColorDepth = *colorDepth
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

		if err := decodeTarget(raw, &value); err != nil {
			return fmt.Errorf("field %q: %w", "target", err)
		}
	}

	*writer = value

	return nil
}

func encodeTarget(writer Writer) (json.RawMessage, error) {
	target := writer.Target
	if target == "" {
		target = WriterTargetBuffered
	}

	if writer.File != nil && target != WriterTargetFile {
		return nil, errors.New("file config requires a file writer target")
	}

	async := target == WriterTargetAsyncStdout || target == WriterTargetAsyncStderr
	if writer.AsyncMode != "" && !async {
		return nil, errors.New("async mode requires an async writer target")
	}

	switch target {
	case WriterTargetStdout, WriterTargetStderr, WriterTargetBuffered:
		return json.Marshal(target)
	case WriterTargetFile:
		if writer.File == nil {
			return nil, errors.New("file writer target requires a file config")
		}

		return json.Marshal(map[string]FileConfig{"file": *writer.File})
	case WriterTargetAsyncStdout, WriterTargetAsyncStderr:
		mode := writer.AsyncMode
		if mode == "" {
			mode = AsyncModeNative
		}

		return json.Marshal(map[string]AsyncMode{string(target): mode})
	default:
		return nil, fmt.Errorf("invalid writer target %q", target)
	}
}

func decodeTarget(data []byte, writer *Writer) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '"' {
		name, err := decodeString(data)
		if err != nil {
			return err
		}

		switch name {
		case "stdout":
			writer.Target = WriterTargetStdout
		case "stderr":
			writer.Target = WriterTargetStderr
		case "buffered":
			writer.Target = WriterTargetBuffered
		default:
			return fmt.Errorf("invalid writer target %q", name)
		}

		return nil
	}

	variants, err := decodeObject(data, "writer target")
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

			writer.Target, writer.File = WriterTargetFile, &file
		case "asyncstdout", "asyncstderr":
			var mode AsyncMode
			if err := json.Unmarshal(raw, &mode); err != nil {
				return fmt.Errorf("%s writer target: %w", name, err)
			}

			if name == "asyncstdout" {
				writer.Target = WriterTargetAsyncStdout
			} else {
				writer.Target = WriterTargetAsyncStderr
			}

			writer.AsyncMode = mode
		default:
			return fmt.Errorf("unknown writer target variant %q", name)
		}
	}

	return nil
}
