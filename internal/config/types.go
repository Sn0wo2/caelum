package config

import (
	"encoding/json"
	"io"

	"github.com/Sn0wo2/caelum/internal/loader"
)

type Level uint32

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
	LevelTrace
)

type ColorDepth uint8

const (
	ColorDepthTrueColor ColorDepth = iota
	ColorDepthAnsi256
	ColorDepthAnsi16
	ColorDepthNoColor
)

type LevelAlignment string

const (
	LevelAlignmentLeft  LevelAlignment = "left"
	LevelAlignmentRight LevelAlignment = "right"
)

type Rotation string

const (
	RotationNone     Rotation = "none"
	RotationRename   Rotation = "rename"
	RotationCompress Rotation = "compress"
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
	Error     string          `json:"error"`
	Warn      string          `json:"warn"`
	Info      string          `json:"info"`
	Debug     string          `json:"debug"`
	Trace     string          `json:"trace"`
	Alignment *LevelAlignment `json:"-"`
}

type Style struct {
	Theme  Theme       `json:"theme"`
	Icons  Icons       `json:"icons"`
	Labels LevelLabels `json:"labels"`
}

type Formatter struct {
	TimeFormat *string     `json:"time_format"`
	PathWidth  *uint       `json:"path_width"`
	ShowPath   *bool       `json:"show_path"`
	ShowSpans  *bool       `json:"show_spans"`
	ColorDepth *ColorDepth `json:"color_depth"`
	Style      *Style      `json:"style"`
}

type PrettyOptions struct {
	Target     bool `json:"target"`
	File       bool `json:"file"`
	LineNumber bool `json:"line_number"`
}

type JsonOptions struct {
	Target       bool `json:"target"`
	File         bool `json:"file"`
	LineNumber   bool `json:"line_number"`
	CurrentSpan  bool `json:"current_span"`
	SpanList     bool `json:"span_list"`
	FlattenEvent bool `json:"flatten_event"`
}

type Format struct {
	Compact *Formatter
	Pretty  *PrettyOptions
	Json    *JsonOptions
}

type FileConfig struct {
	Path     string   `json:"path"`
	Rotation Rotation `json:"rotation"`
}

type AsyncMode struct {
	Native bool
}

type WriterTarget struct {
	Stdout      bool
	Stderr      bool
	Buffered    bool
	File        *FileConfig
	AsyncStdout *AsyncMode
	AsyncStderr *AsyncMode
}

type Writer struct {
	Format     Format
	ColorDepth *ColorDepth
	Stacktrace []Level
	Levels     []Level
	Target     WriterTarget
}

type Config struct {
	Loader         loader.Loader
	Writer         io.Writer
	Capacity       *uint
	NativeConfig   json.RawMessage
	Filter         string
	Writers        []Writer
	DisableCaller  bool
	CallerSkip     int
	Stacktrace     []Level
	filterExplicit bool
}

func DefaultTheme() Theme {
	return Theme{Accent: RGB{91, 206, 250}, Secondary: RGB{245, 169, 184}, Text: RGB{255, 255, 255}, Error: RGB{255, 85, 85}, Warn: RGB{255, 200, 60}, Info: RGB{91, 206, 250}, Debug: RGB{245, 169, 184}, Trace: RGB{240, 240, 240}}
}

func ThemeActa() Theme {
	return DefaultTheme()
}

func ThemeMonokai() Theme {
	return Theme{Accent: RGB{102, 217, 239}, Secondary: RGB{249, 38, 114}, Text: RGB{248, 248, 242}, Error: RGB{255, 85, 85}, Warn: RGB{255, 200, 60}, Info: RGB{102, 217, 239}, Debug: RGB{249, 38, 114}, Trace: RGB{180, 180, 180}}
}

func ThemeDracula() Theme {
	return Theme{Accent: RGB{139, 233, 253}, Secondary: RGB{255, 121, 198}, Text: RGB{248, 248, 242}, Error: RGB{255, 85, 85}, Warn: RGB{255, 200, 60}, Info: RGB{139, 233, 253}, Debug: RGB{255, 121, 198}, Trace: RGB{180, 180, 180}}
}

func ThemeNord() Theme {
	return Theme{Accent: RGB{136, 192, 208}, Secondary: RGB{163, 190, 140}, Text: RGB{216, 222, 233}, Error: RGB{191, 97, 106}, Warn: RGB{235, 203, 139}, Info: RGB{136, 192, 208}, Debug: RGB{163, 190, 140}, Trace: RGB{180, 180, 180}}
}

func ThemeCatppuccinMocha() Theme {
	return Theme{Accent: RGB{137, 180, 250}, Secondary: RGB{203, 166, 247}, Text: RGB{205, 214, 244}, Error: RGB{243, 139, 168}, Warn: RGB{249, 226, 175}, Info: RGB{137, 180, 250}, Debug: RGB{203, 166, 247}, Trace: RGB{180, 180, 180}}
}

func ThemeGruvbox() Theme {
	return Theme{Accent: RGB{131, 165, 152}, Secondary: RGB{254, 128, 25}, Text: RGB{235, 219, 178}, Error: RGB{251, 73, 52}, Warn: RGB{250, 189, 47}, Info: RGB{131, 165, 152}, Debug: RGB{254, 128, 25}, Trace: RGB{180, 180, 180}}
}

func ThemeOneDark() Theme {
	return Theme{Accent: RGB{97, 175, 239}, Secondary: RGB{198, 120, 221}, Text: RGB{171, 178, 191}, Error: RGB{224, 108, 117}, Warn: RGB{229, 192, 123}, Info: RGB{97, 175, 239}, Debug: RGB{198, 120, 221}, Trace: RGB{180, 180, 180}}
}

func ThemeTokyoNight() Theme {
	return Theme{Accent: RGB{122, 162, 247}, Secondary: RGB{187, 154, 247}, Text: RGB{192, 202, 245}, Error: RGB{247, 118, 142}, Warn: RGB{224, 175, 104}, Info: RGB{122, 162, 247}, Debug: RGB{187, 154, 247}, Trace: RGB{180, 180, 180}}
}

func DefaultIcons() Icons {
	return Icons{Name: "unicode", BracketOpen: "(", BracketClose: ")", TimeBracketOpen: "❬", TimeBracketClose: "❭", Separator: "┊", Arrow: "⟶", SpanJoin: "▸"}
}

func IconsNerd() Icons {
	return Icons{Name: "nerd", BracketOpen: "(", BracketClose: ")", TimeBracketOpen: "", TimeBracketClose: "", Separator: "", Arrow: "", SpanJoin: ""}
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

func DefaultFormatter() Formatter {
	pathWidth, showPath, showSpans, depth := uint(28), true, true, ColorDepthTrueColor
	style := DefaultStyle()

	return Formatter{PathWidth: &pathWidth, ShowPath: &showPath, ShowSpans: &showSpans, ColorDepth: &depth, Style: &style}
}

func DefaultPrettyOptions() PrettyOptions {
	return PrettyOptions{Target: true, File: true, LineNumber: true}
}

func DefaultJsonOptions() JsonOptions {
	return JsonOptions{FlattenEvent: true}
}

func DefaultWriter() Writer {
	return Writer{Target: WriterTarget{Stdout: true}}
}

func DefaultConfig() Config {
	return Config{Filter: "info", Writers: []Writer{DefaultWriter()}}
}
