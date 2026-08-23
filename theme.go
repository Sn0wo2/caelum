package caelum

// RGB is a 24-bit color expressed as red, green and blue components.
type RGB struct{ R, G, B uint8 }

// dim darkens a color to roughly a quarter of its brightness, matching acta's
// `Styled::dimmed` (a 2-bit right shift on each channel).
func (c RGB) dim() RGB {
	return RGB{c.R >> 2, c.G >> 2, c.B >> 2}
}

// Theme is the full palette used to colorize a log line.
type Theme struct {
	Accent    RGB // brackets, separators, arrows
	Secondary RGB // field keys
	Text      RGB // timestamps, messages, field values
	Error     RGB
	Warn      RGB
	Info      RGB
	Debug     RGB
	Trace     RGB
}

func palette(accent, secondary, text, err, warn, info, debug, trace RGB) Theme {
	return Theme{accent, secondary, text, err, warn, info, debug, trace}
}

var (
	// ThemeCaelum is the default palette.
	ThemeCaelum = palette(
		RGB{91, 206, 250}, RGB{245, 169, 184}, RGB{255, 255, 255},
		RGB{255, 85, 85}, RGB{255, 200, 60}, RGB{91, 206, 250},
		RGB{245, 169, 184}, RGB{240, 240, 240},
	)

	ThemeMonokai = palette(
		RGB{102, 217, 239}, RGB{249, 38, 114}, RGB{248, 248, 242},
		RGB{255, 85, 85}, RGB{255, 200, 60}, RGB{102, 217, 239},
		RGB{249, 38, 114}, RGB{180, 180, 180},
	)

	ThemeDracula = palette(
		RGB{139, 233, 253}, RGB{255, 121, 198}, RGB{248, 248, 242},
		RGB{255, 85, 85}, RGB{255, 200, 60}, RGB{139, 233, 253},
		RGB{255, 121, 198}, RGB{180, 180, 180},
	)

	ThemeNord = palette(
		RGB{136, 192, 208}, RGB{163, 190, 140}, RGB{216, 222, 233},
		RGB{191, 97, 106}, RGB{235, 203, 139}, RGB{136, 192, 208},
		RGB{163, 190, 140}, RGB{180, 180, 180},
	)

	ThemeCatppuccinMocha = palette(
		RGB{137, 180, 250}, RGB{203, 166, 247}, RGB{205, 214, 244},
		RGB{243, 139, 168}, RGB{249, 226, 175}, RGB{137, 180, 250},
		RGB{203, 166, 247}, RGB{180, 180, 180},
	)

	ThemeGruvbox = palette(
		RGB{131, 165, 152}, RGB{254, 128, 25}, RGB{235, 219, 178},
		RGB{251, 73, 52}, RGB{250, 189, 47}, RGB{131, 165, 152},
		RGB{254, 128, 25}, RGB{180, 180, 180},
	)

	ThemeOneDark = palette(
		RGB{97, 175, 239}, RGB{198, 120, 221}, RGB{171, 178, 191},
		RGB{224, 108, 117}, RGB{229, 192, 123}, RGB{97, 175, 239},
		RGB{198, 120, 221}, RGB{180, 180, 180},
	)

	ThemeTokyoNight = palette(
		RGB{122, 162, 247}, RGB{187, 154, 247}, RGB{192, 202, 245},
		RGB{247, 118, 142}, RGB{224, 175, 104}, RGB{122, 162, 247},
		RGB{187, 154, 247}, RGB{180, 180, 180},
	)
)

// Icons are the decorative glyphs that wrap a log line.
type Icons struct {
	Name             string
	BracketOpen      string
	BracketClose     string
	TimeBracketOpen  string
	TimeBracketClose string
	Separator        string
	Arrow            string
}

// IconsUnicode is the default icon set, safe for any UTF-8 terminal.
var IconsUnicode = Icons{
	Name:             "unicode",
	BracketOpen:      "[",
	BracketClose:     "]",
	TimeBracketOpen:  "｢",
	TimeBracketClose: "｣",
	Separator:        "┇",
	Arrow:            ">",
}

// IconsNerd uses Powerline/Nerd Font glyphs. Requires a patched font.
var IconsNerd = Icons{
	Name:             "nerd",
	BracketOpen:      "", //
	BracketClose:     "", //
	TimeBracketOpen:  "",
	TimeBracketClose: "",
	Separator:        "┇",
	Arrow:            "", //
}

// LevelLabels is the text shown inside the level bracket for each severity.
type LevelLabels struct {
	Error, Warn, Info, Debug, Trace string
}

var (
	LabelsLong   = LevelLabels{"ERROR", " WARN", " INFO", "DEBUG", "TRACE"}
	LabelsMedium = LevelLabels{"ERR", "WRN", "INF", "DBG", "TRC"}
	LabelsShort  = LevelLabels{"E", "W", "I", "D", "T"}
)

// Style bundles the three visual dimensions that can be swapped at runtime.
type Style struct {
	Theme  Theme
	Icons  Icons
	Labels LevelLabels
}

// DefaultStyle mirrors acta's default: Caelum theme, Unicode icons, long labels.
func DefaultStyle() Style {
	return Style{Theme: ThemeCaelum, Icons: IconsUnicode, Labels: LabelsLong}
}
