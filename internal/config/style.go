package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type ColorDepth string

const (
	ColorDepthTrueColor ColorDepth = "TrueColor"
	ColorDepthANSI256   ColorDepth = "Ansi256"
	ColorDepthANSI16    ColorDepth = "Ansi16"
	ColorDepthNoColor   ColorDepth = "NoColor"
)

type LevelAlignment string

const (
	LevelAlignmentLeft  LevelAlignment = "left"
	LevelAlignmentRight LevelAlignment = "right"
)

type RGB [3]uint8

type Theme struct {
	Accent    RGB `json:"accent"`
	Secondary RGB `json:"secondary"`
	Text      RGB `json:"text"`
	Error     RGB `json:"error"`
	Warn      RGB `json:"warn"`
	Info      RGB `json:"info"`
	Debug     RGB `json:"debug"`
	Trace     RGB `json:"trace"`
}

type Icons struct {
	Name             string `json:"name"`
	BracketOpen      string `json:"bracket_open"`
	BracketClose     string `json:"bracket_close"`
	TimeBracketOpen  string `json:"time_bracket_open"`
	TimeBracketClose string `json:"time_bracket_close"`
	Separator        string `json:"separator"`
	Arrow            string `json:"arrow"`
	SpanJoin         string `json:"span_join"`
}

type LevelLabels struct {
	Error     string         `json:"error"`
	Warn      string         `json:"warn"`
	Info      string         `json:"info"`
	Debug     string         `json:"debug"`
	Trace     string         `json:"trace"`
	Alignment LevelAlignment `json:"alignment,omitempty"`
}

type Style struct {
	Theme  Theme       `json:"theme"`
	Icons  Icons       `json:"icons"`
	Labels LevelLabels `json:"labels"`
}

func DefaultTheme() Theme {
	return Theme{
		Accent:    RGB{91, 206, 250},
		Secondary: RGB{245, 169, 184},
		Text:      RGB{255, 255, 255},
		Error:     RGB{255, 85, 85},
		Warn:      RGB{255, 200, 60},
		Info:      RGB{91, 206, 250},
		Debug:     RGB{245, 169, 184},
		Trace:     RGB{240, 240, 240},
	}
}

func ThemeActa() Theme {
	return DefaultTheme()
}

func ThemeMonokai() Theme {
	return Theme{
		Accent:    RGB{102, 217, 239},
		Secondary: RGB{249, 38, 114},
		Text:      RGB{248, 248, 242},
		Error:     RGB{255, 85, 85},
		Warn:      RGB{255, 200, 60},
		Info:      RGB{102, 217, 239},
		Debug:     RGB{249, 38, 114},
		Trace:     RGB{180, 180, 180},
	}
}

func ThemeDracula() Theme {
	return Theme{
		Accent:    RGB{139, 233, 253},
		Secondary: RGB{255, 121, 198},
		Text:      RGB{248, 248, 242},
		Error:     RGB{255, 85, 85},
		Warn:      RGB{255, 200, 60},
		Info:      RGB{139, 233, 253},
		Debug:     RGB{255, 121, 198},
		Trace:     RGB{180, 180, 180},
	}
}

func ThemeNord() Theme {
	return Theme{
		Accent:    RGB{136, 192, 208},
		Secondary: RGB{163, 190, 140},
		Text:      RGB{216, 222, 233},
		Error:     RGB{191, 97, 106},
		Warn:      RGB{235, 203, 139},
		Info:      RGB{136, 192, 208},
		Debug:     RGB{163, 190, 140},
		Trace:     RGB{180, 180, 180},
	}
}

func ThemeCatppuccinMocha() Theme {
	return Theme{
		Accent:    RGB{137, 180, 250},
		Secondary: RGB{203, 166, 247},
		Text:      RGB{205, 214, 244},
		Error:     RGB{243, 139, 168},
		Warn:      RGB{249, 226, 175},
		Info:      RGB{137, 180, 250},
		Debug:     RGB{203, 166, 247},
		Trace:     RGB{180, 180, 180},
	}
}

func ThemeGruvbox() Theme {
	return Theme{
		Accent:    RGB{131, 165, 152},
		Secondary: RGB{254, 128, 25},
		Text:      RGB{235, 219, 178},
		Error:     RGB{251, 73, 52},
		Warn:      RGB{250, 189, 47},
		Info:      RGB{131, 165, 152},
		Debug:     RGB{254, 128, 25},
		Trace:     RGB{180, 180, 180},
	}
}

