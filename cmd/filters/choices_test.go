package filters_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/filters"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// choicesPage builds a page of choices numbered from first, reporting total
// as the API's meta.count.
func choicesPage(first, count, total int) *client.ListFilterChoicesResponse {
	response := &client.ListFilterChoicesResponse{JSONAPIMeta: &client.JSONAPIMeta{}}
	response.Meta.Count = total
	for i := first; i < first+count; i++ {
		response.Results = append(response.Results, model.FilterChoiceSearchResult{
			ID:    fmt.Sprintf("%d", i),
			Label: fmt.Sprintf("Choice %d", i),
		})
	}
	return response
}

func choicesResponse() *client.ListFilterChoicesResponse {
	response := &client.ListFilterChoicesResponse{
		Results: []model.FilterChoiceSearchResult{
			{ID: "0", Label: "Management Occupations", NumChildren: 4, NumDescendants: 476},
			{ID: "1016", Label: "Registered Nurses", ParentID: ptr("0"), NumChildren: 12, NumDescendants: 12},
		},
		JSONAPIMeta: &client.JSONAPIMeta{},
	}
	response.Meta.Count = 2
	return response
}

func TestNewChoicesCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := filters.NewChoicesCommand(mock_client.NewMockAPI(ctrl), nil)

	assert.Equal(t, "choices <filter-id>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	for name, shorthand := range map[string]string{
		"workspace": "w",
		"fields":    "f",
		"limit":     "l",
		"offset":    "o",
		"all":       "a",
		"json":      "j",
		"csv":       "c",
		"table":     "t",
	} {
		flag := cmd.Flags().Lookup(name)
		require.NotNil(t, flag, "expected a --%s flag", name)
		assert.Equal(t, shorthand, flag.Shorthand)
	}

	assert.Equal(t, "200", cmd.Flags().Lookup("limit").DefValue)
}

// The endpoints ignore ordering parameters, so offering sort flags would be
// a promise the API does not keep.
func TestChoicesCommandsOfferNoSortFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	list := filters.NewChoicesCommand(c, nil)
	assert.Nil(t, list.Flags().Lookup("sort"))
	assert.Contains(t, list.Long, "ordering")

	search := filters.NewChoicesSearchCommand(c, nil)
	assert.Nil(t, search.Flags().Lookup("sort"))
	assert.Contains(t, search.Long, "ordering")
}

func TestChoicesSearchCommandIsRegisteredUnderChoices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := filters.NewChoicesCommand(mock_client.NewMockAPI(ctrl), nil)

	var names []string
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}
	assert.Contains(t, names, "search")
}

func TestChoicesRequiresAFilterID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := filters.NewChoicesCommand(mock_client.NewMockAPI(ctrl), nil)
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	assert.Error(t, cmd.Execute())
}

