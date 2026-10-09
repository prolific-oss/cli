package filters_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

// terminalWriter stands in for a terminal destination, so the tests can
// exercise the output a user at a terminal sees without attaching the test
// process to a pty.
type terminalWriter struct{ io.Writer }

func (terminalWriter) IsTerminal() bool { return true }

// atTerminal presents w as a terminal and disables the pager, so the output a
// user would see lands in the buffer rather than in less.
func atTerminal(t *testing.T, w io.Writer) io.Writer {
	t.Helper()
	t.Setenv("PROLIFIC_PAGER", "")
	return terminalWriter{w}
}

// decodeSearchEnvelope reads the CLI envelope --json emits and returns the
// records inside it.
func decodeSearchEnvelope(t *testing.T, raw []byte) []model.FilterSearchResult {
	t.Helper()

	var envelope struct {
		Results []model.FilterSearchResult `json:"results"`
		Count   int                        `json:"count"`
		Limit   int                        `json:"limit"`
		Offset  int                        `json:"offset"`
	}
	require.NoError(t, json.Unmarshal(raw, &envelope))

	// The API's own envelope must never reach our output.
	require.NotContains(t, string(raw), "_links")

	return envelope.Results
}

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
				Matches:     []model.FilterSearchHighlight{},
				Choices: &model.FilterSearchChoices{
					Total:     5,
					Matched:   4,
					Truncated: true,
					Results: []model.FilterChoiceSearchResult{
						{
							ID:             "100",
							Label:          "Software developers",
							NumChildren:    3,
							NumDescendants: 3,
							Matches: []model.FilterSearchHighlight{
								{Field: "label", QueryTerm: "developers", MatchedText: "developers", Start: 9, End: 19},
							},
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
				Matches: []model.FilterSearchHighlight{
					{Field: "title", QueryTerm: "age", MatchedText: "Age", Start: 0, End: 3},
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

	for name, shorthand := range map[string]string{"workspace": "w", "limit": "l", "offset": "o", "all": "a", "fields": "f", "json": "j", "csv": "c", "table": "t"} {
		flag := cmd.Flags().Lookup(name)
		require.NotNil(t, flag, name)
		assert.Equal(t, shorthand, flag.Shorthand)
	}

	// --no-pager is a persistent root flag, not a per-command one.
	assert.Nil(t, cmd.Flags().Lookup("no-pager"))
}

func TestSearchFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilters("software developers", "ws-1", client.FilterSearchPageSize, client.DefaultRecordOffset).
		Return(searchResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, atTerminal(t, w))
	cmd.SetArgs([]string{"software", "developers", "-w", "ws-1"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	output := stripansi.Strip(b.String())

	assert.Contains(t, output, "Filters matching \"software developers\"\n2 matching filters\n\n")
	assert.True(t, strings.HasSuffix(output, "\nShowing 2 records of 2\n"))
	assert.Contains(t, output, "1. Job title\n   select · ChoiceID · Employment\n")
	assert.Contains(t, output, "   Filter ID    job-title\n")
	assert.Contains(t, output, "   Question     What is your job title?\n")
	assert.Contains(t, output, "   Choices      5 total, 4 matching\n")
	assert.Contains(t, output, "Choice ID    Label\n")
	assert.Contains(t, output, "100          Software developers  (+3 nested)\n")
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

	decoded := decodeSearchEnvelope(t, b.Bytes())
	require.Len(t, decoded, 2)
	assert.Equal(t, "job-title", decoded[0].FilterID)
	assert.Equal(t, 4, decoded[0].Choices.Matched)
	assert.Equal(t, "age", decoded[1].FilterID)
}

func TestSearchFiltersNoResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := &client.SearchFiltersResponse{Results: []model.FilterSearchResult{}, JSONAPIMeta: &client.JSONAPIMeta{}}
	c.EXPECT().
		SearchFilters("zzz", "", client.FilterSearchPageSize, client.DefaultRecordOffset).
		Return(response, nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, atTerminal(t, w))
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
		SearchFilters("zzz", "", client.FilterSearchPageSize, client.DefaultRecordOffset).
		Return(response, nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"zzz", "-j"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	assert.Contains(t, b.String(), `"results": []`)
}

func TestSearchFiltersAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilters("developers", "", client.FilterSearchPageSize, client.DefaultRecordOffset).
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
		{name: "negative limit", args: []string{"dev", "--limit", "-1"}, want: "limit must be greater than or equal to 0"},
		{name: "negative offset", args: []string{"dev", "--offset", "-1"}, want: "offset must be greater than or equal to 0"},
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
		c.EXPECT().SearchFilters("dev", "", 200, 0).Return(pageOf(0, 200, 300), nil),
		c.EXPECT().SearchFilters("dev", "", 50, 200).Return(pageOf(200, 50, 300), nil),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, atTerminal(t, w))
	cmd.SetArgs([]string{"dev", "--limit", "250"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	output := stripansi.Strip(b.String())
	assert.Contains(t, output, "300 matching filters. Use --limit or --all to see more\n")
	assert.True(t, strings.HasSuffix(output, "\nShowing 250 records of 300\n"))
	assert.Contains(t, output, "1. Filter 0\n")
	assert.Contains(t, output, "250. Filter 249\n")
	assert.NotContains(t, output, "filter-250")
}

func TestSearchFiltersAllFetchesUntilExhausted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	gomock.InOrder(
		c.EXPECT().SearchFilters("dev", "ws-1", 200, 0).Return(pageOf(0, 200, 230), nil),
		c.EXPECT().SearchFilters("dev", "ws-1", 200, 200).Return(pageOf(200, 30, 230), nil),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"dev", "--all", "-w", "ws-1", "--json"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)

	decoded := decodeSearchEnvelope(t, b.Bytes())
	require.Len(t, decoded, 230)
	assert.Equal(t, "filter-0", decoded[0].FilterID)
	assert.Equal(t, "filter-229", decoded[229].FilterID)
}

func TestSearchFiltersErrorOnLaterPage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	gomock.InOrder(
		c.EXPECT().SearchFilters("dev", "", 200, 0).Return(pageOf(0, 200, 300), nil),
		c.EXPECT().SearchFilters("dev", "", 200, 200).Return(nil, assert.AnError),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, atTerminal(t, w))
	cmd.SetArgs([]string{"dev", "--all"})
	err := cmd.Execute()
	w.Flush()

	// Results stream as pages arrive, so the first page is already written
	// when the second fails; the error must still be reported.
	require.Error(t, err)
	output := stripansi.Strip(b.String())
	assert.Contains(t, output, "200. Filter 199\n")
	assert.NotContains(t, output, "Filter 200")
}

func TestSearchFiltersTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilters("developers", "", client.FilterSearchPageSize, client.DefaultRecordOffset).
		Return(searchResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"developers", "--table"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	output := b.String()
	assert.Regexp(t, `FilterID\s+Title\s+Type`, output)
	assert.Regexp(t, `job-title\s+Job title\s+select`, output)
	assert.Regexp(t, `age\s+Age\s+range`, output)
	assert.NotContains(t, output, "Rank")
	assert.NotContains(t, output, "MatchedOn")
	assert.Contains(t, output, "Showing 2 records of 2")
}

func TestSearchFiltersCSVWithFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilters("developers", "", client.FilterSearchPageSize, client.DefaultRecordOffset).
		Return(searchResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"developers", "--csv", "--fields", "FilterID,Title"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	assert.Equal(t, "FilterID,Title\njob-title,Job title\nage,Age\n", b.String())
}

func TestSearchFiltersHeaderUsesFirstPageTotal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// The header must be written from the first page, before later pages
	// arrive, and reflect how many will be shown out of the total.
	gomock.InOrder(
		c.EXPECT().SearchFilters("dev", "", 200, 0).Return(pageOf(0, 200, 1000), nil),
		c.EXPECT().SearchFilters("dev", "", 20, 200).Return(pageOf(200, 20, 1000), nil),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, atTerminal(t, w))
	cmd.SetArgs([]string{"dev", "--limit", "220"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	output := stripansi.Strip(b.String())
	assert.True(t, strings.HasPrefix(output, "Filters matching \"dev\"\n1000 matching filters. Use --limit or --all to see more\n\n1. Filter 0\n"))
	assert.Contains(t, output, "220. Filter 219\n")
	assert.True(t, strings.HasSuffix(output, "\nShowing 220 records of 1000\n"))
}

func TestSearchFiltersFooterCountsWhatWasRendered(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// The API claims 50 matches but hands back only 7 and then a short page.
	c.EXPECT().SearchFilters("dev", "", client.FilterSearchPageSize, 0).Return(pageOf(0, 7, 50), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, atTerminal(t, w))
	cmd.SetArgs([]string{"dev"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)
	output := stripansi.Strip(b.String())
	// 50 matches sit inside the default limit of 200, so nothing was cut off
	// and the header must not offer --limit or --all.
	assert.Contains(t, output, "50 matching filters\n")
	assert.NotContains(t, output, "to see more")
	assert.True(t, strings.HasSuffix(output, "\nShowing 7 records of 50\n"))
}

func TestSearchFiltersLimitZeroFetchesAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	gomock.InOrder(
		c.EXPECT().SearchFilters("dev", "", 200, 0).Return(pageOf(0, 130, 130), nil),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"dev", "--limit", "0", "--json"})
	err := cmd.Execute()
	w.Flush()

	require.NoError(t, err)

	decoded := decodeSearchEnvelope(t, b.Bytes())
	assert.Len(t, decoded, 130)
}

// --offset skips results the caller has already seen, and paging continues
// from there rather than restarting at the top of the result set.
func TestSearchFiltersOffsetSkipsEarlierMatches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().SearchFilters("dev", "", 120, 50).Return(pageOf(50, 120, 300), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, atTerminal(t, w))
	cmd.SetArgs([]string{"dev", "--offset", "50", "--limit", "120"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)
	output := stripansi.Strip(b.String())

	// Ranks are reported against the whole result set, not this slice of it.
	assert.Contains(t, output, "51. Filter 50\n")
	assert.Contains(t, output, "170. Filter 169\n")
	assert.NotContains(t, output, "1. Filter 0\n")
}

func TestSearchFiltersOffsetReachesTheAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().SearchFilters("dev", "", client.FilterSearchPageSize, 50).Return(pageOf(50, 2, 300), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"dev", "-o", "50", "--csv", "-f", "FilterID"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)
	assert.Equal(t, "FilterID\nfilter-50\nfilter-51\n", b.String())
}

// Piped into another program with no format flag, results have to arrive as a
// table rather than as the reading view meant for a terminal.
func TestSearchFiltersRendersATableWhenNotATerminal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilters("developers", "", client.FilterSearchPageSize, client.DefaultRecordOffset).
		Return(searchResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, w)
	cmd.SetArgs([]string{"developers"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)
	output := b.String()
	assert.Regexp(t, `FilterID\s+Title\s+Type`, output)
	assert.Regexp(t, `job-title\s+Job title`, output)
	assert.NotContains(t, output, "Rank")
	assert.NotContains(t, output, "Filters matching")
}

// A filter whose choices the preview cannot cover has to name the command
// that lists the rest, or the choice IDs behind it are unreachable.
func TestSearchFiltersNamesTheChoicesCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		SearchFilters("software developers", "", client.FilterSearchPageSize, client.DefaultRecordOffset).
		Return(searchResponse(), nil)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewSearchCommand(c, atTerminal(t, w))
	cmd.SetArgs([]string{"software", "developers"})
	err := cmd.Execute()
	require.NoError(t, w.Flush())

	require.NoError(t, err)
	assert.Contains(t, stripansi.Strip(b.String()), `See them all: prolific filters choices search job-title "software developers"`)
}

