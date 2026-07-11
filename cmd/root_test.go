package cmd_test

import (
	"testing"

	"github.com/prolific-oss/cli/cmd"
	"github.com/prolific-oss/cli/version"
)

func TestNewGitHubCommand(t *testing.T) {
	cmd := cmd.NewRootCommand()

	use := "prolific"
	short := "CLI application for retrieving data from the Prolific Platform"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewRootCommandVersion(t *testing.T) {
	original := version.Version
	defer func() { version.Version = original }()

	version.Version = "1.2.3"
	rootCmd := cmd.NewRootCommand()

	if rootCmd.Version != "1.2.3" {
		t.Errorf("expected cmd.Version %q, got %q", "1.2.3", rootCmd.Version)
	}
}
