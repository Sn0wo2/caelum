package config

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

func encodeTimeLayout(layout string) (string, error) {
	if !utf8.ValidString(layout) {
		return "", errors.New("time layout must be valid UTF-8")
	}

	var result strings.Builder

	literalStart := 0

	for i := 0; i < len(layout); {
		n, replacement, err := goTimeLayoutToken(layout, i)
		if err != nil {
			return "", err
		}

		if n == 0 {
			i++

			continue
		}

		writeChronoLiteral(&result, layout[literalStart:i])
		result.WriteString(replacement)

		i += n
		literalStart = i
	}

	writeChronoLiteral(&result, layout[literalStart:])

	return result.String(), nil
}

func goTimeLayoutToken(layout string, i int) (int, string, error) {
	rest := layout[i:]
	switch layout[i] {
	case 'J':
		if strings.HasPrefix(rest, "January") {
			return len("January"), "%B", nil
		}

		if strings.HasPrefix(rest, "Jan") && !startsWithASCIILower(rest[3:]) {
			return 3, "%b", nil
		}
	case 'M':
		if strings.HasPrefix(rest, "Monday") {
			return len("Monday"), "%A", nil
		}

		if strings.HasPrefix(rest, "Mon") && !startsWithASCIILower(rest[3:]) {
			return 3, "%a", nil
		}

		if strings.HasPrefix(rest, "MST") {
			return 0, "", errors.New("unsupported timezone token \"MST\" in Go time layout; remove it or use a layout without a timezone")
		}
	case '0':
		if len(rest) >= 2 && rest[1] >= '1' && rest[1] <= '6' {
			replacements := [...]string{"%m", "%d", "%I", "%M", "%S", "%y"}

			return 2, replacements[rest[1]-'1'], nil
		}

		if strings.HasPrefix(rest, "002") {
			return 3, "%j", nil
		}
	case '1':
		if strings.HasPrefix(rest, "15") {
			return 2, "%H", nil
		}

		return 1, "%-m", nil
	case '2':
		if strings.HasPrefix(rest, "2006") {
			return 4, "%Y", nil
		}

		return 1, "%-d", nil
	case '_':
		if strings.HasPrefix(rest, "_2006") {
			return 0, "", nil
		}

		if strings.HasPrefix(rest, "_2") {
			return 2, "%e", nil
		}

		if strings.HasPrefix(rest, "__2") {
			return 3, "%_j", nil
		}
	case '3':
		return 1, "%-I", nil
	case '4':
		return 1, "%-M", nil
	case '5':
		return 1, "%-S", nil
	case 'P':
		if strings.HasPrefix(rest, "PM") {
			return 2, "%p", nil
		}
	case 'p':
		if strings.HasPrefix(rest, "pm") {
			return 2, "%P", nil
		}
	case '-':
		for _, token := range [...]string{"-070000", "-07:00:00", "-0700", "-07:00", "-07"} {
			if strings.HasPrefix(rest, token) {
				return 0, "", fmt.Errorf("unsupported timezone token %q in Go time layout; remove it or use a layout without a timezone", token)
			}
		}
	case 'Z':
		for _, token := range [...]string{"Z070000", "Z07:00:00", "Z0700", "Z07:00", "Z07"} {
			if strings.HasPrefix(rest, token) {
				return 0, "", fmt.Errorf("unsupported timezone token %q in Go time layout; remove it or use a layout without a timezone", token)
			}
		}
	case '.', ',':
		if len(rest) >= 2 && (rest[1] == '0' || rest[1] == '9') {
			digit := rest[1]

			end := 2
			for end < len(rest) && rest[end] == digit {
				end++
			}

			if end == len(rest) || rest[end] < '0' || rest[end] > '9' {
				token := rest[:end]
				if digit == '9' {
					return 0, "", fmt.Errorf("unsupported variable-width fractional-second token %q in Go time layout; use a fixed-width run of 0s (3, 6, or 9 digits)", token)
				}

				if end != 4 && end != 7 && end != 10 {
					return 0, "", fmt.Errorf("unsupported fractional-second token %q in Go time layout; fixed-width fractions must use 3, 6, or 9 zeros", token)
				}

				precision := end - 1

				separator := string(layout[i])
				if separator == "." {
					return end, fmt.Sprintf("%%.%df", precision), nil
				}

				return end, fmt.Sprintf(",%%%df", precision), nil
			}
		}
	}

	return 0, "", nil
}

func startsWithASCIILower(value string) bool {
	return len(value) > 0 && value[0] >= 'a' && value[0] <= 'z'
}

func writeChronoLiteral(builder *strings.Builder, value string) {
	builder.WriteString(strings.ReplaceAll(value, "%", "%%"))
}

var chronoTimeLayouts = [...]struct {
	format string
	layout string
}{
	{".%3f", ".000"},
	{".%6f", ".000000"},
	{".%9f", ".000000000"},
	{".%f", ".000000000"},
	{",%3f", ",000"},
	{",%6f", ",000000"},
	{",%9f", ",000000000"},
	{",%f", ",000000000"},
	{"%.3f", ".000"},
	{"%.6f", ".000000"},
	{"%.9f", ".000000000"},
	{"%Y", "2006"},
	{"%y", "06"},
	{"%m", "01"},
	{"%-m", "1"},
	{"%d", "02"},
	{"%e", "_2"},
	{"%_d", "_2"},
	{"%-d", "2"},
	{"%j", "002"},
	{"%_j", "__2"},
	{"%H", "15"},
	{"%I", "03"},
	{"%-I", "3"},
	{"%M", "04"},
	{"%-M", "4"},
	{"%S", "05"},
	{"%-S", "5"},
	{"%p", "PM"},
	{"%P", "pm"},
	{"%a", "Mon"},
	{"%A", "Monday"},
	{"%b", "Jan"},
	{"%h", "Jan"},
	{"%B", "January"},
	{"%F", "2006-01-02"},
	{"%T", "15:04:05"},
	{"%R", "15:04"},
	{"%D", "01/02/06"},
	{"%r", "03:04:05 PM"},
	{"%%", "%"},
}

func decodeTimeLayout(format string) (string, error) {
	var layout, canonical strings.Builder

	for i := 0; i < len(format); {
		matched := false

		if format[i] == '%' || format[i] == '.' || format[i] == ',' {
			for _, token := range chronoTimeLayouts {
				if !strings.HasPrefix(format[i:], token.format) {
					continue
				}

				encoded, err := encodeTimeLayout(token.layout)
				if err != nil {
					return "", err
				}

				layout.WriteString(token.layout)
				canonical.WriteString(encoded)

				i += len(token.format)
				matched = true

				break
			}
		}

		if matched {
			continue
		}

		if format[i] == '%' {
			return "", fmt.Errorf("unsupported chrono time format %q near %q; use NativeConfig for formats without a Go layout equivalent", format, format[i:])
		}

		layout.WriteByte(format[i])
		canonical.WriteByte(format[i])
		i++
	}

	result := layout.String()

	encoded, err := encodeTimeLayout(result)
	if err != nil {
		return "", fmt.Errorf("chrono time format %q cannot be represented exactly as a Go layout: %w", format, err)
	}

	if encoded != canonical.String() {
		return "", fmt.Errorf("chrono time format %q cannot be represented exactly as a Go layout because Go would reinterpret a literal as a layout token", format)
	}

	return result, nil
}
