//go:build linux && amd64

package native

import "embed"

//go:embed linux_amd64/libacta_ffi.so
var embeddedNativeAssets embed.FS