// The envelope's window is the one thing no other test pins: --all is the
// only path where limit must be 0, and an offset search must report where
// its window started.
func TestSearchFiltersEnvelopeReportsTheRequestedWindow(t *testing.T) {
	decode := func(t *testing.T, raw []byte) (count, limit, offset int) {
		t.Helper()
		var envelope struct {
			Count  int `json:"count"`
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
		}
		require.NoError(t, json.Unmarshal(raw, &envelope))
		return envelope.Count, envelope.Limit, envelope.Offset
	}

	t.Run("--all is an unbounded window", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := mock_client.NewMockAPI(ctrl)

		gomock.InOrder(
			c.EXPECT().SearchFilters("dev", "", client.FilterSearchPageSize, 0).Return(pageOf(0, 130, 130), nil),
		)

		var b bytes.Buffer
		w := bufio.NewWriter(&b)
		cmd := filters.NewSearchCommand(c, w)
		cmd.SetArgs([]string{"dev", "--all", "--json"})
		require.NoError(t, cmd.Execute())
		require.NoError(t, w.Flush())

		count, limit, offset := decode(t, b.Bytes())
		assert.Equal(t, 130, count)
		assert.Equal(t, 0, limit, "--all asks for no limit, so the envelope must not claim one")
		assert.Equal(t, 0, offset)
	})

	t.Run("a limited, offset window is reported as asked for", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		c := mock_client.NewMockAPI(ctrl)

		c.EXPECT().SearchFilters("dev", "", 10, 50).Return(pageOf(50, 10, 300), nil)

		var b bytes.Buffer
		w := bufio.NewWriter(&b)
		cmd := filters.NewSearchCommand(c, w)
		cmd.SetArgs([]string{"dev", "--limit", "10", "--offset", "50", "--json"})
		require.NoError(t, cmd.Execute())
		require.NoError(t, w.Flush())

		count, limit, offset := decode(t, b.Bytes())
		assert.Equal(t, 300, count)
		assert.Equal(t, 10, limit)
		assert.Equal(t, 50, offset)
	})
}
