package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Level string

const (
	LevelError Level = "error"
	LevelWarn  Level = "warn"
	LevelInfo  Level = "info"
	LevelDebug Level = "debug"
	LevelTrace Level = "trace"
)

func (level Level) String() string {
	return string(level)
}

func (level Level) Valid() bool {
	switch level {
	case LevelError, LevelWarn, LevelInfo, LevelDebug, LevelTrace:
		return true
	default:
		return false
	}
}

func ParseLevel(value string) (Level, error) {
	level := Level(strings.ToLower(strings.TrimSpace(value)))
	if !level.Valid() {
		return "", fmt.Errorf("invalid log level %q", value)
	}

	return level, nil
}

func (level Level) MarshalJSON() ([]byte, error) {
	if !level.Valid() {
		return nil, fmt.Errorf("invalid log level %q", level)
	}

	return json.Marshal(string(level))
}

func (level *Level) UnmarshalJSON(data []byte) error {
	value, err := decodeString(data)
	if err != nil {
		return err
	}

	parsed, err := ParseLevel(value)
	if err != nil {
		return err
	}

	*level = parsed

	return nil
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
		if err := json.Unmarshal(raw, &levels[i]); err != nil {
			return nil, fmt.Errorf("level at index %d: %w", i, err)
		}
	}

	return levels, nil
}

func encodeRustLevels(levels []Level) ([]string, error) {
	values := make([]string, len(levels))
	for i, level := range levels {
		if !level.Valid() {
			return nil, fmt.Errorf("invalid log level %q", level)
		}

		values[i] = strings.ToUpper(string(level))
	}

	return values, nil
}
