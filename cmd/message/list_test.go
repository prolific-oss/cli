package message_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/message"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := message.NewListCommand("list", c, os.Stdout)

	use := "list"
	short := "View all your messages"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewListCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "I am titanium"

	c.
		EXPECT().
		GetMessages(nil, nil).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := message.NewListCommand("list", c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewListCommandCallsTheAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	userID := "user-id"
	createdAfter := "2023-01-01"
	response := client.ListMessagesResponse{
		Results: []model.Message{
			{
				SenderID:        "sender-id",
				StudyID:         "study-id",
				ChannelID:       "channel-id",
				DatetimeCreated: "datetime-created",
				Body:            "body",
			},
		},
		JSONAPIMeta: &client.JSONAPIMeta{
			Meta: struct {
				Count int `json:"count"`
			}{
				Count: 10,
			},
		},
	}

	c.
		EXPECT().
		GetMessages(gomock.Any(), gomock.Any()).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := message.NewListCommand("list", c, writer)
	_ = cmd.Flags().Set("user_id", userID)
	_ = cmd.Flags().Set("created_after", createdAfter)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	actual := b.String()
	expected := `Sender ID Study ID Channel ID Datetime Created Body
sender-id study-id channel-id datetime-created body
`

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}
