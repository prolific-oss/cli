package invitation_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/invitation"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewCreateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := invitation.NewCreateCommand("create", c, os.Stdout)

	use := "create"
	short := "Create workspace invitations"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestCreateCommandErrorsIfNoWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		CreateInvitation(gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := invitation.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("email", "user@example.com")
	_ = cmd.Flags().Set("role", "WORKSPACE_COLLABORATOR")
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	expected := "error: workspace is required"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestCreateCommandErrorsIfNoEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		CreateInvitation(gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := invitation.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("workspace", "60d9aa5fa100c40b8c3fac61")
	_ = cmd.Flags().Set("role", "WORKSPACE_COLLABORATOR")
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	expected := "error: at least one email is required"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestCreateCommandErrorsIfNoRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		CreateInvitation(gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := invitation.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("workspace", "60d9aa5fa100c40b8c3fac61")
	_ = cmd.Flags().Set("email", "user@example.com")
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	expected := "error: role is required"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestCreateCommandErrorsIfInvalidRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		CreateInvitation(gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := invitation.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("workspace", "60d9aa5fa100c40b8c3fac61")
	_ = cmd.Flags().Set("email", "user@example.com")
	_ = cmd.Flags().Set("role", "INVALID_ROLE")
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `error: invalid role "INVALID_ROLE": must be one of WORKSPACE_ADMIN, WORKSPACE_COLLABORATOR`

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestCreateCommandCreatesInvitation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	r := client.CreateInvitationResponse{
		Invitations: []model.Invitation{
			{
				Association: "60d9aa5fa100c40b8c3fac61",
				Invitee: model.Invitee{
					Email: "user@example.com",
				},
				InvitedBy: "abc123",
				Status:    "INVITED",
				Role:      "WORKSPACE_COLLABORATOR",
			},
		},
	}

	c.
		EXPECT().
		CreateInvitation(gomock.Any()).
		Return(&r, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := invitation.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("workspace", "60d9aa5fa100c40b8c3fac61")
	_ = cmd.Flags().Set("email", "user@example.com")
	_ = cmd.Flags().Set("role", "WORKSPACE_COLLABORATOR")
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	expected := fmt.Sprintf("Invited %s as %s\n", "user@example.com", "WORKSPACE_COLLABORATOR")

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestCreateCommandCreatesMultipleInvitations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	r := client.CreateInvitationResponse{
		Invitations: []model.Invitation{
			{
				Invitee: model.Invitee{Email: "user1@example.com"},
				Role:    "WORKSPACE_ADMIN",
			},
			{
				Invitee: model.Invitee{Email: "user2@example.com"},
				Role:    "WORKSPACE_ADMIN",
			},
		},
	}

	c.
		EXPECT().
		CreateInvitation(gomock.Any()).
		Return(&r, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := invitation.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("workspace", "60d9aa5fa100c40b8c3fac61")
	_ = cmd.Flags().Set("email", "user1@example.com")
	_ = cmd.Flags().Set("email", "user2@example.com")
	_ = cmd.Flags().Set("role", "WORKSPACE_ADMIN")
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	expected := "Invited user1@example.com as WORKSPACE_ADMIN\nInvited user2@example.com as WORKSPACE_ADMIN\n"

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestCreateCommandHandlesFailureToCreateInvitation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		CreateInvitation(gomock.Any()).
		Return(nil, errors.New("unable to create invitation")).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := invitation.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("workspace", "60d9aa5fa100c40b8c3fac61")
	_ = cmd.Flags().Set("email", "user@example.com")
	_ = cmd.Flags().Set("role", "WORKSPACE_COLLABORATOR")
	err := cmd.RunE(cmd, nil)

	expected := "error: unable to create invitation"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