func ThemeOneDark() Theme {
	return Theme{
		Accent:    RGB{97, 175, 239},
		Secondary: RGB{198, 120, 221},
		Text:      RGB{171, 178, 191},
		Error:     RGB{224, 108, 117},
		Warn:      RGB{229, 192, 123},
		Info:      RGB{97, 175, 239},
		Debug:     RGB{198, 120, 221},
		Trace:     RGB{180, 180, 180},
	}
}

func ThemeTokyoNight() Theme {
	return Theme{
		Accent:    RGB{122, 162, 247},
		Secondary: RGB{187, 154, 247},
		Text:      RGB{192, 202, 245},
		Error:     RGB{247, 118, 142},
		Warn:      RGB{224, 175, 104},
		Info:      RGB{122, 162, 247},
		Debug:     RGB{187, 154, 247},
		Trace:     RGB{180, 180, 180},
	}
}

func DefaultIcons() Icons {
	return Icons{
		Name:             "unicode",
		BracketOpen:      "(",
		BracketClose:     ")",
		TimeBracketOpen:  "❬",
		TimeBracketClose: "❭",
		Separator:        "┊",
		Arrow:            "⟶",
		SpanJoin:         "▸",
	}
}

func IconsNerd() Icons {
	return Icons{
		Name:             "nerd",
		BracketOpen:      "(",
		BracketClose:     ")",
		TimeBracketOpen:  "",
		TimeBracketClose: "",
		Separator:        "",
		Arrow:            "",
		SpanJoin:         "",
	}
}

func DefaultLevelLabels() LevelLabels {
	return LevelLabels{Error: "E", Warn: "W", Info: "I", Debug: "D", Trace: "T"}
}

func LevelLabelsLong() LevelLabels {
	return LevelLabels{Error: "ERROR", Warn: "WARN ", Info: "INFO ", Debug: "DEBUG", Trace: "TRACE"}
}

func LevelLabelsMedium() LevelLabels {
	return LevelLabels{Error: "ERR", Warn: "WRN", Info: "INF", Debug: "DBG", Trace: "TRC"}
}

func LevelLabelsShort() LevelLabels {
	return DefaultLevelLabels()
}

func DefaultStyle() Style {
	return Style{Theme: DefaultTheme(), Icons: DefaultIcons(), Labels: LevelLabelsLong()}
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
	fields, err := decodeObject(data, "theme")
	if err != nil {
		return err
	}

	value := DefaultTheme()
	for _, item := range []struct {
		name string
		dest *RGB
	}{
		{"accent", &value.Accent},
		{"secondary", &value.Secondary},
		{"text", &value.Text},
		{"error", &value.Error},
		{"warn", &value.Warn},
		{"info", &value.Info},
		{"debug", &value.Debug},
		{"trace", &value.Trace},
	} {
		if _, err := decodeField(fields, item.name, item.dest); err != nil {
			return err
		}
	}

	*theme = value

	return nil
}

func (icons *Icons) UnmarshalJSON(data []byte) error {
	fields, err := decodeObject(data, "icons")
	if err != nil {
		return err
	}

	value := DefaultIcons()
	for _, item := range []struct {
		name string
		dest *string
	}{
		{"name", &value.Name},
		{"bracket_open", &value.BracketOpen},
		{"bracket_close", &value.BracketClose},
		{"time_bracket_open", &value.TimeBracketOpen},
		{"time_bracket_close", &value.TimeBracketClose},
		{"separator", &value.Separator},
		{"arrow", &value.Arrow},
		{"span_join", &value.SpanJoin},
	} {
		if _, err := decodeField(fields, item.name, item.dest); err != nil {
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

func (style *Style) UnmarshalJSON(data []byte) error {
	fields, err := decodeObject(data, "style")
	if err != nil {
		return err
	}

	value := DefaultStyle()
	if _, err := decodeField(fields, "theme", &value.Theme); err != nil {
		return err
	}

	if _, err := decodeField(fields, "icons", &value.Icons); err != nil {
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

func decodeLevelLabels(data []byte, defaults LevelLabels) (LevelLabels, error) {
	fields, err := decodeObject(data, "level labels")
	if err != nil {
		return LevelLabels{}, err
	}

	for _, item := range []struct {
		name string
		dest *string
	}{
		{"error", &defaults.Error},
		{"warn", &defaults.Warn},
		{"info", &defaults.Info},
		{"debug", &defaults.Debug},
		{"trace", &defaults.Trace},
	} {
		if _, err := decodeField(fields, item.name, item.dest); err != nil {
			return LevelLabels{}, err
		}
	}

	alignment, _, err := decodeOptionalField[LevelAlignment](fields, "alignment")
	if err != nil {
		return LevelLabels{}, err
	}

	if alignment != nil {
		defaults.Alignment = *alignment
	}

	return defaults, nil
}
