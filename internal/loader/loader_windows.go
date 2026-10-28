//go:build windows && (amd64 || arm64)

package loader

import (
	"errors"
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

func Open(path string) (Loader, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve library path: %w", err)
	}

	pathPointer, err := syscall.UTF16PtrFromString(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("encode library path: %w", err)
	}

	libraryHandle, _, callErr := syscall.NewLazyDLL("kernel32.dll").NewProc("LoadLibraryExW").Call(
		uintptr(unsafe.Pointer(pathPointer)),
		0,
		0x00000100|0x00001000,
	)
	if libraryHandle == 0 {
		if errors.Is(callErr, syscall.Errno(0)) {
			callErr = syscall.GetLastError()
		}

		return nil, fmt.Errorf("LoadLibraryExW: %w", callErr)
	}

	return handle(libraryHandle), nil
}

func (library handle) Lookup(name string) (uintptr, error) {
	address, err := syscall.GetProcAddress(syscall.Handle(library), name)
	if err != nil {
		return 0, fmt.Errorf("resolve symbol %q: %w", name, err)
	}

	return address, nil
}
