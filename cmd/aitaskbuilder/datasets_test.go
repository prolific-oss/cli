package aitaskbuilder_test

import (
	"os"
	"testing"

	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
)

func TestNewDatasetsCommand(t *testing.T) {
	c := setupMockClient(t)

	cmd := aitaskbuilder.NewDatasetsCommand(c, os.Stdout)

	use := "dataset"
	short := "Manage your datasets"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}

	// Check that subcommands are registered
	if !cmd.HasSubCommands() {
		t.Fatal("expected dataset command to have subcommands")
	}

	// Verify specific subcommands exist
	checkCmd := cmd.Commands()[0]
	if checkCmd.Use != "check" {
		t.Fatalf("expected first subcommand to be 'check', got '%s'", checkCmd.Use)
	}

	createCmd := cmd.Commands()[1]
	if createCmd.Use != "create" {
		t.Fatalf("expected second subcommand to be 'create', got '%s'", createCmd.Use)
	}
}
