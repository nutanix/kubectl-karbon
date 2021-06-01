package version

// Default build-time variable.
// These values are overridden via ldflags
var (
	Version      = "dev"
	Commit       = "none"
	Date         = "unknown"
	BuiltBy      = "unknown"
	OsName       = ""
	PlatformName = ""
)
