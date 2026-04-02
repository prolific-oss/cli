package participantgroup_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/participantgroup"
	"github.com/prolific-oss/cli/mock_client"
)

const removeSuccessGroupID = "group-1"

func TestNewRemoveCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := participantgroup.NewRemoveCommand("remove", c, os.Stdout)

	if cmd.Use != "remove <group-id>" {
		t.Fatalf("expected use: 'remove <group-id>'; got %q", cmd.Use)
	}
	if cmd.Short != "Remove participants from a participant group" {
		t.Fatalf("unexpected short description: %q", cmd.Short)
	}
}

func TestRemoveCommandErrorsIfNoParticipants(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().RemoveParticipantGroupMembers(gomock.Any(), gomock.Any()).MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewRemoveCommand("remove", c, writer)
	cmd.SetArgs([]string{removeSuccessGroupID})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expected := "error: you must provide at least one participant ID via --participant-id or --file"
	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestRemoveCommandErrorsIfFileAndInlineIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().RemoveParticipantGroupMembers(gomock.Any(), gomock.Any()).MaxTimes(0)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "participants.csv")
	if err := os.WriteFile(filePath, []byte("part-1\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewRemoveCommand("remove", c, writer)
	_ = cmd.Flags().Set("file", filePath)
	_ = cmd.Flags().Set("participant-id", "part-1")
	cmd.SetArgs([]string{removeSuccessGroupID})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expected := "error: cannot use --file together with --participant-id"
	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestRemoveCommandRemovesByInlineIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		RemoveParticipantGroupMembers(
			gomock.Eq(removeSuccessGroupID),
			gomock.Eq([]string{"part-1", "part-2"}),
		).
		Return(nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewRemoveCommand("remove", c, writer)
	_ = cmd.Flags().Set("participant-id", "part-1")
	_ = cmd.Flags().Set("participant-id", "part-2")
	cmd.SetArgs([]string{removeSuccessGroupID})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	writer.Flush()

	expected := "Removed 2 participant(s) from group group-1\n"
	if b.String() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}

func TestRemoveCommandRemovesByFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		RemoveParticipantGroupMembers(
			gomock.Eq(removeSuccessGroupID),
			gomock.Eq([]string{"part-1", "part-2", "part-3"}),
		).
		Return(nil).
		MaxTimes(1)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "participants.csv")
	if err := os.WriteFile(filePath, []byte("part-1\npart-2\npart-3\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewRemoveCommand("remove", c, writer)
	_ = cmd.Flags().Set("file", filePath)
	cmd.SetArgs([]string{removeSuccessGroupID})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	writer.Flush()

	expected := "Removed 3 participant(s) from group group-1\n"
	if b.String() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}

func TestRemoveCommandReturnsErrorIfRemoveFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		RemoveParticipantGroupMembers(gomock.Any(), gomock.Any()).
		Return(errors.New("api error")).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewRemoveCommand("remove", c, writer)
	_ = cmd.Flags().Set("participant-id", "part-1")
	cmd.SetArgs([]string{removeSuccessGroupID})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expected := "error: api error"
	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestRemoveCommandFileSkipsBlankLines(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		RemoveParticipantGroupMembers(
			gomock.Eq(removeSuccessGroupID),
			gomock.Eq([]string{"part-1", "part-2"}),
		).
		Return(nil).
		MaxTimes(1)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "participants.csv")
	if err := os.WriteFile(filePath, []byte("part-1\n\n  \npart-2\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewRemoveCommand("remove", c, writer)
	_ = cmd.Flags().Set("file", filePath)
	cmd.SetArgs([]string{removeSuccessGroupID})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
