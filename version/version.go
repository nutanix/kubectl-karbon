package version

import "runtime"

// Default build-time variable.
// These values are overridden via ldflags
var (
	Version      = "dev"
	Commit       = "none"
	Date         = "unknown"
	BuiltBy      = "unknown"
	OsName       = runtime.GOOS
	PlatformName = runtime.GOARCH
)
