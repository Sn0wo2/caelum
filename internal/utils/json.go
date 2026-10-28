package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

func DecodeObject(data []byte, name string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}

	if fields == nil {
		return nil, fmt.Errorf("%s must be an object", name)
	}

	return fields, nil
}

func DecodeField[T any](fields map[string]json.RawMessage, name string, value *T) (bool, error) {
	raw, ok := fields[name]
	if !ok {
		return false, nil
	}

	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return true, fmt.Errorf("field %q cannot be null", name)
	}

	if err := json.Unmarshal(raw, value); err != nil {
		return true, fmt.Errorf("field %q: %w", name, err)
	}

	return true, nil
}

func DecodeOptionalField[T any](fields map[string]json.RawMessage, name string) (*T, bool, error) {
	raw, ok := fields[name]
	if !ok {
		return nil, false, nil
	}

	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, true, nil
	}

	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, true, fmt.Errorf("field %q: %w", name, err)
	}

	return &value, true, nil
}

func DecodeString(data []byte) (string, error) {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return "", errors.New("expected string")
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return "", err
	}

	return value, nil
}
