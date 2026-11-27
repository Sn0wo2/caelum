package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

func decodeObject(data []byte, name string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}

	if fields == nil {
		return nil, fmt.Errorf("%s must be an object", name)
	}

	return fields, nil
}

func decodeField[T any](fields map[string]json.RawMessage, name string, value *T) error {
	raw, ok := fields[name]
	if !ok {
		return nil
	}

	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("field %q cannot be null", name)
	}

	if err := json.Unmarshal(raw, value); err != nil {
		return fmt.Errorf("field %q: %w", name, err)
	}

	return nil
}

func decodeOptionalField[T any](fields map[string]json.RawMessage, name string, value **T) error {
	raw, ok := fields[name]
	if !ok {
		return nil
	}

	if err := json.Unmarshal(raw, value); err != nil {
		return fmt.Errorf("field %q: %w", name, err)
	}

	return nil
}

func decodeString(data []byte) (string, error) {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return "", errors.New("expected string")
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return "", err
	}

	return value, nil
}
