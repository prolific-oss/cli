package participantgroup_test

import (
	"bytes"
	"encoding/json"
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

func TestSearch(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		asJSON, empty, fail, noMeta bool
	}{
		{name: "table"}, {name: "empty table", empty: true}, {name: "table without metadata", noMeta: true}, {name: "json", asJSON: true}, {name: "empty json", asJSON: true, empty: true}, {name: "API error", fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			response := &client.ListParticipantGroupsResponse{Results: []model.ParticipantGroup{{ID: "g1", Name: "Memory"}}, JSONAPIMeta: &client.JSONAPIMeta{}}
			response.Meta.Count = 42
			response.JSONAPILinks = &client.JSONAPILinks{}
			response.Links.Next.Href = "https://api.prolific.com/api/v1/participant-groups/?search=memory&limit=10&offset=30"
			if tc.noMeta {
				response.JSONAPIMeta = nil
			}
			if tc.empty {
				response.Results = nil
			}
			var apiErr error
			if tc.fail {
				apiErr = errors.New("access denied")
			}
			c.EXPECT().SearchParticipantGroups("memory & attention", "ws", 10, 20).Return(response, apiErr)
			var output bytes.Buffer
			cmd := participantgroup.NewParticipantCommand(c, &output)
			argv := []string{"search", "memory & attention", "--workspace", "ws", "--limit", "10", "--offset", "20"}
			if tc.asJSON {
				argv = append(argv, "--json")
			}
			cmd.SetArgs(argv)
			err := cmd.Execute()
			if tc.fail {
				require.ErrorContains(t, err, "access denied")
				require.Empty(t, output.String())
				return
			}
			require.NoError(t, err)
			if tc.asJSON {
				var got map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(output.Bytes(), &got))
				require.JSONEq(t, `{"count":42}`, string(got["meta"]))
				var links client.JSONAPILinks
				require.NoError(t, json.Unmarshal(output.Bytes(), &links))
				require.Equal(t, response.Links.Next.Href, links.Links.Next.Href)
				if tc.empty {
					require.JSONEq(t, "[]", string(got["results"]))
				} else {
					require.Contains(t, string(got["results"]), "Memory")
				}
			} else {
				if !tc.empty {
					require.Contains(t, output.String(), "Memory")
				}
				switch {
				case tc.noMeta:
					require.NotContains(t, output.String(), "Showing")
				case tc.empty:
					require.Contains(t, output.String(), "Showing 0 records of 42")
				default:
					require.Contains(t, output.String(), "Showing 1 record of 42")
				}
			}
		})
	}
}

func TestSearchRejectsInvalidInput(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	for _, args := range [][]string{{}, {" "}, {" ", "  "}, {"query", "--workspace", ""}, {"query", "--workspace", "ws", "--limit", "0"}, {"query", "--workspace", "ws", "--offset", "-1"}} {
		c := mock_client.NewMockAPI(gomock.NewController(t))
		cmd := participantgroup.NewSearchCommand(c, &bytes.Buffer{})
		cmd.SetArgs(args)
		require.Error(t, cmd.Execute())
	}
}

func TestSearchJoinsMultiWordQuery(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().SearchParticipantGroups("pilot cohort", "ws", client.DefaultRecordLimit, client.DefaultRecordOffset).
		Return(&client.ListParticipantGroupsResponse{}, nil)
	cmd := participantgroup.NewSearchCommand(c, &bytes.Buffer{})
	cmd.SetArgs([]string{"pilot", "cohort", "--workspace", "ws", "-t"})
	require.NoError(t, cmd.Execute())
}

func TestSearchOutputFormats(t *testing.T) {
	for _, tc := range []struct {
		name  string
		flags []string
		csv   bool
	}{
		{name: "table", flags: []string{"-t"}},
		{name: "non-interactive", flags: []string{"-n"}},
		{name: "CSV", flags: []string{"-c"}, csv: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			response := &client.ListParticipantGroupsResponse{Results: []model.ParticipantGroup{{ID: "id1", Name: "Memory, pilot"}}, JSONAPIMeta: &client.JSONAPIMeta{}}
			response.Meta.Count = 42
			c.EXPECT().SearchParticipantGroups("memory", "ws", client.DefaultRecordLimit, client.DefaultRecordOffset).Return(response, nil)
			var output bytes.Buffer
			cmd := participantgroup.NewSearchCommand(c, &output)
			args := []string{"memory", "--workspace", "ws"}
			args = append(args, tc.flags...)
			args = append(args, "--fields", "Name")
			cmd.SetArgs(args)
			require.NoError(t, cmd.Execute())
			if tc.csv {
				require.Equal(t, "Name\n\"Memory, pilot\"\n", output.String())
			} else {
				require.Contains(t, output.String(), "Memory, pilot")
				require.NotContains(t, output.String(), "id1")
				require.Contains(t, output.String(), "Showing 1 record of 42")
			}
		})
	}
}
