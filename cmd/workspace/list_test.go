package workspace_test

import (
	"bufio"
	"bytes"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/workspace"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := workspace.NewListCommand("workspaces", c, os.Stdout)

	use := "workspaces"
	short := "Provide details about your workspaces"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewEventTypeCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.ListWorkspacesResponse{
		Results: []model.Workspace{
			{
				ID:          "444",
				Title:       "Office",
				Description: "The office workspace",
			},
			{
				ID:          "555",
				Title:       "Home",
				Description: "The home workspace",
			},
		},
	}

	c.
		EXPECT().
		GetWorkspaces().
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := workspace.NewListCommand("workspaces", c, writer)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `ID  Title  Description
444 Office The office workspace
555 Home   The home workspace
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}
