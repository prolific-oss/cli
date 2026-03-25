// Package version holds build-time version information for the CLI.
package version

import (
	"runtime/debug"
	"strings"
)

// Version is injected at build time via -ldflags (e.g. "0.0.62").
// Defaults to "dev" when not set.
var Version string = "dev"

// Get returns the CLI version. It prefers the value injected via ldflags,
// then falls back to the Go module version embedded in the binary (e.g. when
// installed via `go install github.com/prolific-oss/cli/...@vX.Y.Z`).
func Get() string {
	if Version != "dev" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if version := versionFromBuildInfo(info); version != "" {
			return version
		}
	}
	return Version
}

func versionFromBuildInfo(info *debug.BuildInfo) string {
	version := strings.TrimPrefix(info.Main.Version, "v")
	if version == "" || version == "(devel)" || strings.Contains(version, "-") {
		return ""
	}
	return version
}
