package logger

import (
	"runtime/debug"
	"strings"

	"github.com/Sn0wo2/caelum/internal/native"
)

const modulePath = "github.com/Sn0wo2/caelum"

func (l *Logger) CaelumVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "devel"
	}

	for _, dependency := range info.Deps {
		if dependency.Path == modulePath {
			return strings.TrimPrefix(dependency.Version, "v")
		}
	}

	if info.Main.Path == modulePath {
		return strings.TrimPrefix(info.Main.Version, "v")
	}

	return "devel"
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
