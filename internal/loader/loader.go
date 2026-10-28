package loader

import (
	"fmt"
	"io/fs"

	"github.com/Sn0wo2/caelum/internal/utils"
)

type Loader interface {
	Lookup(name string) (uintptr, error)
}

type handle uintptr

func Extract(assets fs.FS, name string) (string, error) {
	if !fs.ValidPath(name) {
		return "", fmt.Errorf("invalid embedded file path %q", name)
	}

	data, err := fs.ReadFile(assets, name)
	if err != nil {
		return "", fmt.Errorf("read embedded file %q: %w", name, err)
	}

	if len(data) == 0 {
		return "", fmt.Errorf("embedded file %q is empty", name)
	}

	return utils.CacheFile("caelum/libraries", name, data)
}
