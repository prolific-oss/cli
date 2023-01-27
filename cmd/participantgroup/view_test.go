package participantgroup_test

import (
	"bufio"
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/participantgroup"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewViewCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := participantgroup.NewViewCommand("view", client, os.Stdout)

	use := "view"
	short := "Provide details about your participant group"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewViewCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	groupID := "66554488"

	response := client.ViewParticipantGroupResponse{
		Results: []model.ParticipantGroupMembership{
			{
				ParticipantID:   "00000000000000007",
				DatetimeCreated: time.Date(2023, 01, 27, 19, 39, 0, 0, time.UTC),
			},
			{
				ParticipantID:   "00000000000000006",
				DatetimeCreated: time.Date(2023, 01, 27, 19, 39, 0, 0, time.UTC),
			},
		},
	}

	c.
		EXPECT().
		GetParticipantGroup(gomock.Eq(groupID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewViewCommand("view", c, writer)
	_ = cmd.RunE(cmd, []string{groupID})

	writer.Flush()

	expected := `Participant ID    Date added
00000000000000007 27-01-2023 19:39
00000000000000006 27-01-2023 19:39
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}
