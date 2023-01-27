package participantgroup_test

import (
	"bufio"
	"bytes"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/participantgroup"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := participantgroup.NewListCommand("list", client, os.Stdout)

	use := "list"
	short := "Provide details about your participant groups"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewListCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	projectID := "777111999"

	response := client.ListParticipantGroupsResponse{
		Results: []model.ParticipantGroup{
			{
				ID:        "1122",
				Name:      "R.E.M. fans",
				ProjectID: projectID,
			},
			{
				ID:        "3344",
				Name:      "Radiohead fans",
				ProjectID: projectID,
			},
		},
	}

	c.
		EXPECT().
		GetParticipantGroups(gomock.Eq(projectID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := participantgroup.NewListCommand("list", c, writer)
	_ = cmd.Flags().Set("project", projectID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `ID   Name
1122 R.E.M. fans
3344 Radiohead fans
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}
