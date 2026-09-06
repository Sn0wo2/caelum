package caelum

import (
	"fmt"
	"strings"

	"github.com/muesli/termenv"
)

type ColorDepth int

const (
	ColorAuto ColorDepth = iota
	TrueColor
	Ansi256
	Ansi16
	NoColor
)

const reset = "\x1b[0m"

func (d ColorDepth) paint(b *strings.Builder, c RGB, background bool, text string) {
	if d == NoColor || d == ColorAuto || text == "" {
		b.WriteString(text)
		return
	}

	var profile termenv.Profile
	switch d {
	case TrueColor:
		profile = termenv.TrueColor
	case Ansi256:
		profile = termenv.ANSI256
	case Ansi16:
		profile = termenv.ANSI
	default:
		b.WriteString(text)
		return
	}

	color := profile.Convert(termenv.RGBColor(fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)))
	fmt.Fprintf(b, "\x1b[%sm%s%s", color.Sequence(background), text, reset)
}
