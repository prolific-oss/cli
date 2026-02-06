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

func TestNewSendGroupCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := message.NewSendGroupCommand("send-group", c, os.Stdout)

	use := "send-group"
	short := "Send a message to a participant group"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewSendGroupCommandCallsTheAPIWithStudy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	groupID := "group-id"
	studyID := "study-id"
	body := "Hello group"

	c.
		EXPECT().
		SendGroupMessage(groupID, body, &studyID).
		Return(nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := message.NewSendGroupCommand("send-group", c, writer)
	_ = cmd.Flags().Set("group", groupID)
	_ = cmd.Flags().Set("study", studyID)
	_ = cmd.Flags().Set("body", body)
	_ = cmd.Execute()

	writer.Flush()

	actual := b.String()
	expected := `Group ID Study ID Body
group-id study-id Hello group
`

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewSendGroupCommandCallsTheAPIWithoutStudy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	groupID := "group-id"
	body := "Hello group"

	c.
		EXPECT().
		SendGroupMessage(groupID, body, (*string)(nil)).
		Return(nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := message.NewSendGroupCommand("send-group", c, writer)
	_ = cmd.Flags().Set("group", groupID)
	_ = cmd.Flags().Set("body", body)
	_ = cmd.Execute()

	writer.Flush()

	actual := b.String()
	expected := `Group ID Study ID Body
group-id N/A      Hello group
`

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewSendGroupCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "group send failed"

	c.
		EXPECT().
		SendGroupMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New(errorMessage)).
		AnyTimes()

	cmd := message.NewSendGroupCommand("send-group", c, os.Stdout)
	_ = cmd.Flags().Set("group", "group-id")
	_ = cmd.Flags().Set("body", "body")
	err := cmd.Execute()

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
