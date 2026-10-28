package caelum

import (
	"github.com/Sn0wo2/caelum/internal/config"
	"github.com/Sn0wo2/caelum/internal/loader"
)

type (
	Loader         = loader.Loader
	Config         = config.Config
	Level          = config.Level
	ColorDepth     = config.ColorDepth
	LevelAlignment = config.LevelAlignment
	Format         = config.Format
	PrettyOptions  = config.PrettyOptions
	JSONOptions    = config.JSONOptions
	Formatter      = config.Formatter
	Theme          = config.Theme
	RGB            = config.RGB
	Style          = config.Style
	Icons          = config.Icons
	LevelLabels    = config.LevelLabels
	Rotation       = config.Rotation
	FileConfig     = config.FileConfig
	WriterTarget   = config.WriterTarget
	AsyncMode      = config.AsyncMode
	Writer         = config.Writer
)

const (
	LevelError Level = config.LevelError
	LevelWarn  Level = config.LevelWarn
	LevelInfo  Level = config.LevelInfo
	LevelDebug Level = config.LevelDebug
	LevelTrace Level = config.LevelTrace
)

const (
	ColorDepthTrueColor ColorDepth = config.ColorDepthTrueColor
	ColorDepthANSI256   ColorDepth = config.ColorDepthANSI256
	ColorDepthANSI16    ColorDepth = config.ColorDepthANSI16
	ColorDepthNoColor   ColorDepth = config.ColorDepthNoColor
)

const (
	FormatCompact Format = config.FormatCompact
	FormatPretty  Format = config.FormatPretty
	FormatJSON    Format = config.FormatJSON
)

const (
	LevelAlignmentLeft  LevelAlignment = config.LevelAlignmentLeft
	LevelAlignmentRight LevelAlignment = config.LevelAlignmentRight
)

const (
	RotationNone     Rotation = config.RotationNone
	RotationRename   Rotation = config.RotationRename
	RotationCompress Rotation = config.RotationCompress
)

const (
	WriterTargetStdout      WriterTarget = config.WriterTargetStdout
	WriterTargetStderr      WriterTarget = config.WriterTargetStderr
	WriterTargetBuffered    WriterTarget = config.WriterTargetBuffered
	WriterTargetFile        WriterTarget = config.WriterTargetFile
	WriterTargetAsyncStdout WriterTarget = config.WriterTargetAsyncStdout
	WriterTargetAsyncStderr WriterTarget = config.WriterTargetAsyncStderr
)

const AsyncModeNative AsyncMode = config.AsyncModeNative

func ParseLevel(value string) (Level, error) {
	return config.ParseLevel(value)
}

func DefaultConfig() Config {
	return config.DefaultConfig()
}

func DefaultWriter() Writer {
	return config.DefaultWriter()
}

func DefaultFormatter() Formatter {
	return config.DefaultFormatter()
}

func DefaultPrettyOptions() PrettyOptions {
	return config.DefaultPrettyOptions()
}

func DefaultJSONOptions() JSONOptions {
	return config.DefaultJSONOptions()
}

func DefaultTheme() Theme {
	return config.DefaultTheme()
}

func ThemeActa() Theme {
	return config.ThemeActa()
}

func ThemeMonokai() Theme {
	return config.ThemeMonokai()
}

func ThemeDracula() Theme {
	return config.ThemeDracula()
}

func ThemeNord() Theme {
	return config.ThemeNord()
}

func ThemeCatppuccinMocha() Theme {
	return config.ThemeCatppuccinMocha()
}

func ThemeGruvbox() Theme {
	return config.ThemeGruvbox()
}

func ThemeOneDark() Theme {
	return config.ThemeOneDark()
}

func ThemeTokyoNight() Theme {
	return config.ThemeTokyoNight()
}

func DefaultIcons() Icons {
	return config.DefaultIcons()
}

func IconsNerd() Icons {
	return config.IconsNerd()
}

func DefaultLevelLabels() LevelLabels {
	return config.DefaultLevelLabels()
}

func LevelLabelsLong() LevelLabels {
	return config.LevelLabelsLong()
}

func LevelLabelsMedium() LevelLabels {
	return config.LevelLabelsMedium()
}

func LevelLabelsShort() LevelLabels {
	return config.LevelLabelsShort()
}

func DefaultStyle() Style {
	return config.DefaultStyle()
}
