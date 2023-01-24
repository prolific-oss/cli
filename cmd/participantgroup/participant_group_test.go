package participantgroup_test

import (
	"os"
	"testing"

	"github.com/benmatselby/prolificli/cmd/participantgroup"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/golang/mock/gomock"
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
