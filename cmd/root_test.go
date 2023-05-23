package cmd_test

import (
	"testing"

	"github.com/prolific-oss/cli/cmd"
)

func TestNewGitHubCommand(t *testing.T) {
	cmd := cmd.NewRootCommand()

	use := "prolific"
	short := "CLI application for retrieving data from Prolific"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}
