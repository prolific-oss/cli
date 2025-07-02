package submission_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/submission"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewSubmissionCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := submission.NewSubmissionCommand(client, os.Stdout)

	use := "submission"
	short := "Manage and view your study submissions"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}
