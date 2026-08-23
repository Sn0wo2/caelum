package caelum

import (
	"strconv"
	"strings"
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

	b.WriteString("\x1b[")
	switch d {
	case TrueColor:
		// 38;2;r;g;b (fg) or 48;2;r;g;b (bg)
		b.WriteByte(fgBg(background, '3', '4'))
		b.WriteString("8;2;")
		writeUint(b, c.R)
		b.WriteByte(';')
		writeUint(b, c.G)
		b.WriteByte(';')
		writeUint(b, c.B)
	case Ansi256:
		b.WriteByte(fgBg(background, '3', '4'))
		b.WriteString("8;5;")
		writeUint(b, rgbTo256(c))
	case Ansi16:
		// 16-color SGR codes: 30-37 (fg) / 40-47 (bg), plus 90-97 / 100-107
		// for the bright variants. rgbTo16 returns 0..15.
		code := rgbTo16(c)
		base := 30
		if background {
			base = 40
		}
		if code >= 8 {
			base += 60 // bright range
			code -= 8
		}
		writeUint(b, uint8(base+int(code)))
	}
	b.WriteByte('m')
	b.WriteString(text)
	b.WriteString(reset)
}

func fgBg(background bool, fg, bg byte) byte {
	if background {
		return bg
	}
	return fg
}

func writeUint(b *strings.Builder, v uint8) {
	var tmp [3]byte
	b.Write(strconv.AppendUint(tmp[:0], uint64(v), 10))
}

// rgbTo256 maps a 24-bit color onto the xterm 256-color palette. Grays collapse
// onto the 24-step grayscale ramp (232-255); everything else snaps to the
// nearest cell of the 6x6x6 color cube (16-231).
func rgbTo256(c RGB) uint8 {
	if c.R == c.G && c.G == c.B {
		if c.R < 8 {
			return 16
		}
		if c.R > 248 {
			return 231
		}
		return uint8(232 + (int(c.R)-8)*24/247)
	}
	r := cubeComponent(c.R)
	g := cubeComponent(c.G)
	b := cubeComponent(c.B)
	return uint8(16 + 36*r + 6*g + b)
}

func cubeComponent(v uint8) int {
	if v < 48 {
		return 0
	}
	if v < 115 {
		return 1
	}
	return (int(v) - 35) / 40
}

// basic16 holds the canonical RGB values of the 16 ANSI terminal colors, used
// to find the nearest match by squared Euclidean distance.
var basic16 = [16]RGB{
	{0, 0, 0}, {205, 0, 0}, {0, 205, 0}, {205, 205, 0},
	{0, 0, 238}, {205, 0, 205}, {0, 205, 205}, {229, 229, 229},
	{127, 127, 127}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
	{92, 92, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
}

func rgbTo16(c RGB) uint8 {
	best, bestDist := 0, 1<<31
	for i, p := range basic16 {
		dr := int(c.R) - int(p.R)
		dg := int(c.G) - int(p.G)
		db := int(c.B) - int(p.B)
		dist := dr*dr + dg*dg + db*db
		if dist < bestDist {
			best, bestDist = i, dist
		}
	}
	return uint8(best)
}
