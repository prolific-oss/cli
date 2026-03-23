// Package version holds build-time version information for the CLI.
package version

import (
	"runtime/debug"
	"strings"
)

// GITCOMMIT is injected at build time via -ldflags (e.g. "0.0.62").
// Defaults to "dev" when not set.
var GITCOMMIT string = "dev"

// Get returns the CLI version. It prefers the value injected via ldflags,
// then falls back to the Go module version embedded in the binary (e.g. when
// installed via `go install github.com/prolific-oss/cli/...@vX.Y.Z`).
func Get() string {
	if GITCOMMIT != "dev" {
		return GITCOMMIT
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := versionFromBuildInfo(info); v != "" {
			return v
		}
	}
	return GITCOMMIT
}

// versionFromBuildInfo extracts a clean release version from build info,
// stripping the v prefix to match the ldflags convention. Returns empty string
// for pseudo-versions (vX.Y.Z-TIMESTAMP-HASH), "(devel)", or empty values.
func versionFromBuildInfo(info *debug.BuildInfo) string {
	v := strings.TrimPrefix(info.Main.Version, "v")
	if v == "" || v == "(devel)" || strings.Contains(v, "-") {
		return ""
	}
	return v
}
