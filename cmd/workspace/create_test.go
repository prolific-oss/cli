package workspace_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/workspace"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
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

	expected := "title is required"

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
		ID:    "123123",
		Title: "Titan",
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

	expected := fmt.Sprintf("[ok] Created workspace: %v\n", model.ID)

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}

func TestCreateCommandHandlesFailureToCreateWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	r := client.CreateWorkspacesResponse{}

	model := model.Workspace{
		ID:    "123123",
		Title: "Titan",
	}
	r.ID = model.ID

	c.
		EXPECT().
		CreateWorkspace(gomock.Any()).
		Return(nil, errors.New("unable to create workspace")).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := workspace.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("title", model.Title)
	err := cmd.RunE(cmd, nil)

	expected := "unable to create workspace"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
