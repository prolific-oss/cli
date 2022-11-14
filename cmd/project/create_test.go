package project_test

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/workspace"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewCreateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := workspace.NewCreateCommand("create", client, os.Stdout)

	use := "create"
	short := "Create a workspace"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestCreateCommandErrorsIfNoTitle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		CreateWorkspace(gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := workspace.NewCreateCommand("create", c, writer)
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	expected := "error: title is required"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}

func TestCreateCommandCreatesWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	r := client.CreateWorkspacesResponse{}

	model := model.Workspace{
		ID:                      "123123",
		Title:                   "Titan",
		NaivetyDistributionRate: 0,
	}
	r.ID = model.ID

	c.
		EXPECT().
		CreateWorkspace(gomock.Any()).
		Return(&r, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := workspace.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("title", model.Title)
	err := cmd.RunE(cmd, nil)

	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	expected := fmt.Sprintf("Created workspace: %v\n", model.ID)

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}
