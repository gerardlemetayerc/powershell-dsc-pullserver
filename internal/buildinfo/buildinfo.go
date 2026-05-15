package buildinfo

import "fmt"

// Values overridden at build time with: go build -ldflags "-X ..."
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// Short returns a compact version string suitable for logs.
func Short() string {
	return fmt.Sprintf("version=%s commit=%s buildDate=%s", Version, Commit, BuildDate)
}
