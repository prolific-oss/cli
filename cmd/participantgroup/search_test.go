package participantgroup_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/participantgroup"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// groupResponse is a single-page response whose count matches the results, so
// the paging walk stops after one request.
func groupResponse(groups ...model.ParticipantGroup) *client.ListParticipantGroupsResponse {
	response := &client.ListParticipantGroupsResponse{Results: groups, JSONAPIMeta: &client.JSONAPIMeta{}}
	response.Meta.Count = len(groups)
	return response
}

func TestSearchPassesTheQueryAndPagingToTheClient(t *testing.T) {
	for _, tc := range []struct {
		name       string
		args       []string
		wantQuery  string
		wantLimit  int
		wantOffset int
	}{
		{
			name:      "defaults to the standard record limit",
			args:      []string{"search", "memory & attention", "-w", "ws"},
			wantQuery: "memory & attention",
			wantLimit: client.DefaultRecordLimit,
		},
		{
			name:      "joins an unquoted multi-word query",
			args:      []string{"search", "pilot", "cohort", "-w", "ws"},
			wantQuery: "pilot cohort",
			wantLimit: client.DefaultRecordLimit,
		},
		{
			name:      "--limit caps the first request",
			args:      []string{"search", "memory", "-w", "ws", "--limit", "5"},
			wantQuery: "memory",
			wantLimit: 5,
		},
		{
			name:      "--all fetches every match",
			args:      []string{"search", "memory", "-w", "ws", "--all"},
			wantQuery: "memory",
			wantLimit: client.DefaultRecordLimit,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			c.EXPECT().
				SearchParticipantGroups(tc.wantQuery, "ws", tc.wantLimit, tc.wantOffset).
				Return(groupResponse(model.ParticipantGroup{ID: "g1", Name: "Memory"}), nil)

			var output bytes.Buffer
			cmd := participantgroup.NewParticipantCommand(c, &output)
			cmd.SetArgs(tc.args)

			require.NoError(t, cmd.Execute())
			require.Contains(t, output.String(), "Memory")
		})
	}
}

func TestSearchRendersTheSelectedFormat(t *testing.T) {
	for _, tc := range []struct {
		name    string
		flags   []string
		want    string
		absent  string
		isExact bool
	}{
		{name: "table by default", want: "Showing 1 record of 1"},
		{name: "table", flags: []string{"-t"}, want: "Showing 1 record of 1"},
		{name: "non-interactive", flags: []string{"-n"}, want: "Showing 1 record of 1"},
		{name: "CSV", flags: []string{"-c", "--fields", "Name"}, want: "Name\n\"Memory, pilot\"\n", isExact: true},
		{name: "JSON is an array, not the API envelope", flags: []string{"-j"}, want: `"id": "id1"`, absent: "_links"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			c.EXPECT().
				SearchParticipantGroups("memory", "ws", client.DefaultRecordLimit, 0).
				Return(groupResponse(model.ParticipantGroup{ID: "id1", Name: "Memory, pilot"}), nil)

			var output bytes.Buffer
			cmd := participantgroup.NewSearchCommand(c, &output)
			cmd.SetArgs(append([]string{"memory", "-w", "ws"}, tc.flags...))

			require.NoError(t, cmd.Execute())
			if tc.isExact {
				require.Equal(t, tc.want, output.String())
			} else {
				require.Contains(t, output.String(), tc.want)
			}
			if tc.absent != "" {
				require.NotContains(t, output.String(), tc.absent)
			}
		})
	}
}

func TestSearchRendersAnEmptyJSONArray(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().
		SearchParticipantGroups("memory", "ws", client.DefaultRecordLimit, 0).
		Return(groupResponse(), nil)

	var output bytes.Buffer
	cmd := participantgroup.NewSearchCommand(c, &output)
	cmd.SetArgs([]string{"memory", "-w", "ws", "--json"})

	require.NoError(t, cmd.Execute())
	require.JSONEq(t, "[]", output.String())
}

func TestSearchReportsAPIErrors(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().
		SearchParticipantGroups("memory", "ws", client.DefaultRecordLimit, 0).
		Return(nil, errors.New("access denied"))

	var output bytes.Buffer
	cmd := participantgroup.NewSearchCommand(c, &output)
	cmd.SetArgs([]string{"memory", "-w", "ws"})

	require.ErrorContains(t, cmd.Execute(), "error: access denied")
	require.Empty(t, output.String())
}

func TestSearchRejectsInvalidInput(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{name: "no query", args: []string{}, want: "requires at least 1 arg"},
		{name: "blank query", args: []string{" ", "-w", "ws"}, want: "please provide a search query"},
		{name: "no workspace", args: []string{"memory"}, want: "please provide a workspace ID"},
		{name: "blank workspace", args: []string{"memory", "-w", " "}, want: "please provide a workspace ID"},
		{name: "negative limit", args: []string{"memory", "-w", "ws", "--limit", "-1"}, want: "limit must be greater than or equal to 0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			cmd := participantgroup.NewSearchCommand(c, &bytes.Buffer{})
			cmd.SetArgs(tc.args)

			require.ErrorContains(t, cmd.Execute(), tc.want)
		})
	}
}
