package project_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/cmd/project"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewViewCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := project.NewViewCommand("view", c, os.Stdout)

	use := "view"
	short := "Provide details about your project"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewViewCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	projectID := "991199"
	response := model.Project{
		ID:                      projectID,
		Title:                   "Titan",
		Description:             "Project about moons",
		Workspace:               "777777",
		Owner:                   "Dr. Who",
		NaivetyDistributionRate: 0.6,
		Users: []model.User{
			{
				ID:    "123",
				Name:  "Dr Who",
				Email: "dr@who.me",
			},
		},
	}

	c.
		EXPECT().
		GetProject(gomock.Eq(projectID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := project.NewViewCommand("view", c, writer)
	_ = cmd.RunE(cmd, []string{projectID})

	writer.Flush()

	expected := `Titan
Project about moons

Workspace:                 777777
Owner:                     Dr. Who
Naivety distribution rate: 0.6

Users:
ID  Name   Email
123 Dr Who dr@who.me

---

View project in the application: https://app.prolific.com/researcher/workspaces/projects/991199/
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}

func TestNewViewCommandHandlesErrorsFromTheCliParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "please provide a project ID"

	cmd := project.NewViewCommand("view", c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewViewCommandHandlesErrorsFromTheAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	projectID := "project-id"
	errorMessage := "API says no"

	c.
		EXPECT().
		GetProject(gomock.Eq(projectID)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := project.NewViewCommand("view", c, os.Stdout)
	err := cmd.RunE(cmd, []string{projectID})

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
