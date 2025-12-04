package project_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/project"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewCreateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := project.NewCreateCommand("create", client, os.Stdout)

	use := "create"
	short := "Create a project"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestCreateCommandErrorsIfNoDescription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		CreateProject(gomock.Any(), gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := project.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("title", "Test Project")
	_ = cmd.Flags().Set("workspace", "workspace-id")
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	expected := "description is required"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestCreateCommandCreatesProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	r := client.CreateProjectResponse{}
	userID := "user-123"

	workspaceID := "8888"
	description := "A test project"
	record := model.Project{
		ID:    "123123",
		Title: "Titan",
	}

	c.
		EXPECT().
		GetMe().
		Return(&client.MeResponse{ID: userID}, nil).
		MaxTimes(1)

	c.
		EXPECT().
		CreateProject(gomock.Eq(workspaceID), gomock.Eq(model.Project{
			Title:       record.Title,
			Description: description,
			Owner:       userID,
		})).
		Return(&r, nil).
		MaxTimes(1)

	r.ID = record.ID

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := project.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("title", record.Title)
	_ = cmd.Flags().Set("workspace", workspaceID)
	_ = cmd.Flags().Set("description", description)
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	expected := fmt.Sprintf("[ok] Created project: %v\n", record.ID)

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}

func TestCreateCommandReturnsErrorIfCreateProjectFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "Unable to create project, because the flim flam is broken"
	r := client.CreateProjectResponse{}
	userID := "user-123"

	workspaceID := "workspace-id"
	description := "A test project"
	record := model.Project{
		ID:    "123123",
		Title: "Titan",
	}

	c.
		EXPECT().
		GetMe().
		Return(&client.MeResponse{ID: userID}, nil).
		MaxTimes(1)

	c.
		EXPECT().
		CreateProject(gomock.Eq(workspaceID), gomock.Eq(model.Project{
			Title:       record.Title,
			Description: description,
			Owner:       userID,
		})).
		Return(nil, errors.New(errorMessage)).
		MaxTimes(1)

	r.ID = record.ID

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := project.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("title", record.Title)
	_ = cmd.Flags().Set("workspace", workspaceID)
	_ = cmd.Flags().Set("description", description)
	err := cmd.RunE(cmd, nil)

	expected := errorMessage

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestCreateCommandHandlesErrorIfNoWorkspaceProvided(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	r := client.CreateProjectResponse{}

	model := model.Project{
		ID:    "123123",
		Title: "Titan",
	}
	r.ID = model.ID

	c.
		EXPECT().
		CreateProject(gomock.Any(), gomock.Any()).
		Return(&r, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := project.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("title", model.Title)
	err := cmd.RunE(cmd, nil)

	expected := "workspace is required"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