func TestChoicesListRendersATable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		GetFilterChoices("job-title", "", client.FilterChoicesPageSize, client.DefaultRecordOffset).
		Return(choicesResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesCommand(c, w)
	cmd.SetArgs([]string{"job-title"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)
	output := b.String()
	assert.Regexp(t, `ID\s+Label`, output)
	assert.Regexp(t, `0\s+Management Occupations`, output)
	assert.Regexp(t, `1016\s+Registered Nurses`, output)
	assert.NotContains(t, output, "NumDescendants", "the hierarchy columns belong to the CSV")
	assert.Contains(t, output, "Showing 2 records of 2")
}

func TestChoicesListRendersCsvWithChosenFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		GetFilterChoices("job-title", "", client.FilterChoicesPageSize, client.DefaultRecordOffset).
		Return(choicesResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesCommand(c, w)
	cmd.SetArgs([]string{"job-title", "--csv", "-f", "ID,Label,ParentID"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)
	assert.Equal(t, "ID,Label,ParentID\n0,Management Occupations,\n1016,Registered Nurses,0\n", b.String())
}

func TestChoicesListRendersTheCLIJSONEnvelope(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		GetFilterChoices("job-title", "", client.FilterChoicesPageSize, client.DefaultRecordOffset).
		Return(choicesResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesCommand(c, w)
	cmd.SetArgs([]string{"job-title", "--json"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)

	var envelope struct {
		Results []model.FilterChoiceSearchResult `json:"results"`
		Count   int                              `json:"count"`
		Limit   int                              `json:"limit"`
		Offset  int                              `json:"offset"`
	}
	require.NoError(t, json.Unmarshal(b.Bytes(), &envelope))

	require.Len(t, envelope.Results, 2)
	assert.Equal(t, "0", envelope.Results[0].ID)
	assert.Equal(t, 2, envelope.Count)
	assert.Equal(t, 200, envelope.Limit)
	assert.Equal(t, 0, envelope.Offset)

	// The API's own envelope must not reach our output.
	assert.NotContains(t, b.String(), "_links")
}

// The endpoint caps a page at 100, but the CLI-wide default is 200, so the
// default ask spans two requests.
func TestChoicesListDefaultLimitSpansTwoPages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	gomock.InOrder(
		c.EXPECT().GetFilterChoices("job-title", "", client.FilterChoicesPageSize, client.DefaultRecordOffset).Return(choicesPage(0, 100, 4123), nil),
		c.EXPECT().GetFilterChoices("job-title", "", client.FilterChoicesPageSize, 100).Return(choicesPage(100, 100, 4123), nil),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesCommand(c, w)
	cmd.SetArgs([]string{"job-title", "--csv", "-f", "ID"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)
	assert.Equal(t, 201, strings.Count(b.String(), "\n"), "expected a header plus 200 rows")
}

// --all keeps fetching until the reported total is covered.
func TestChoicesListAllFetchesEveryPage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	gomock.InOrder(
		c.EXPECT().GetFilterChoices("job-title", "", client.FilterChoicesPageSize, client.DefaultRecordOffset).Return(choicesPage(0, 100, 230), nil),
		c.EXPECT().GetFilterChoices("job-title", "", client.FilterChoicesPageSize, 100).Return(choicesPage(100, 100, 230), nil),
		c.EXPECT().GetFilterChoices("job-title", "", client.FilterChoicesPageSize, 200).Return(choicesPage(200, 30, 230), nil),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesCommand(c, w)
	cmd.SetArgs([]string{"job-title", "--all", "--json"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)

	var envelope struct {
		Results []model.FilterChoiceSearchResult `json:"results"`
		Count   int                              `json:"count"`
	}
	require.NoError(t, json.Unmarshal(b.Bytes(), &envelope))
	assert.Len(t, envelope.Results, 230)
	assert.Equal(t, 230, envelope.Count)
}

func TestChoicesListScopesToAWorkspaceAndOffsets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		GetFilterChoices("job-title", "ws-1", client.FilterChoicesPageSize, 50).
		Return(choicesPage(50, 2, 4123), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesCommand(c, w)
	cmd.SetArgs([]string{"job-title", "-w", "ws-1", "-o", "50", "--csv", "-f", "ID"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)
	assert.Equal(t, "ID\n50\n51\n", b.String())
}

// A 404 covers both an unknown filter and one absent from the workspace; the
// API's own message is what the user needs to see.
func TestChoicesListSurfacesNotFoundUnchanged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	notFound := fmt.Errorf("The resource requested was not found.")
	c.EXPECT().GetFilterChoices("nope", "", client.FilterChoicesPageSize, client.DefaultRecordOffset).Return(nil, notFound)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesCommand(c, w)
	cmd.SetArgs([]string{"nope"})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "The resource requested was not found.")
}

func TestChoicesListWithNoChoices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	empty := &client.ListFilterChoicesResponse{Results: []model.FilterChoiceSearchResult{}, JSONAPIMeta: &client.JSONAPIMeta{}}
	c.EXPECT().GetFilterChoices("age", "", client.FilterChoicesPageSize, client.DefaultRecordOffset).Return(empty, nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesCommand(c, w)
	cmd.SetArgs([]string{"age", "--json"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)
	assert.Contains(t, b.String(), `"results": []`)
}

func searchChoicesResponse() *client.SearchFilterChoicesResponse {
	response := &client.SearchFilterChoicesResponse{
		Results: []model.FilterChoiceSearchResult{
			{
				ID:       "18873",
				Label:    "Obstetrics Nurse",
				ParentID: ptr("1016"),
				Matches: []model.FilterSearchHighlight{
					{Field: "label", QueryTerm: "nurse", MatchedText: "Nurse", Start: 11, End: 16},
				},
			},
		},
		JSONAPIMeta: &client.JSONAPIMeta{},
	}
	response.Meta.Count = 128
	return response
}

func TestNewChoicesSearchCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := filters.NewChoicesSearchCommand(mock_client.NewMockAPI(ctrl), nil)

	assert.Equal(t, "search <filter-id> <query>", cmd.Use)
	assert.Equal(t, "200", cmd.Flags().Lookup("limit").DefValue)

	for _, name := range []string{"workspace", "fields", "limit", "offset", "all", "json", "csv", "table"} {
		assert.NotNil(t, cmd.Flags().Lookup(name), "expected a --%s flag", name)
	}
}

func TestChoicesSearchRequiresFilterIDAndQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := filters.NewChoicesSearchCommand(mock_client.NewMockAPI(ctrl), nil)
	cmd.SetArgs([]string{"job-title"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	assert.Error(t, cmd.Execute())
}

// In a terminal the search view highlights why each choice matched, reusing
// the catalogue search's rendering.
func TestChoicesSearchRendersHighlightedRowsAtATerminal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilterChoices("job-title", "nurse", "", client.FilterChoicesPageSize, client.DefaultRecordOffset).
		Return(searchChoicesResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesSearchCommand(c, atTerminal(t, w))
	cmd.SetArgs([]string{"job-title", "nurse"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)
	output := stripansi.Strip(b.String())
	assert.Contains(t, output, `Choices in "job-title" matching "nurse"`)
	assert.Contains(t, output, "128 matching choices")
	assert.Contains(t, output, "18873        Obstetrics Nurse  (parent 1016)")
	assert.Contains(t, output, "Showing 1 record of 128")
}

// Piped, the same search has to produce a table instead.
func TestChoicesSearchRendersATableWhenNotATerminal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilterChoices("job-title", "nurse", "", client.FilterChoicesPageSize, client.DefaultRecordOffset).
		Return(searchChoicesResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesSearchCommand(c, w)
	cmd.SetArgs([]string{"job-title", "nurse"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)
	output := b.String()
	assert.Regexp(t, `ID\s+Label`, output)
	assert.Regexp(t, `18873\s+Obstetrics Nurse`, output)
	assert.NotContains(t, output, "Rank")
	assert.NotContains(t, output, "Choices in")
}

func TestChoicesSearchJoinsAMultiWordQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilterChoices("job-title", "registered nurse", "ws-1", client.FilterChoicesPageSize, client.DefaultRecordOffset).
		Return(searchChoicesResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesSearchCommand(c, w)
	cmd.SetArgs([]string{"job-title", "registered", "nurse", "-w", "ws-1", "--json"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)
	assert.Contains(t, b.String(), `"id": "18873"`)
}

func TestChoicesSearchSurfacesNotFoundUnchanged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilterChoices("nope", "nurse", "", client.FilterChoicesPageSize, client.DefaultRecordOffset).
		Return(nil, fmt.Errorf("The resource requested was not found."))

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesSearchCommand(c, w)
	cmd.SetArgs([]string{"nope", "nurse"})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "The resource requested was not found.")
}

func TestChoicesSearchWithNoMatches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	empty := &client.SearchFilterChoicesResponse{Results: []model.FilterChoiceSearchResult{}, JSONAPIMeta: &client.JSONAPIMeta{}}
	c.EXPECT().SearchFilterChoices("job-title", "zzz", "", client.FilterChoicesPageSize, client.DefaultRecordOffset).Return(empty, nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesSearchCommand(c, atTerminal(t, w))
	cmd.SetArgs([]string{"job-title", "zzz"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)
	assert.Equal(t, "No choices found in \"job-title\" matching \"zzz\"\n", b.String())
}

func TestChoicesValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "negative limit", args: []string{"job-title", "--limit", "-1"}, want: "limit must be greater than or equal to 0"},
		{name: "negative offset", args: []string{"job-title", "--offset", "-1"}, want: "offset must be greater than or equal to 0"},
		{name: "all with limit", args: []string{"job-title", "--all", "--limit", "10"}, want: "[all limit] were all set"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			var b bytes.Buffer
			w := bufio.NewWriter(&b)

			cmd := filters.NewChoicesCommand(c, w)
			cmd.SetArgs(tt.args)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			err := cmd.Execute()

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

// A blank query must not quietly fall back to listing the whole filter: the
// caller asked to search, and MinimumNArgs cannot catch a whitespace argument.
func TestChoicesSearchRejectsABlankQuery(t *testing.T) {
	for _, query := range []string{"", "   "} {
		t.Run(fmt.Sprintf("query=%q", query), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			// Neither endpoint may be called.
			c.EXPECT().SearchFilterChoices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			c.EXPECT().GetFilterChoices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			var b bytes.Buffer
			w := bufio.NewWriter(&b)

			cmd := filters.NewChoicesSearchCommand(c, w)
			cmd.SetArgs([]string{"job-title", query})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			err := cmd.Execute()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "please provide a search query")
		})
	}
}

func TestChoicesSearchRejectsAnOverlongQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().SearchFilterChoices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesSearchCommand(c, w)
	cmd.SetArgs([]string{"job-title", strings.Repeat("a", 201)})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "search query must be at most 200 characters")
}

func TestChoicesRejectsABlankFilterID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().GetFilterChoices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesCommand(c, w)
	cmd.SetArgs([]string{"   "})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "please provide a filter ID")
}

// --all is the same request as --limit 0, so the envelope has to report the
// window that was actually asked for rather than the unused default.
func TestChoicesListEnvelopeReportsTheRequestedWindow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	gomock.InOrder(
		c.EXPECT().GetFilterChoices("job-title", "", client.FilterChoicesPageSize, 0).Return(choicesPage(0, 100, 230), nil),
		c.EXPECT().GetFilterChoices("job-title", "", client.FilterChoicesPageSize, 100).Return(choicesPage(100, 100, 230), nil),
		c.EXPECT().GetFilterChoices("job-title", "", client.FilterChoicesPageSize, 200).Return(choicesPage(200, 30, 230), nil),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesCommand(c, w)
	cmd.SetArgs([]string{"job-title", "--all", "--json"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())
	require.NoError(t, err)

	var envelope struct {
		Results []model.FilterChoiceSearchResult `json:"results"`
		Count   int                              `json:"count"`
		Limit   int                              `json:"limit"`
		Offset  int                              `json:"offset"`
	}
	require.NoError(t, json.Unmarshal(b.Bytes(), &envelope))

	assert.Len(t, envelope.Results, 230)
	assert.Equal(t, 230, envelope.Count)
	assert.Equal(t, 0, envelope.Limit, "--all is an unbounded window, not the 200 default")
}

// The terminal view has to carry the same hierarchy facts as the columns do.
func TestChoicesListRowShowsParentAndChildCounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		GetFilterChoices("job-title", "", client.FilterChoicesPageSize, 0).
		Return(choicesResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewChoicesCommand(c, atTerminal(t, w))
	cmd.SetArgs([]string{"job-title"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())
	require.NoError(t, err)

	output := stripansi.Strip(b.String())
	assert.Contains(t, output, "0            Management Occupations  (4 children, +476 nested)")
	assert.Contains(t, output, "1016         Registered Nurses  (parent 0, 12 children)")
}
