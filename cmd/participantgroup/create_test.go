package participantgroup_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/participantgroup"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewCreateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := participantgroup.NewCreateCommand("create", c, os.Stdout)

	use := "create"
	short := "Create a participant group"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestCreateCommandErrorsIfNoName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		CreateParticipantGroup(gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("workspace", "workspace-id")
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	expected := "error: name is required"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestCreateCommandErrorsIfNoWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		CreateParticipantGroup(gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("name", "My Group")
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	expected := "error: workspace is required"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestCreateCommandCreatesParticipantGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := "8888"
	groupName := "My Group"
	description := "A test group"
	recordID := "group-123"

	r := client.CreateParticipantGroupResponse{}
	r.ID = recordID

	c.
		EXPECT().
		CreateParticipantGroup(gomock.Eq(model.CreateParticipantGroup{
			Name:        groupName,
			WorkspaceID: workspaceID,
			Description: description,
		})).
		Return(&r, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("name", groupName)
	_ = cmd.Flags().Set("workspace", workspaceID)
	_ = cmd.Flags().Set("description", description)
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	expected := fmt.Sprintf("Created participant group: %s\n", recordID)

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestCreateCommandReturnsErrorIfCreateFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "Unable to create participant group"
	workspaceID := "workspace-id"
	groupName := "My Group"

	c.
		EXPECT().
		CreateParticipantGroup(gomock.Eq(model.CreateParticipantGroup{
			Name:        groupName,
			WorkspaceID: workspaceID,
		})).
		Return(nil, errors.New(errorMessage)).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("name", groupName)
	_ = cmd.Flags().Set("workspace", workspaceID)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
