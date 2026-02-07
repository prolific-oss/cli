package message_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/message"
	"github.com/prolific-oss/cli/mock_client"
)

const errBodyRequired = "error: body is required"

func TestNewMessageCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := message.NewMessageCommand(c, os.Stdout)

	use := "message"
	short := "Send and retrieve messages"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}
