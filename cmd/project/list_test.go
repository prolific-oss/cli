package project_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/project"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := project.NewListCommand("projects", c, os.Stdout)

	use := "projects"
	short := "Provide details about your projects"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewListCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.ListProjectsResponse{
		Results: []model.Project{
			{
				ID:          "123",
				Title:       "Titan",
				Description: "Project about moons",
			},
			{
				ID:          "8889991",
				Title:       "Beans",
				Description: "Project about beans",
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
		GetProjects(gomock.Eq("991199"), gomock.Eq(10), gomock.Eq(2)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := project.NewListCommand("projects", c, writer)
	_ = cmd.Flags().Set("workspace", "991199")
	_ = cmd.Flags().Set("limit", "10")
	_ = cmd.Flags().Set("offset", "2")
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `ID      Title Description
123     Titan Project about moons
8889991 Beans Project about beans

Showing 2 records of 10
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}

func TestNewListCommandHandlesErrorsFromTheCliParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "please provide a workspace ID"

	cmd := project.NewListCommand("list", c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewListCommandHandlesErrorsFromTheAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := "workspace-id"
	errorMessage := "API says no"

	c.
		EXPECT().
		GetProjects(gomock.Eq(workspaceID), client.DefaultRecordLimit, client.DefaultRecordOffset).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := project.NewListCommand("list", c, os.Stdout)
	_ = cmd.Flags().Set("workspace", workspaceID)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
