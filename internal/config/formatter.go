package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type Format string

const (
	FormatCompact Format = "compact"
	FormatPretty  Format = "pretty"
	FormatJSON    Format = "json"
)

type Formatter struct {
	TimeFormat  string
	DisableTime bool
	PathWidth   int
	ShowPath    bool
	ShowSpans   bool
	ColorDepth  ColorDepth
	Style       *Style
}

type PrettyOptions struct {
	Target     bool `json:"target"`
	File       bool `json:"file"`
	LineNumber bool `json:"line_number"`
}

type JSONOptions struct {
	Target       bool `json:"target"`
	File         bool `json:"file"`
	LineNumber   bool `json:"line_number"`
	CurrentSpan  bool `json:"current_span"`
	SpanList     bool `json:"span_list"`
	FlattenEvent bool `json:"flatten_event"`
}

func DefaultFormatter() Formatter {
	style := DefaultStyle()

	return Formatter{PathWidth: 28, ShowPath: true, ShowSpans: true, Style: &style}
}

func DefaultPrettyOptions() PrettyOptions {
	return PrettyOptions{Target: true, File: true, LineNumber: true}
}

func DefaultJSONOptions() JSONOptions {
	return JSONOptions{FlattenEvent: true}
}

func (formatter Formatter) MarshalJSON() ([]byte, error) {
	if formatter.PathWidth < 0 {
		return nil, errors.New("formatter path width cannot be negative")
	}

	if formatter.DisableTime && formatter.TimeFormat != "" {
		return nil, errors.New("formatter time format cannot be set when time is disabled")
	}

	var timeFormat *string
	if formatter.DisableTime {
		timeFormat = new("")
	} else if formatter.TimeFormat != "" {
		format, err := encodeTimeLayout(formatter.TimeFormat)
		if err != nil {
			return nil, fmt.Errorf("formatter time layout: %w", err)
		}

		timeFormat = &format
	}

	if formatter.ColorDepth == "" {
		formatter.ColorDepth = ColorDepthTrueColor
	}

	if formatter.Style == nil {
		style := DefaultStyle()
		formatter.Style = &style
	}

	return json.Marshal(struct {
		TimeFormat *string    `json:"time_format"`
		PathWidth  int        `json:"path_width"`
		ShowPath   bool       `json:"show_path"`
		ShowSpans  bool       `json:"show_spans"`
		ColorDepth ColorDepth `json:"color_depth"`
		Style      *Style     `json:"style"`
	}{
		TimeFormat: timeFormat,
		PathWidth:  formatter.PathWidth,
		ShowPath:   formatter.ShowPath,
		ShowSpans:  formatter.ShowSpans,
		ColorDepth: formatter.ColorDepth,
		Style:      formatter.Style,
	})
}

func (formatter *Formatter) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("formatter must be an object")
	}

	var fields struct {
		TimeFormat *string     `json:"time_format"`
		PathWidth  *int        `json:"path_width"`
		ShowPath   *bool       `json:"show_path"`
		ShowSpans  *bool       `json:"show_spans"`
		ColorDepth *ColorDepth `json:"color_depth"`
		Style      *Style      `json:"style"`
	}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	value := DefaultFormatter()
	if fields.TimeFormat != nil {
		value.DisableTime = *fields.TimeFormat == ""
		if !value.DisableTime {
			layout, err := decodeTimeLayout(*fields.TimeFormat)
			if err != nil {
				return fmt.Errorf("formatter time format: %w", err)
			}

			value.TimeFormat = layout
		}
	}

	if fields.PathWidth != nil {
		value.PathWidth = *fields.PathWidth
	}

	if value.PathWidth < 0 {
		return errors.New("formatter path width cannot be negative")
	}

	if fields.ShowPath != nil {
		value.ShowPath = *fields.ShowPath
	}

	if fields.ShowSpans != nil {
		value.ShowSpans = *fields.ShowSpans
	}

	if fields.ColorDepth != nil {
		value.ColorDepth = *fields.ColorDepth
	}

	if fields.Style != nil {
		value.Style = fields.Style
	}

	*formatter = value

	return nil
}

func (options *PrettyOptions) UnmarshalJSON(data []byte) error {
	fields, err := decodeObject(data, "pretty options")
	if err != nil {
		return err
	}

	value := DefaultPrettyOptions()
	for _, item := range []struct {
		name string
		dest *bool
	}{
		{"target", &value.Target},
		{"file", &value.File},
		{"line_number", &value.LineNumber},
	} {
		if _, err := decodeField(fields, item.name, item.dest); err != nil {
			return err
		}
	}

	*options = value

	return nil
}

func (options *JSONOptions) UnmarshalJSON(data []byte) error {
	fields, err := decodeObject(data, "JSON options")
	if err != nil {
		return err
	}

	value := DefaultJSONOptions()
	for _, item := range []struct {
		name string
		dest *bool
	}{
		{"target", &value.Target},
		{"file", &value.File},
		{"line_number", &value.LineNumber},
		{"current_span", &value.CurrentSpan},
		{"span_list", &value.SpanList},
		{"flatten_event", &value.FlattenEvent},
	} {
		if _, err := decodeField(fields, item.name, item.dest); err != nil {
			return err
		}
	}

	*options = value

	return nil
}

func encodeFormat(writer Writer) (json.RawMessage, error) {
	format := writer.Format
	if format == "" {
		format = FormatCompact
	}

	if writer.Compact != nil && format != FormatCompact ||
		writer.Pretty != nil && format != FormatPretty ||
		writer.JSON != nil && format != FormatJSON {
		return nil, errors.New("format options must match the selected format")
	}

	switch format {
	case FormatCompact:
		formatter := DefaultFormatter()
		if writer.Compact != nil {
			formatter = *writer.Compact
		}

		return json.Marshal(map[string]Formatter{"compact": formatter})
	case FormatPretty:
		options := DefaultPrettyOptions()
		if writer.Pretty != nil {
			options = *writer.Pretty
		}

		return json.Marshal(map[string]PrettyOptions{"pretty": options})
	case FormatJSON:
		options := DefaultJSONOptions()
		if writer.JSON != nil {
			options = *writer.JSON
		}

		return json.Marshal(map[string]JSONOptions{"json": options})
	default:
		return nil, fmt.Errorf("invalid format %q", format)
	}
}

func decodeFormat(data []byte, writer *Writer) error {
	variants, err := decodeObject(data, "format")
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

			writer.Format, writer.Compact = FormatCompact, &formatter
		case "pretty":
			var options PrettyOptions
			if err := json.Unmarshal(raw, &options); err != nil {
				return fmt.Errorf("pretty format: %w", err)
			}

			writer.Format, writer.Pretty = FormatPretty, &options
		case "json":
			var options JSONOptions
			if err := json.Unmarshal(raw, &options); err != nil {
				return fmt.Errorf("JSON format: %w", err)
			}

			writer.Format, writer.JSON = FormatJSON, &options
		default:
			return fmt.Errorf("unknown format variant %q", name)
		}
	}

	return nil
}
