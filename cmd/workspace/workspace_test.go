package workspace_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/workspace"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewWorkspaceCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := workspace.NewWorkspaceCommand(client, os.Stdout)

	use := "workspace"
	short := "Manage and view your workspaces"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}
