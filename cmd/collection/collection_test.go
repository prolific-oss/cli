package collection_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/collection"
	"github.com/prolific-oss/cli/mock_client"
)

func setupMockClient(t *testing.T) *mock_client.MockAPI {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })
	return mock_client.NewMockAPI(ctrl)
}

func TestNewCollectionCommand(t *testing.T) {
	c := setupMockClient(t)

	cmd := collection.NewCollectionCommand(c, os.Stdout)

	use := "collection"
	short := "Manage and view your collections"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}

	// Check that subcommands are registered
	if !cmd.HasSubCommands() {
		t.Fatal("expected collection command to have subcommands")
	}

	// Verify create subcommand exists
	createCmd := cmd.Commands()[0]
	if createCmd.Use != "create" {
		t.Fatalf("expected first subcommand to be 'create', got '%s'", createCmd.Use)
	}
}
