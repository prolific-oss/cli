package message_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/message"
	"github.com/benmatselby/prolificli/config"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
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

	receiver := "wibblet"
	createdAfter := "2023-01-01"
	response := client.ListMessagesResponse{
		Results: []model.Message{
			{
				SenderID:        "sender-id",
				StudyID:         "study-id",
				DatetimeCreated: time.Date(2023, 01, 27, 19, 39, 0, 0, time.UTC),
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
	_ = cmd.Flags().Set("user", receiver)
	_ = cmd.Flags().Set("created_after", createdAfter)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	actual := b.String()
	expected := fmt.Sprintf(`Sender ID Study ID Datetime Created Body
sender-id study-id 27-01-2023 19:39 body

---

View messages in the application: %s/messages/inbox
`, config.GetApplicationURL())

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewListCommandWithUnreadFlagCallsTheAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.ListUnreadMessagesResponse{
		Results: []model.UnreadMessage{
			{
				Sender:          "sender-id",
				DatetimeCreated: time.Date(2023, 01, 27, 19, 39, 0, 0, time.UTC),
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
		GetUnreadMessages().
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := message.NewListCommand("list", c, writer)
	_ = cmd.Flags().Set("unread", "true")
	err := cmd.RunE(cmd, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	writer.Flush()

	actual := b.String()
	expected := fmt.Sprintf(`Sender ID Datetime Created Body
sender-id 27-01-2023 19:39 body

---

View messages in the application: %s/messages/inbox
`, config.GetApplicationURL())

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewListCommandWithUnreadFlagAndOtherFlagsReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := message.NewListCommand("list", c, os.Stdout)
	_ = cmd.Flags().Set("unread", "true")
	_ = cmd.Flags().Set("user", "user-id") // Set another flag along with 'unread'
	err := cmd.RunE(cmd, nil)

	expectedError := `error: 'unread' cannot be used with any other flags`
	if err == nil || err.Error() != expectedError {
		t.Fatalf("expected error: '%s'; got error: '%v'", expectedError, err)
	}
}
