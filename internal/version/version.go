package version

import (
	"fmt"
	"runtime"
)

var (
	Version   = "dev"
	Commit    = "none"
	Date      = "unknown"
	GoVersion = runtime.Version()
)

func Print() string {
	return fmt.Sprintf("replay %s (commit: %s, built: %s, %s/%s, %s)",
		Version, Commit, Date, runtime.GOOS, runtime.GOARCH, GoVersion)
}
