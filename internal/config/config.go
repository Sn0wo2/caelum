package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/Sn0wo2/caelum/internal/loader"
)

type Config struct {
	Loader         loader.Loader
	Writer         io.Writer
	Capacity       int
	NativeConfig   json.RawMessage
	Filter         string
	Writers        []Writer
	DisableCaller  bool
	CallerSkip     int
	Stacktrace     []Level
	filterExplicit bool
}

func DefaultConfig() Config {
	return Config{Filter: "info", Writers: []Writer{DefaultWriter()}}
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
	fields, err := decodeObject(data, "config")
	if err != nil {
		return err
	}

	value := DefaultConfig()
	if _, ok := fields["filter"]; ok {
		if _, err := decodeField(fields, "filter", &value.Filter); err != nil {
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
