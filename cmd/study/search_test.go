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
			c.EXPECT().SearchStudies("memory & attention", "ws", 1).Return(response, apiErr)
			var output bytes.Buffer
			cmd := study.NewStudyCommand(c, &output)
			// --limit 1 keeps this to a single page even though the API
			// reports more matches.
			argv := []string{"search", "memory & attention", "--workspace", "ws", "--limit", "1"}
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
				var got []model.Study
				require.NoError(t, json.Unmarshal(output.Bytes(), &got))
				if tc.empty {
					require.Empty(t, got)
				} else {
					require.Len(t, got, 1)
					require.Equal(t, "Memory", got[0].Name)
				}
				return
			}
			if !tc.empty {
				require.Contains(t, output.String(), "Memory")
			}
			switch {
			case tc.noMeta:
				// With no meta the counter falls back to what was collected.
				require.Contains(t, output.String(), "Showing 1 record of 1")
			case tc.empty:
				require.Contains(t, output.String(), "Showing 0 records of 42")
			default:
				require.Contains(t, output.String(), "Showing 1 record of 42")
			}
		})
	}
}

func TestSearchFollowsPagesUpToLimit(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	c := mock_client.NewMockAPI(gomock.NewController(t))

	first := &client.ListStudiesResponse{Results: []model.Study{{ID: "s1", Name: "One"}, {ID: "s2", Name: "Two"}}, JSONAPIMeta: &client.JSONAPIMeta{}}
	first.Meta.Count = 3
	second := &client.ListStudiesResponse{Results: []model.Study{{ID: "s3", Name: "Three"}}, JSONAPIMeta: &client.JSONAPIMeta{}}
	second.Meta.Count = 3

	c.EXPECT().SearchStudies("memory", "", 1).Return(first, nil)
	c.EXPECT().SearchStudies("memory", "", 2).Return(second, nil)

	var output bytes.Buffer
	cmd := study.NewSearchCommand(c, &output)
	cmd.SetArgs([]string{"memory", "--all", "-t"})
	require.NoError(t, cmd.Execute())

	require.Contains(t, output.String(), "One")
	require.Contains(t, output.String(), "Three")
	require.Contains(t, output.String(), "Showing 3 records of 3")
}

func TestSearchRejectsInvalidInput(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	for _, args := range [][]string{{}, {" "}, {" ", "  "}, {"query", "--limit", "-1"}, {"query", "--all", "--limit", "5"}} {
		c := mock_client.NewMockAPI(gomock.NewController(t))
		cmd := study.NewSearchCommand(c, &bytes.Buffer{})
		cmd.SetArgs(args)
		require.Error(t, cmd.Execute())
	}
}

func TestSearchJoinsMultiWordQuery(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().SearchStudies("memory attention", "", 1).Return(&client.ListStudiesResponse{}, nil)
	cmd := study.NewSearchCommand(c, &bytes.Buffer{})
	cmd.SetArgs([]string{"memory", "attention", "-t"})
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
			response := &client.ListStudiesResponse{Results: []model.Study{{ID: "id1", Name: "Memory, pilot"}}, JSONAPIMeta: &client.JSONAPIMeta{}}
			response.Meta.Count = 1
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
				require.Contains(t, output.String(), "Showing 1 record of 1")
			}
		})
	}
}
