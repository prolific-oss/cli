package cmd_test

import (
	"testing"

	"github.com/prolific-oss/cli/cmd"
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

func TestNewRootCommandRegistersSkillFlag(t *testing.T) {
	root := cmd.NewRootCommand()

	flag := root.PersistentFlags().Lookup("skill")
	if flag == nil {
		t.Fatal("expected --skill persistent flag to be registered")
	}

	if flag.DefValue != "" {
		t.Fatalf("expected --skill default value to be empty, got %q", flag.DefValue)
	}
}
