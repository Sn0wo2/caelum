//go:build windows && (amd64 || arm64)

package loader

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func Open(path string) (Loader, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve library path: %w", err)
	}

	libraryHandle, err := windows.LoadLibraryEx(
		absolutePath,
		0,
		windows.LOAD_LIBRARY_SEARCH_DLL_LOAD_DIR|windows.LOAD_LIBRARY_SEARCH_DEFAULT_DIRS,
	)
	if err != nil {
		return nil, fmt.Errorf("LoadLibraryExW: %w", err)
	}

	return handle(libraryHandle), nil
}

func (library handle) Lookup(name string) (uintptr, error) {
	address, err := windows.GetProcAddress(windows.Handle(library), name)
	if err != nil {
		return 0, fmt.Errorf("resolve symbol %q: %w", name, err)
	}

	return address, nil
}
