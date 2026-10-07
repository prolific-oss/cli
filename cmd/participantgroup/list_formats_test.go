package participantgroup_test

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/participantgroup"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
	"github.com/stretchr/testify/require"
)

// The list command gained the standard output flags when it moved onto the
// shared renderer; before that it could only ever print a table.
func TestListRendersTheSelectedFormat(t *testing.T) {
	for _, tc := range []struct {
		name    string
		flags   []string
		want    string
		isExact bool
	}{
		{name: "table by default", want: "Showing 1 record of 10"},
		{name: "non-interactive", flags: []string{"-n"}, want: "Showing 1 record of 10"},
		{name: "CSV", flags: []string{"-c", "--fields", "Name"}, want: "Name\nR.E.M. fans\n", isExact: true},
		{name: "JSON", flags: []string{"-j"}, want: `"name": "R.E.M. fans"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := &client.ListParticipantGroupsResponse{
				Results:     []model.ParticipantGroup{{ID: "1122", Name: "R.E.M. fans"}},
				JSONAPIMeta: &client.JSONAPIMeta{},
			}
			response.Meta.Count = 10

			c := mock_client.NewMockAPI(gomock.NewController(t))
			c.EXPECT().
				GetParticipantGroups("ws", client.DefaultRecordLimit, client.DefaultRecordOffset).
				Return(response, nil)

			var output bytes.Buffer
			cmd := participantgroup.NewListCommand("list", c, &output)
			cmd.SetArgs(append([]string{"-w", "ws"}, tc.flags...))

			require.NoError(t, cmd.Execute())
			if tc.isExact {
				require.Equal(t, tc.want, output.String())
			} else {
				require.Contains(t, output.String(), tc.want)
			}
		})
	}
}
