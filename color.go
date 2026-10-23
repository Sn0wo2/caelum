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

type ansiStyle string

func (s ansiStyle) paint(b *strings.Builder, text string) {
	if text == "" {
		return
	}

	b.WriteString(string(s))
	b.WriteString(text)

	if s != "" {
		b.WriteString(reset)
	}
}

type levelPalette struct {
	outline, body, transition, close ansiStyle
}

type palette struct {
	accent, text, secondary, accentDimmed, textDimmed ansiStyle
	timeOutline, timeBody                             ansiStyle
	levels                                            [5]levelPalette
}

func makePalette(d ColorDepth, t Theme) palette {
	var profile termenv.Profile

	switch d {
	case ColorAuto, NoColor:
		return palette{}
	case TrueColor:
		profile = termenv.TrueColor
	case Ansi256:
		profile = termenv.ANSI256
	case Ansi16:
		profile = termenv.ANSI
	default:
		return palette{}
	}

	color := func(c RGB, background bool) string {
		if d == TrueColor {
			code := 38
			if background {
				code = 48
			}

			return fmt.Sprintf("%d;2;%d;%d;%d", code, c.R, c.G, c.B)
		}

		return profile.Convert(termenv.RGBColor(fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B))).Sequence(background)
	}
	style := func(codes ...string) ansiStyle {
		return ansiStyle("\x1b[" + strings.Join(codes, ";") + "m")
	}
	background := func(c RGB) RGB {
		return RGB{
			c.R/5*4 + c.R%5*4/5,
			c.G/5*4 + c.G%5*4/5,
			c.B/5*4 + c.B%5*4/5,
		}
	}
	time := background(t.Accent)
	dim := time.dim()
	time = RGB{time.R - dim.R, time.G - dim.G, time.B - dim.B}

	p := palette{
		accent:       style(color(t.Accent, false)),
		text:         style(color(t.Text, false)),
		secondary:    style(color(t.Secondary, false)),
		accentDimmed: style(color(t.Accent, false), "2"),
		textDimmed:   style(color(t.Text, false), "2"),
		timeOutline:  style(color(time, false)),
		timeBody:     style("97", color(time, true)),
	}
	for i, c := range [5]RGB{t.Error, t.Warn, t.Info, t.Debug, t.Trace} {
		bg := background(c)
		p.levels[i] = levelPalette{
			outline:    style(color(c, false)),
			body:       style("97", color(bg, true)),
			transition: style(color(time, false), color(bg, true)),
			close:      style(color(bg, false)),
		}
	}

	return p
}
