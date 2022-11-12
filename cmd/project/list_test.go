package project_test

import (
	"bufio"
	"bytes"
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

func TestNewEventTypeCommandCallsAPI(t *testing.T) {
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
	}

	c.
		EXPECT().
		GetProjects(gomock.Eq("991199")).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := project.NewListCommand("projects", c, writer)
	_ = cmd.Flags().Set("workspace", "991199")
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `ID      Title Description
123     Titan Project about moons
8889991 Beans Project about beans
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}
