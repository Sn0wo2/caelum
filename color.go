package caelum

import (
	"fmt"
	"strings"

	"github.com/muesli/termenv"
)

// ColorDepth selects how RGB colors are encoded into ANSI escape sequences.
type ColorDepth int

const (
	// ColorAuto (the zero value) detects the depth from the writer.
	ColorAuto ColorDepth = iota
	// TrueColor emits 24-bit `38;2;r;g;b` sequences.
	TrueColor
	// Ansi256 maps RGB onto the xterm 256-color cube.
	Ansi256
	// Ansi16 maps RGB onto the 16 basic terminal colors.
	Ansi16
	// NoColor disables all escape sequences.
	NoColor
)

const reset = "\x1b[0m"

// paint appends text wrapped in an SGR sequence for the given color, honoring
// the depth. When background is true the color is applied to the background
// instead of the foreground. NoColor (and an unresolved ColorAuto) append the
// text untouched.
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
	b.WriteString("\x1b[")
	b.WriteString(color.Sequence(background))
	b.WriteByte('m')
	b.WriteString(text)
	b.WriteString(reset)
}
