package logger

import "github.com/Sn0wo2/caelum/internal/native"

const caelumVersion = "0.5.0"

func (l *Logger) CaelumVersion() string {
	return caelumVersion
}

func (l *Logger) ActaVersion() string {
	if version := l.api.ActaVersion(); version != "" {
		return version
	}

	if l.embedded {
		return native.EmbeddedActaVersion
	}

	return ""
}
