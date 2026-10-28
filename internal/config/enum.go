package config

import (
	"encoding/json"
	"fmt"
	"slices"
)

func encodeEnum[T ~string](value T, name string, values ...T) ([]byte, error) {
	if !slices.Contains(values, value) {
		return nil, fmt.Errorf("invalid %s %q", name, value)
	}

	return json.Marshal(string(value))
}

func decodeEnum[T ~string](data []byte, dest *T, name string, values ...T) error {
	value, err := decodeString(data)
	if err != nil {
		return err
	}

	parsed := T(value)
	if !slices.Contains(values, parsed) {
		return fmt.Errorf("invalid %s %q", name, value)
	}

	*dest = parsed

	return nil
}

func (depth ColorDepth) MarshalJSON() ([]byte, error) {
	return encodeEnum(
		depth, "color depth",
		ColorDepthTrueColor,
		ColorDepthANSI256,
		ColorDepthANSI16,
		ColorDepthNoColor,
	)
}

func (depth *ColorDepth) UnmarshalJSON(data []byte) error {
	return decodeEnum(
		data, depth, "color depth",
		ColorDepthTrueColor,
		ColorDepthANSI256,
		ColorDepthANSI16,
		ColorDepthNoColor,
	)
}

func (alignment LevelAlignment) MarshalJSON() ([]byte, error) {
	return encodeEnum(alignment, "level alignment", LevelAlignmentLeft, LevelAlignmentRight)
}

func (alignment *LevelAlignment) UnmarshalJSON(data []byte) error {
	return decodeEnum(data, alignment, "level alignment", LevelAlignmentLeft, LevelAlignmentRight)
}

func (rotation Rotation) MarshalJSON() ([]byte, error) {
	if rotation == "" {
		rotation = RotationNone
	}

	return encodeEnum(rotation, "rotation", RotationNone, RotationRename, RotationCompress)
}

func (rotation *Rotation) UnmarshalJSON(data []byte) error {
	return decodeEnum(data, rotation, "rotation", RotationNone, RotationRename, RotationCompress)
}

func (format Format) MarshalJSON() ([]byte, error) {
	return encodeEnum(format, "format", FormatCompact, FormatPretty, FormatJSON)
}

func (format *Format) UnmarshalJSON(data []byte) error {
	return decodeEnum(data, format, "format", FormatCompact, FormatPretty, FormatJSON)
}

func (target WriterTarget) MarshalJSON() ([]byte, error) {
	return encodeEnum(
		target, "writer target",
		WriterTargetStdout,
		WriterTargetStderr,
		WriterTargetBuffered,
		WriterTargetFile,
		WriterTargetAsyncStdout,
		WriterTargetAsyncStderr,
	)
}

func (target *WriterTarget) UnmarshalJSON(data []byte) error {
	return decodeEnum(
		data, target, "writer target",
		WriterTargetStdout,
		WriterTargetStderr,
		WriterTargetBuffered,
		WriterTargetFile,
		WriterTargetAsyncStdout,
		WriterTargetAsyncStderr,
	)
}

func (mode AsyncMode) MarshalJSON() ([]byte, error) {
	return encodeEnum(mode, "async mode", AsyncModeNative)
}

func (mode *AsyncMode) UnmarshalJSON(data []byte) error {
	return decodeEnum(data, mode, "async mode", AsyncModeNative)
}
