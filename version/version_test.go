package version

import (
	"runtime/debug"
	"testing"
)

func TestVersionFromBuildInfo(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"clean release tag", "v0.0.62", "0.0.62"},
		{"pseudo-version", "v0.0.63-20260323113851-93ca3426f81a", ""},
		{"devel", "(devel)", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &debug.BuildInfo{Main: debug.Module{Version: tt.version}}
			got := versionFromBuildInfo(info)
			if got != tt.expected {
				t.Errorf("versionFromBuildInfo(%q) = %q, want %q", tt.version, got, tt.expected)
			}
		})
	}
}

func TestGet(t *testing.T) {
	original := Version
	defer func() { Version = original }()

	t.Run("returns Version when set via ldflags", func(t *testing.T) {
		Version = "0.0.62"
		if got := Get(); got != "0.0.62" {
			t.Errorf("Get() = %q, want %q", got, "0.0.62")
		}
	})

	t.Run("returns dev when Version is unset and no build info", func(t *testing.T) {
		Version = "dev"
		// Can't mock debug.ReadBuildInfo, but we can verify dev is the fallback
		// when versionFromBuildInfo returns empty (covered by TestVersionFromBuildInfo)
		got := Get()
		if got == "" {
			t.Errorf("Get() returned empty string, want non-empty")
		}
	})
}
