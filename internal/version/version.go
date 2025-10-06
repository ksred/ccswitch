package version

import (
	"fmt"
	"runtime"
)

// These variables are set during build time via ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

// Info holds version information
type Info struct {
	Version   string
	Commit    string
	BuildTime string
	GoVersion string
	OS        string
	Arch      string
}

// Get returns version information
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("ccswitch %s (commit: %s, built: %s, go: %s, %s/%s)",
		i.Version, i.Commit, i.BuildTime, i.GoVersion, i.OS, i.Arch)
}

// Short returns a short version string
func (i Info) Short() string {
	return fmt.Sprintf("ccswitch %s", i.Version)
}
