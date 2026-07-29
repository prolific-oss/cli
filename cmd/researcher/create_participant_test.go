package researcher_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/researcher"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewCreateParticipantCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := researcher.NewCreateParticipantCommand(c, os.Stdout)

	use := "create-participant"
	short := "Create a test participant"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestCreateParticipantCallsClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	email := "test@example.com"
	participantID := "participant-123"

	response := client.CreateTestParticipantResponse{
		ParticipantID: participantID,
	}

	c.
		EXPECT().
		CreateTestParticipant(gomock.Eq(email)).
		Return(&response, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := researcher.NewCreateParticipantCommand(c, writer)
	_ = cmd.Flags().Set("email", email)
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("was not expecting error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	expected := fmt.Sprintf("Created test participant: %s\n", participantID)

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestCreateParticipantHandlesApiErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	email := "test@example.com"

	c.
		EXPECT().
		CreateTestParticipant(gomock.Eq(email)).
		Return(nil, errors.New("email already registered")).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := researcher.NewCreateParticipantCommand(c, writer)
	_ = cmd.Flags().Set("email", email)
	err := cmd.RunE(cmd, nil)

	expected := "email already registered"
	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
