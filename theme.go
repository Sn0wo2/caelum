package caelum

type RGB struct{ R, G, B uint8 }

func (c RGB) dim() RGB {
	return RGB{c.R >> 2, c.G >> 2, c.B >> 2}
}

type Theme struct {
	Accent    RGB
	Secondary RGB
	Text      RGB
	Error     RGB
	Warn      RGB
	Info      RGB
	Debug     RGB
	Trace     RGB
}

var (
	ThemeCaelum = Theme{
		RGB{91, 206, 250}, RGB{245, 169, 184}, RGB{255, 255, 255},
		RGB{255, 85, 85}, RGB{255, 200, 60}, RGB{91, 206, 250},
		RGB{245, 169, 184}, RGB{240, 240, 240},
	}

	ThemeMonokai = Theme{
		RGB{102, 217, 239}, RGB{249, 38, 114}, RGB{248, 248, 242},
		RGB{255, 85, 85}, RGB{255, 200, 60}, RGB{102, 217, 239},
		RGB{249, 38, 114}, RGB{180, 180, 180},
	}

	ThemeDracula = Theme{
		RGB{139, 233, 253}, RGB{255, 121, 198}, RGB{248, 248, 242},
		RGB{255, 85, 85}, RGB{255, 200, 60}, RGB{139, 233, 253},
		RGB{255, 121, 198}, RGB{180, 180, 180},
	}

	ThemeNord = Theme{
		RGB{136, 192, 208}, RGB{163, 190, 140}, RGB{216, 222, 233},
		RGB{191, 97, 106}, RGB{235, 203, 139}, RGB{136, 192, 208},
		RGB{163, 190, 140}, RGB{180, 180, 180},
	}

	ThemeCatppuccinMocha = Theme{
		RGB{137, 180, 250}, RGB{203, 166, 247}, RGB{205, 214, 244},
		RGB{243, 139, 168}, RGB{249, 226, 175}, RGB{137, 180, 250},
		RGB{203, 166, 247}, RGB{180, 180, 180},
	}

	ThemeGruvbox = Theme{
		RGB{131, 165, 152}, RGB{254, 128, 25}, RGB{235, 219, 178},
		RGB{251, 73, 52}, RGB{250, 189, 47}, RGB{131, 165, 152},
		RGB{254, 128, 25}, RGB{180, 180, 180},
	}

	ThemeOneDark = Theme{
		RGB{97, 175, 239}, RGB{198, 120, 221}, RGB{171, 178, 191},
		RGB{224, 108, 117}, RGB{229, 192, 123}, RGB{97, 175, 239},
		RGB{198, 120, 221}, RGB{180, 180, 180},
	}

	ThemeTokyoNight = Theme{
		RGB{122, 162, 247}, RGB{187, 154, 247}, RGB{192, 202, 245},
		RGB{247, 118, 142}, RGB{224, 175, 104}, RGB{122, 162, 247},
		RGB{187, 154, 247}, RGB{180, 180, 180},
	}
)

type Icons struct {
	Name             string
	BracketOpen      string
	BracketClose     string
	TimeBracketOpen  string
	TimeBracketClose string
	Separator        string
	Arrow            string
}

var IconsUnicode = Icons{
	Name:             "unicode",
	BracketOpen:      "[",
	BracketClose:     "]",
	TimeBracketOpen:  "｢",
	TimeBracketClose: "｣",
	Separator:        "┇",
	Arrow:            ">",
}

var IconsNerd = Icons{
	Name:             "nerd",
	BracketOpen:      "",
	BracketClose:     "",
	TimeBracketOpen:  "",
	TimeBracketClose: "",
	Separator:        "┇",
	Arrow:            "",
}

type LevelLabels struct {
	Error, Warn, Info, Debug, Trace string
}

var (
	LabelsLong   = LevelLabels{"ERROR", " WARN", " INFO", "DEBUG", "TRACE"}
	LabelsMedium = LevelLabels{"ERR", "WRN", "INF", "DBG", "TRC"}
	LabelsShort  = LevelLabels{"E", "W", "I", "D", "T"}
)

type Style struct {
	Theme  Theme
	Icons  Icons
	Labels LevelLabels
}

func DefaultStyle() Style {
	return Style{Theme: ThemeCaelum, Icons: IconsUnicode, Labels: LabelsLong}
}

func mergeStyle(override Style) Style {
	style := DefaultStyle()
	if override.Theme != (Theme{}) {
		style.Theme = override.Theme
	}
	if override.Icons != (Icons{}) {
		style.Icons = override.Icons
	}
	if override.Labels != (LevelLabels{}) {
		style.Labels = override.Labels
	}
	return style
}
