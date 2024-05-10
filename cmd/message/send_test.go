package message_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/cmd/message"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/golang/mock/gomock"
)

func TestNewSendCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := message.NewSendCommand("send", c, os.Stdout)

	use := "send"
	short := "Send a message"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewSendCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "I am titanium"

	c.
		EXPECT().
		SendMessage(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New(errorMessage)).
		AnyTimes()

	cmd := message.NewSendCommand("send", c, os.Stdout)
	_ = cmd.Flags().Set("recipient", "recipient-id")
	_ = cmd.Flags().Set("study", "study-id")
	_ = cmd.Flags().Set("body", "body")
	err := cmd.Execute()

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewSendCommandCallsTheAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	recipientID := "recipient-id"
	studyID := "study-id"
	body := "body"

	c.
		EXPECT().
		SendMessage(body, recipientID, studyID).
		Return(nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := message.NewSendCommand("send", c, writer)
	_ = cmd.Flags().Set("recipient", recipientID)
	_ = cmd.Flags().Set("study", studyID)
	_ = cmd.Flags().Set("body", body)
	_ = cmd.Execute()

	writer.Flush()

	actual := b.String()
	expected := `Recipient ID Study ID Body
recipient-id study-id body
`

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}
