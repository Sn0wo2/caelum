//go:build darwin && amd64

package native

import "embed"

//go:embed darwin_amd64/libacta_ffi.dylib
var embeddedNativeAssets embed.FS
