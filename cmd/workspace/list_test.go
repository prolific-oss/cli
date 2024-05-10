package workspace_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
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
		JSONAPIMeta: &client.JSONAPIMeta{
			Meta: struct {
				Count int `json:"count"`
			}{
				Count: 10,
			},
		},
	}

	c.
		EXPECT().
		GetWorkspaces(client.DefaultRecordLimit, client.DefaultRecordOffset).
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

Showing 2 records of 10
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}

func TestNewListCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "I am titanium"

	c.
		EXPECT().
		GetWorkspaces(client.DefaultRecordLimit, client.DefaultRecordOffset).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := workspace.NewListCommand("list", c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
