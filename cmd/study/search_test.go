package study_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/study"
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
			response := &client.ListStudiesResponse{Results: []model.Study{{ID: "s1", Name: "Memory", InternalName: "Pilot"}}, JSONAPIMeta: &client.JSONAPIMeta{}}
			response.Meta.Count = 42
			response.JSONAPILinks = &client.JSONAPILinks{}
			response.Links.Next.Href = "https://api.prolific.com/api/v1/studies/?search=memory&page=3"
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
			c.EXPECT().SearchStudies("memory & attention", "ws", 2).Return(response, apiErr)
			var output bytes.Buffer
			cmd := study.NewStudyCommand(c, &output)
			argv := []string{"search", "memory & attention", "--workspace", "ws", "--page", "2"}
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

				// The CLI owns the envelope: the API's meta block and link
				// relations must not reach our output.
				require.NotContains(t, got, "meta")
				require.NotContains(t, got, "_links")
				require.NotContains(t, output.String(), "api.prolific.com")

				require.JSONEq(t, "42", string(got["count"]))
				// Page 2 of a 200-per-page endpoint starts at record 200.
				require.JSONEq(t, "200", string(got["limit"]))
				require.JSONEq(t, "200", string(got["offset"]))

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
					// Without a meta block the counter falls back to the
					// records in hand rather than disappearing, so the table
					// always reports how much it is showing.
					require.Contains(t, output.String(), "Showing 1 record of 1")
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
	for _, args := range [][]string{{}, {"one", "two"}, {" "}, {"query", "--page", "0"}} {
		c := mock_client.NewMockAPI(gomock.NewController(t))
		cmd := study.NewSearchCommand(c, &bytes.Buffer{})
		cmd.SetArgs(args)
		require.Error(t, cmd.Execute())
	}
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
			response := &client.ListStudiesResponse{Results: []model.Study{{ID: "id1", Name: "Memory, pilot"}}, JSONAPIMeta: &client.JSONAPIMeta{}}
			response.Meta.Count = 42
			c.EXPECT().SearchStudies("memory", "ws", 1).Return(response, nil)
			var output bytes.Buffer
			cmd := study.NewSearchCommand(c, &output)
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
