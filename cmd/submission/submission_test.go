package submission_test

import (
	"os"
	"testing"

	"github.com/benmatselby/prolificli/cmd/submission"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/golang/mock/gomock"
)

func TestNewSubmissionCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := submission.NewSubmissionCommand(client, os.Stdout)

	use := "submission"
	short := "Submission related commands"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}
