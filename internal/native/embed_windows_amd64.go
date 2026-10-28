//go:build windows && amd64

package native

import "embed"

//go:embed windows_amd64/acta_ffi.dll
var embeddedNativeAssets embed.FS
