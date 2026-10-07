package study_test

import (
	"bytes"
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

// studyResponse is a page of studies reporting total as the overall count, so
// a response whose count matches its results ends the paging walk.
func studyResponse(total int, studies ...model.Study) *client.ListStudiesResponse {
	response := &client.ListStudiesResponse{Results: studies, JSONAPIMeta: &client.JSONAPIMeta{}}
	response.Meta.Count = total
	return response
}

func TestSearchPassesTheQueryAndWorkspaceToTheClient(t *testing.T) {
	for _, tc := range []struct {
		name          string
		args          []string
		wantQuery     string
		wantWorkspace string
	}{
		{
			name:          "scoped to a workspace",
			args:          []string{"search", "memory & attention", "-w", "ws"},
			wantQuery:     "memory & attention",
			wantWorkspace: "ws",
		},
		{
			name:      "workspace is optional",
			args:      []string{"search", "memory"},
			wantQuery: "memory",
		},
		{
			name:      "joins an unquoted multi-word query",
			args:      []string{"search", "memory", "task"},
			wantQuery: "memory task",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)

			c := mock_client.NewMockAPI(gomock.NewController(t))
			c.EXPECT().
				SearchStudies(tc.wantQuery, tc.wantWorkspace, 1).
				Return(studyResponse(1, model.Study{ID: "s1", Name: "Memory"}), nil)

			var output bytes.Buffer
			cmd := study.NewStudyCommand(c, &output)
			cmd.SetArgs(tc.args)

			require.NoError(t, cmd.Execute())
			require.Contains(t, output.String(), "Memory")
		})
	}
}

func TestSearchWalksPagesUntilTheCountIsReached(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	gomock.InOrder(
		c.EXPECT().SearchStudies("memory", "ws", 1).
			Return(studyResponse(3, model.Study{ID: "s1", Name: "First"}, model.Study{ID: "s2", Name: "Second"}), nil),
		c.EXPECT().SearchStudies("memory", "ws", 2).
			Return(studyResponse(3, model.Study{ID: "s3", Name: "Third"}), nil),
	)

	var output bytes.Buffer
	cmd := study.NewSearchCommand(c, &output)
	cmd.SetArgs([]string{"memory", "-w", "ws", "--all"})

	require.NoError(t, cmd.Execute())
	require.Contains(t, output.String(), "Third")
	require.Contains(t, output.String(), "Showing 3 records of 3")
}

func TestSearchStopsAtTheRequestedLimit(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().SearchStudies("memory", "ws", 1).
		Return(studyResponse(50, model.Study{ID: "s1", Name: "First"}, model.Study{ID: "s2", Name: "Second"}), nil)

	var output bytes.Buffer
	cmd := study.NewSearchCommand(c, &output)
	cmd.SetArgs([]string{"memory", "-w", "ws", "--limit", "2"})

	require.NoError(t, cmd.Execute())
	require.Contains(t, output.String(), "Showing 2 records of 50")
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
			c.EXPECT().SearchStudies("memory", "ws", 1).
				Return(studyResponse(1, model.Study{ID: "id1", Name: "Memory, pilot"}), nil)

			var output bytes.Buffer
			cmd := study.NewSearchCommand(c, &output)
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
	c.EXPECT().SearchStudies("memory", "ws", 1).Return(studyResponse(0), nil)

	var output bytes.Buffer
	cmd := study.NewSearchCommand(c, &output)
	cmd.SetArgs([]string{"memory", "-w", "ws", "--json"})

	require.NoError(t, cmd.Execute())
	require.JSONEq(t, "[]", output.String())
}

func TestSearchReportsAPIErrors(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	c.EXPECT().SearchStudies("memory", "ws", 1).Return(nil, errors.New("access denied"))

	var output bytes.Buffer
	cmd := study.NewSearchCommand(c, &output)
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
		{name: "blank query", args: []string{" "}, want: "please provide a search query"},
		{name: "negative limit", args: []string{"memory", "--limit", "-1"}, want: "limit must be greater than or equal to 0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			cmd := study.NewSearchCommand(c, &bytes.Buffer{})
			cmd.SetArgs(tc.args)

			require.ErrorContains(t, cmd.Execute(), tc.want)
		})
	}
}
