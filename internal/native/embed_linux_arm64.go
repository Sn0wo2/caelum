//go:build linux && arm64

package native

import "embed"

//go:embed linux_arm64/libacta_ffi.so
var embeddedNativeAssets embed.FS
