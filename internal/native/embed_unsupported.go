//go:build !((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package native

import "embed"

var embeddedNativeAssets embed.FS
