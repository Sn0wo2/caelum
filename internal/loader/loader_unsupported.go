//go:build !((windows || linux || darwin) && (amd64 || arm64))

package loader

import (
	"fmt"
	"runtime"
)

func Open(string) (Loader, error) {
	return nil, fmt.Errorf("dynamic library loading is unsupported on %s/%s", runtime.GOOS, runtime.GOARCH)
}
