package participantgroup_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/participantgroup"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewRemoveCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := participantgroup.NewRemoveCommand("remove", c, os.Stdout)

	use := "remove"
	short := "Remove participants from a participant group"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestRemoveCommandCallsClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	groupID := "6429b0ea05b2a24cac83c3a4"
	participantIDs := []string{"abc123", "def456"}

	remaining := &client.ViewParticipantGroupResponse{
		Results: []model.ParticipantGroupMembership{
			{
				ParticipantID:   "00000000000000001",
				DatetimeCreated: time.Date(2023, 01, 27, 19, 39, 0, 0, time.UTC),
			},
		},
	}

	c.
		EXPECT().
		RemoveParticipantsFromGroup(gomock.Eq(groupID), gomock.Eq(participantIDs)).
		Return(remaining, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewRemoveCommand("remove", c, writer)
	_ = cmd.RunE(cmd, append([]string{groupID}, participantIDs...))
	writer.Flush()

	expected := `Participant ID    Date added
00000000000000001 27-01-2023 19:39
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestRemoveCommandHandlesAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	groupID := "6429b0ea05b2a24cac83c3a4"
	participantIDs := []string{"abc123"}

	errorMessage := "participant group not found"

	c.
		EXPECT().
		RemoveParticipantsFromGroup(gomock.Eq(groupID), gomock.Eq(participantIDs)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewRemoveCommand("remove", c, writer)
	err := cmd.RunE(cmd, append([]string{groupID}, participantIDs...))
	writer.Flush()

	expected := fmt.Sprintf("error: %s", errorMessage)
	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestRemoveCommandWithFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	groupID := "6429b0ea05b2a24cac83c3a4"
	participantIDs := []string{"abc123", "def456"}

	dir := t.TempDir()
	filePath := filepath.Join(dir, "participants.csv")
	if err := os.WriteFile(filePath, []byte("abc123\ndef456\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	remaining := &client.ViewParticipantGroupResponse{
		Results: []model.ParticipantGroupMembership{
			{
				ParticipantID:   "00000000000000001",
				DatetimeCreated: time.Date(2023, 01, 27, 19, 39, 0, 0, time.UTC),
			},
		},
	}

	c.
		EXPECT().
		RemoveParticipantsFromGroup(gomock.Eq(groupID), gomock.Eq(participantIDs)).
		Return(remaining, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewRemoveCommand("remove", c, writer)
	_ = cmd.Flags().Set("file", filePath)
	_ = cmd.RunE(cmd, []string{groupID})
	writer.Flush()

	expected := `Participant ID    Date added
00000000000000001 27-01-2023 19:39
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestRemoveCommandFileAndPositionalIDsMutuallyExclusive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "participants.csv")
	if err := os.WriteFile(filePath, []byte("abc123\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewRemoveCommand("remove", c, writer)
	_ = cmd.Flags().Set("file", filePath)
	err := cmd.RunE(cmd, []string{"groupID", "positionalID"})

	expected := "error: cannot use both --file and positional participant IDs"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%v'\n", expected, err)
	}
}

func TestRemoveCommandNoIDsAndNoFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewRemoveCommand("remove", c, writer)
	err := cmd.RunE(cmd, []string{"groupID"})

	expected := "error: provide participant IDs as arguments or via --file"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%v'\n", expected, err)
	}
}

func TestRemoveCommandFileNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewRemoveCommand("remove", c, writer)
	_ = cmd.Flags().Set("file", "/non/existent/file.csv")
	err := cmd.RunE(cmd, []string{"groupID"})

	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}
