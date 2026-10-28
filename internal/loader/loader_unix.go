//go:build (linux || darwin) && (amd64 || arm64)

package loader

import (
	"fmt"

	"github.com/ebitengine/purego"
)

func Open(path string) (Loader, error) {
	libraryHandle, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		return nil, fmt.Errorf("open shared library: %w", err)
	}
	return handle(libraryHandle), nil
}

func (library handle) Lookup(name string) (uintptr, error) {
	address, err := purego.Dlsym(uintptr(library), name)
	if err != nil {
		return 0, fmt.Errorf("resolve symbol %q: %w", name, err)
	}

	return address, nil
}
