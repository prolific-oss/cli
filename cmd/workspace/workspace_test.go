package workspace_test

import (
	"os"
	"testing"

	"github.com/benmatselby/prolificli/cmd/workspace"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/golang/mock/gomock"
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
