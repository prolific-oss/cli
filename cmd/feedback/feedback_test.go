package feedback_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/feedback"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewFeedbackCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := feedback.NewFeedbackCommand(client, os.Stdout)

	use := "feedback"
	short := "View participant feedback for your studies"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}
