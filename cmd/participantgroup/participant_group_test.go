package participantgroup_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/prolificli/cmd/participantgroup"
	"github.com/prolific-oss/prolificli/mock_client"
)

func TestNewStudyCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := participantgroup.NewParticipantCommand(client, os.Stdout)

	use := "participant"
	short := "Manage and view your participant groups"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}
