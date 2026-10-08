package filters_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	filters "github.com/prolific-oss/cli/cmd/filters"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func listResponse() client.ListFiltersResponse {
	return client.ListFiltersResponse{
		Results: []model.Filter{
			{
				FilterID:          "age",
				FilterTitle:       "Age",
				FilterDescription: "Filter by age",
				Question:          "How old are you?",
				Type:              "range",
				DataType:          "integer",
			},
			{
				FilterID:          "handedness",
				FilterTitle:       "Handedness",
				FilterDescription: "Filter by handedness",
				Question:          "Are you left or right handed?",
				Type:              "select",
				DataType:          "string",
				Choices: map[string]string{
					"1": "Right-handed",
					"2": "Left-handed",
					"3": "Ambidextrous",
				},
			},
		},
	}
}

// runList executes the command with the given arguments against a client that
// expects a single catalogue request for workspaceID.
func runList(t *testing.T, workspaceID string, args ...string) (string, error) {
	t.Helper()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := listResponse()
	c.EXPECT().GetFilters(workspaceID).Return(&response, nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewListCommand(c, w)
	cmd.SetArgs(args)
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	return b.String(), err
}

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := filters.NewListCommand(c, nil)

	assert.Equal(t, "list", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	for name, shorthand := range map[string]string{
		"json":      "j",
		"csv":       "c",
		"table":     "t",
		"fields":    "f",
		"workspace": "w",
	} {
		flag := cmd.Flags().Lookup(name)
		require.NotNil(t, flag, "expected a --%s flag", name)
		assert.Equal(t, shorthand, flag.Shorthand)
		assert.False(t, flag.Hidden, "--%s should not be hidden", name)
	}

	nonInteractive := cmd.Flags().Lookup("non-interactive")
	require.NotNil(t, nonInteractive)
	assert.Equal(t, "n", nonInteractive.Shorthand)
	assert.True(t, nonInteractive.Hidden, "--non-interactive should be hidden")
}

// The help text has to explain the absence of --limit and --offset, which
// every other list command offers.
func TestListCommandDocumentsThatTheCatalogueDoesNotPaginate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := filters.NewListCommand(mock_client.NewMockAPI(ctrl), nil)

	assert.Contains(t, cmd.Long, "does not paginate")
	assert.Nil(t, cmd.Flags().Lookup("limit"))
	assert.Nil(t, cmd.Flags().Lookup("offset"))
}

// Writing to a buffer stands in for a pipe: with no format flag the command
// must fall back to a table rather than trying to start a TUI.
func TestListFiltersRendersATableWhenNotATerminal(t *testing.T) {
	output, err := runList(t, "")

	assert.NoError(t, err)
	assert.Contains(t, output, "FilterID")
	assert.Contains(t, output, "age")
	assert.Contains(t, output, "Handedness")
	assert.Contains(t, output, "Showing 2 records of 2")
}

func TestListFiltersRendersATable(t *testing.T) {
	output, err := runList(t, "", "--table")

	assert.NoError(t, err)
	assert.Contains(t, output, "FilterID")
	assert.Contains(t, output, "handedness")
}

// -n is the hidden spelling of --table, kept for existing scripts.
func TestListFiltersNonInteractiveRendersATable(t *testing.T) {
	output, err := runList(t, "", "-n")

	assert.NoError(t, err)
	assert.Contains(t, output, "FilterID")
	assert.Contains(t, output, "handedness")
}

func TestListFiltersRendersCsv(t *testing.T) {
	output, err := runList(t, "", "--csv")

	assert.NoError(t, err)
	assert.Contains(t, output, "FilterID,Title,Type,DataType,Choices\n")
	assert.Contains(t, output, "age,Age,range,integer,0\n")
	assert.Contains(t, output, "handedness,Handedness,select,string,3\n")
	assert.NotContains(t, output, "Showing")
}

func TestListFiltersRendersTheChosenFields(t *testing.T) {
	output, err := runList(t, "", "--csv", "-f", "Title,Question")

	assert.NoError(t, err)
	assert.Contains(t, output, "Title,Question\n")
	assert.Contains(t, output, "Age,How old are you?\n")
}

func TestListFiltersRendersTheCLIJSONEnvelope(t *testing.T) {
	output, err := runList(t, "", "--json")
	assert.NoError(t, err)

	var envelope struct {
		Results []model.Filter `json:"results"`
		Count   int            `json:"count"`
		Limit   int            `json:"limit"`
		Offset  int            `json:"offset"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &envelope))

	assert.Len(t, envelope.Results, 2)
	assert.Equal(t, "age", envelope.Results[0].FilterID)
	assert.Equal(t, 2, envelope.Count)
	assert.Equal(t, 2, envelope.Limit)
	assert.Equal(t, 0, envelope.Offset)

	// The API's own envelope must not reach our output.
	assert.NotContains(t, output, "_links")
	assert.NotContains(t, output, "meta")
}

func TestListFiltersScopesToAWorkspace(t *testing.T) {
	output, err := runList(t, "ws-id", "--csv", "-w", "ws-id")

	assert.NoError(t, err)
	assert.Contains(t, output, "age,Age,range,integer,0\n")
}

func TestListFiltersWithNoResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.ListFiltersResponse{Results: []model.Filter{}}
	c.EXPECT().GetFilters("").Return(&response, nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := filters.NewListCommand(c, w)
	cmd.SetArgs([]string{"--json"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	assert.NoError(t, err)
	assert.Contains(t, b.String(), `"results": []`)
}

func TestListFiltersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().GetFilters("").Return(nil, assert.AnError)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewListCommand(c, w)
	cmd.SetArgs([]string{"-n"})

	assert.Error(t, cmd.Execute())
}
