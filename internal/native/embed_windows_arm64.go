//go:build windows && arm64

package native

import "embed"

//go:embed windows_arm64/acta_ffi.dll
var embeddedNativeAssets embed.FS
