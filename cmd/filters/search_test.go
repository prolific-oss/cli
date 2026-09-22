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

func ptr[T any](v T) *T { return &v }

func searchResponse() *client.SearchFiltersResponse {
	response := &client.SearchFiltersResponse{
		Results: []model.FilterSearchResult{
			{
				FilterID:    "job-title",
				Title:       "Job title",
				Description: "Select participants by occupation.",
				Question:    ptr("What is your job title?"),
				Category:    ptr("Employment"),
				Type:        "select",
				DataType:    "ChoiceID",
				Match: model.FilterSearchMatch{
					Fields:     []string{"choices"},
					Highlights: []model.FilterSearchHighlight{},
				},
				NumChoices: ptr(5),
				MatchedChoices: &model.FilterMatchingChoices{
					Matched:   4,
					Truncated: true,
					Results: []model.FilterChoiceSearchResult{
						{
							ID:             "100",
							Label:          "Software developers",
							NumChildren:    3,
							NumDescendants: 3,
							Match: model.FilterChoiceMatch{Highlights: []model.FilterSearchHighlight{
								{Field: "label", QueryTerm: "developers", MatchedText: "developers", Start: 9, End: 19},
							}},
						},
					},
				},
			},
			{
				FilterID:    "age",
				Title:       "Age",
				Description: "Filter by age",
				Type:        "range",
				DataType:    "integer",
				Min:         18,
				Max:         100,
				Match: model.FilterSearchMatch{
					Fields: []string{"title"},
					Highlights: []model.FilterSearchHighlight{
						{Field: "title", QueryTerm: "age", MatchedText: "Age", Start: 0, End: 3},
					},
				},
			},
		},
		JSONAPIMeta: &client.JSONAPIMeta{},
	}
	response.Meta.Count = 2
	return response
}

func TestNewSearchCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := filters.NewSearchCommand(c, nil)

	assert.Equal(t, "search <query>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	for name, shorthand := range map[string]string{"workspace": "w", "limit": "l", "all": "a", "json": "j", "csv": "c", "table": "t"} {
		flag := cmd.Flags().Lookup(name)
		require.NotNil(t, flag, name)
		assert.Equal(t, shorthand, flag.Shorthand)
	}

	for _, name := range []string{"no-pager", "fields"} {
		require.NotNil(t, cmd.Flags().Lookup(name), name)
	}
}

func TestSearchFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilters("software developers", "ws-1", filters.DefaultSearchLimit, client.DefaultRecordOffset).
		Return(searchResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"software", "developers", "-w", "ws-1"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	output := stripansi.Strip(b.String())

	assert.Contains(t, output, "Filters matching \"software developers\"\nShowing 2 records of 2\n\n")
	assert.Contains(t, output, "1. Job title\n   select · ChoiceID · Employment\n")
	assert.Contains(t, output, "   Filter ID    job-title\n")
	assert.Contains(t, output, "   Question     What is your job title?\n")
	assert.Contains(t, output, "   Choices      5 (4 matching)\n")
	assert.Contains(t, output, "100  Software developers  (+3 nested)\n")
	assert.Contains(t, output, "…and 3 more matching choices\n")
	assert.Contains(t, output, "   Matched on   choices\n")

	assert.Contains(t, output, "2. Age\n   range · integer\n")
	assert.Contains(t, output, "   Range        18 to 100\n")
	assert.Contains(t, output, "   Matched on   title\n")

	// A rule separates the two results.
	assert.Equal(t, 1, strings.Count(output, strings.Repeat("─", 60)+"\n"))
}

func TestSearchFiltersJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilters("developers", "", 10, 0).
		Return(searchResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"developers", "--json", "--limit", "10"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)

	var decoded []model.FilterSearchResult
	require.NoError(t, json.Unmarshal(b.Bytes(), &decoded))
	require.Len(t, decoded, 2)
	assert.Equal(t, "job-title", decoded[0].FilterID)
	assert.Equal(t, 4, decoded[0].MatchedChoices.Matched)
	assert.Equal(t, "age", decoded[1].FilterID)
}

func TestSearchFiltersNoResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := &client.SearchFiltersResponse{Results: []model.FilterSearchResult{}, JSONAPIMeta: &client.JSONAPIMeta{}}
	c.EXPECT().
		SearchFilters("zzz", "", filters.DefaultSearchLimit, client.DefaultRecordOffset).
		Return(response, nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"zzz"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	assert.Equal(t, "No filters found matching \"zzz\"\n", b.String())
}

func TestSearchFiltersNoResultsJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := &client.SearchFiltersResponse{Results: nil}
	c.EXPECT().
		SearchFilters("zzz", "", filters.DefaultSearchLimit, client.DefaultRecordOffset).
		Return(response, nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"zzz", "-j"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	assert.JSONEq(t, "[]", b.String())
}

func TestSearchFiltersAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilters("developers", "", filters.DefaultSearchLimit, client.DefaultRecordOffset).
		Return(nil, assert.AnError)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"developers"})
	err := cmd.Execute()

	assert.Error(t, err)
}

func TestSearchFiltersValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing query", args: []string{}, want: "requires at least 1 arg(s)"},
		{name: "blank query", args: []string{"   "}, want: "please provide a search query"},
		{name: "query too long", args: []string{string(make([]rune, 201))}, want: "search query must be at most 200 characters"},
		{name: "limit too low", args: []string{"dev", "--limit", "0"}, want: "limit must be greater than or equal to 1"},
		{name: "all with limit", args: []string{"dev", "--all", "--limit", "10"}, want: "[all limit] were all set"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			var b bytes.Buffer
			w := bufio.NewWriter(&b)

			cmd := filters.NewSearchCommand(c, w)
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

// pageOf builds a search response holding n placeholder results starting at
// index start, reporting the given total.
func pageOf(start, n, total int) *client.SearchFiltersResponse {
	response := &client.SearchFiltersResponse{JSONAPIMeta: &client.JSONAPIMeta{}}
	response.Meta.Count = total
	for i := start; i < start+n; i++ {
		response.Results = append(response.Results, model.FilterSearchResult{
			FilterID: fmt.Sprintf("filter-%d", i),
			Title:    fmt.Sprintf("Filter %d", i),
			Type:     "select",
			DataType: "ChoiceID",
		})
	}
	return response
}

func TestSearchFiltersLimitAbovePageSizeFetchesMultiplePages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	gomock.InOrder(
		c.EXPECT().SearchFilters("dev", "", 100, 0).Return(pageOf(0, 100, 300), nil),
		c.EXPECT().SearchFilters("dev", "", 50, 100).Return(pageOf(100, 50, 300), nil),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"dev", "--limit", "150"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	output := stripansi.Strip(b.String())
	assert.Contains(t, output, "Showing 150 records of 300. Use --limit or --all to see more\n")
	assert.Contains(t, output, "1. Filter 0\n")
	assert.Contains(t, output, "150. Filter 149\n")
	assert.NotContains(t, output, "filter-150")
}

func TestSearchFiltersAllFetchesUntilExhausted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	gomock.InOrder(
		c.EXPECT().SearchFilters("dev", "ws-1", 100, 0).Return(pageOf(0, 100, 230), nil),
		c.EXPECT().SearchFilters("dev", "ws-1", 100, 100).Return(pageOf(100, 100, 230), nil),
		c.EXPECT().SearchFilters("dev", "ws-1", 100, 200).Return(pageOf(200, 30, 230), nil),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"dev", "--all", "-w", "ws-1", "--json"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)

	var decoded []model.FilterSearchResult
	require.NoError(t, json.Unmarshal(b.Bytes(), &decoded))
	require.Len(t, decoded, 230)
	assert.Equal(t, "filter-0", decoded[0].FilterID)
	assert.Equal(t, "filter-229", decoded[229].FilterID)
}

func TestSearchFiltersErrorOnLaterPage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	gomock.InOrder(
		c.EXPECT().SearchFilters("dev", "", 100, 0).Return(pageOf(0, 100, 300), nil),
		c.EXPECT().SearchFilters("dev", "", 100, 100).Return(nil, assert.AnError),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"dev", "--all"})
	err := cmd.Execute()
	w.Flush()

	// Results stream as pages arrive, so the first page is already written
	// when the second fails; the error must still be reported.
	require.Error(t, err)
	output := stripansi.Strip(b.String())
	assert.Contains(t, output, "100. Filter 99\n")
	assert.NotContains(t, output, "Filter 100")
}

func TestSearchFiltersTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilters("developers", "", filters.DefaultSearchLimit, client.DefaultRecordOffset).
		Return(searchResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"developers", "--table"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	output := b.String()
	assert.Contains(t, output, "Rank")
	assert.Contains(t, output, "FilterID")
	assert.Contains(t, output, "MatchedOn")
	assert.Regexp(t, `1\s+job-title\s+Job title\s+select\s+ChoiceID\s+Employment\s+choices`, output)
	assert.Regexp(t, `2\s+age\s+Age\s+range\s+integer\s+title`, output)
	assert.Contains(t, output, "Showing 2 records of 2")
}

func TestSearchFiltersCSVWithFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilters("developers", "", filters.DefaultSearchLimit, client.DefaultRecordOffset).
		Return(searchResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"developers", "--csv", "--fields", "Rank,FilterID,Title"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	assert.Equal(t, "Rank,FilterID,Title\n1,job-title,Job title\n2,age,Age\n", b.String())
}

func TestSearchFiltersHeaderUsesFirstPageTotal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// The header must be written from the first page, before later pages
	// arrive, and reflect how many will be shown out of the total.
	gomock.InOrder(
		c.EXPECT().SearchFilters("dev", "", 100, 0).Return(pageOf(0, 100, 1000), nil),
		c.EXPECT().SearchFilters("dev", "", 20, 100).Return(pageOf(100, 20, 1000), nil),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"dev", "--limit", "120"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	output := stripansi.Strip(b.String())
	assert.True(t, strings.HasPrefix(output, "Filters matching \"dev\"\nShowing 120 records of 1000. Use --limit or --all to see more\n\n1. Filter 0\n"))
	assert.Contains(t, output, "120. Filter 119\n")
}
