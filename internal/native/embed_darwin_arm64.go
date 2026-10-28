//go:build darwin && arm64

package native

import "embed"

//go:embed darwin_arm64/libacta_ffi.dylib
var embeddedNativeAssets embed.FS
