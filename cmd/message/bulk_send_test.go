package message_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/message"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewBulkSendCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := message.NewBulkSendCommand("bulk-send", c, os.Stdout)

	use := "bulk-send"
	short := "Send a message to multiple participants"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewBulkSendCommandCallsTheAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	ids := []string{"id1", "id2"}
	studyID := "bulk-study-id"
	body := "Hello participants"

	c.
		EXPECT().
		BulkSendMessage(ids, body, studyID).
		Return(nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := message.NewBulkSendCommand("bulk-send", c, writer)
	_ = cmd.Flags().Set("ids", "id1,id2")
	_ = cmd.Flags().Set("study", studyID)
	_ = cmd.Flags().Set("body", body)
	_ = cmd.Execute()

	writer.Flush()

	actual := b.String()
	expected := `Recipients Study ID      Body
2          bulk-study-id Hello participants
`

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewBulkSendCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "bulk send failed"

	c.
		EXPECT().
		BulkSendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New(errorMessage)).
		AnyTimes()

	cmd := message.NewBulkSendCommand("bulk-send", c, os.Stdout)
	_ = cmd.Flags().Set("ids", "id1")
	_ = cmd.Flags().Set("study", "bulk-study-id")
	_ = cmd.Flags().Set("body", "body")
	err := cmd.Execute()

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewBulkSendCommandValidatesEmptyIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := message.NewBulkSendCommand("bulk-send", c, os.Stdout)
	_ = cmd.Flags().Set("ids", "")
	_ = cmd.Flags().Set("study", "bulk-study-id")
	_ = cmd.Flags().Set("body", "body")
	err := cmd.Execute()

	expectedError := "error: at least one participant ID is required"
	if err == nil || err.Error() != expectedError {
		t.Fatalf("expected error: '%s'; got error: '%v'", expectedError, err)
	}
}
