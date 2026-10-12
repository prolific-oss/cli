package client_test

import (
	"errors"
	"testing"

	"github.com/prolific-oss/cli/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type call struct{ limit, offset int }

// fakeFetcher serves a fixed collection of ints, recording each page request.
func fakeFetcher(collection []int, reportTotal bool) (client.PageFetcher[int], *[]call) {
	calls := &[]call{}
	return func(limit, offset int) (client.Page[int], error) {
		*calls = append(*calls, call{limit, offset})
		end := offset + limit
		if offset > len(collection) {
			offset = len(collection)
		}
		if end > len(collection) {
			end = len(collection)
		}
		page := client.Page[int]{Results: collection[offset:end]}
		if reportTotal {
			page.Total = len(collection)
		}
		return page, nil
	}, calls
}

func seq(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	return s
}

// The offsets handed to the fetcher are absolute, so a caller fetching
// everything from part way through asks the API for the right rows.
func TestFetchPagesStartOffsetReachesTheFetcher(t *testing.T) {
	fetch, calls := fakeFetcher(seq(400), true)

	items, total, err := client.FetchPages(50, 100, 200, fetch)

	require.NoError(t, err)
	assert.Equal(t, []call{{50, 200}}, *calls)
	assert.Equal(t, seq(400)[200:250], items)
	assert.Equal(t, 400, total)
}

// Termination compares absolute position against the API's own count, so the
// last full page of an offset --all is not followed by a wasted empty request.
func TestFetchPagesAllFromAnOffsetDoesNotOverfetch(t *testing.T) {
	fetch, calls := fakeFetcher(seq(400), true)

	items, total, err := client.FetchPages(0, 200, 200, fetch)

	require.NoError(t, err)
	assert.Equal(t, []call{{200, 200}}, *calls, "the 200 rows from offset 200 cover the collection")
	assert.Len(t, items, 200)
	assert.Equal(t, 400, total)
}

func TestFetchPagesSinglePageWhenWantFitsInPageSize(t *testing.T) {
	fetch, calls := fakeFetcher(seq(500), true)

	items, total, err := client.FetchPages(25, 100, 0, fetch)

	require.NoError(t, err)
	assert.Equal(t, seq(25), items)
	assert.Equal(t, 500, total)
	assert.Equal(t, []call{{25, 0}}, *calls)
}

func TestFetchPagesSpansMultiplePagesAndTrimsLastRequest(t *testing.T) {
	fetch, calls := fakeFetcher(seq(500), true)

	items, total, err := client.FetchPages(250, 100, 0, fetch)

	require.NoError(t, err)
	assert.Equal(t, seq(250), items)
	assert.Equal(t, 500, total)
	assert.Equal(t, []call{{100, 0}, {100, 100}, {50, 200}}, *calls)
}

func TestFetchPagesAllStopsAtReportedTotal(t *testing.T) {
	fetch, calls := fakeFetcher(seq(230), true)

	items, total, err := client.FetchPages(0, 100, 0, fetch)

	require.NoError(t, err)
	assert.Equal(t, seq(230), items)
	assert.Equal(t, 230, total)
	assert.Equal(t, []call{{100, 0}, {100, 100}, {100, 200}}, *calls)
}

func TestFetchPagesAllStopsOnShortPageWithoutTotal(t *testing.T) {
	fetch, calls := fakeFetcher(seq(230), false)

	items, total, err := client.FetchPages(0, 100, 0, fetch)

	require.NoError(t, err)
	assert.Equal(t, seq(230), items)
	assert.Equal(t, 230, total)
	assert.Equal(t, []call{{100, 0}, {100, 100}, {100, 200}}, *calls)
}

func TestFetchPagesExactMultipleOfPageSizeDoesNotOverfetchWhenTotalKnown(t *testing.T) {
	fetch, calls := fakeFetcher(seq(200), true)

	items, _, err := client.FetchPages(0, 100, 0, fetch)

	require.NoError(t, err)
	assert.Len(t, items, 200)
	assert.Equal(t, []call{{100, 0}, {100, 100}}, *calls)
}

func TestFetchPagesEmptyCollection(t *testing.T) {
	fetch, calls := fakeFetcher(nil, true)

	items, total, err := client.FetchPages(25, 100, 0, fetch)

	require.NoError(t, err)
	assert.Empty(t, items)
	assert.Equal(t, 0, total)
	assert.Equal(t, []call{{25, 0}}, *calls)
}

func TestFetchPagesWantLargerThanCollection(t *testing.T) {
	fetch, _ := fakeFetcher(seq(7), true)

	items, total, err := client.FetchPages(50, 100, 0, fetch)

	require.NoError(t, err)
	assert.Equal(t, seq(7), items)
	assert.Equal(t, 7, total)
}

func TestFetchPagesPropagatesError(t *testing.T) {
	boom := errors.New("boom")
	fetch := func(limit, offset int) (client.Page[int], error) {
		if offset == 0 {
			return client.Page[int]{Results: seq(100), Total: 300}, nil
		}
		return client.Page[int]{}, boom
	}

	items, total, err := client.FetchPages(0, 100, 0, fetch)

	assert.ErrorIs(t, err, boom)
	assert.Nil(t, items)
	assert.Equal(t, 0, total)
}

func TestFetchPagesTrimsWhenAPIIgnoresLimit(t *testing.T) {
	// Always return 100 items regardless of the requested limit.
	fetch := func(limit, offset int) (client.Page[int], error) {
		return client.Page[int]{Results: seq(100), Total: 500}, nil
	}

	items, total, err := client.FetchPages(30, 100, 0, fetch)

	require.NoError(t, err)
	assert.Len(t, items, 30)
	assert.Equal(t, 500, total)
}

func TestFetchPagesKeepsTotalWhenLaterPageOmitsCount(t *testing.T) {
	fetch := func(limit, offset int) (client.Page[int], error) {
		if offset == 0 {
			return client.Page[int]{Results: seq(100), Total: 300}, nil
		}
		return client.Page[int]{Results: seq(50)}, nil
	}

	items, total, err := client.FetchPages(0, 100, 0, fetch)

	require.NoError(t, err)
	assert.Len(t, items, 150)
	assert.Equal(t, 300, total)
}

func TestFetchPagesGuardsAgainstEndlessFullPages(t *testing.T) {
	calls := 0
	fetch := func(limit, offset int) (client.Page[int], error) {
		calls++
		return client.Page[int]{Results: seq(limit)}, nil
	}

	_, _, err := client.FetchPages(0, 100, 0, fetch)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "stopped after fetching 1000 pages")
	assert.Equal(t, 1000, calls)
}

func TestEachPageYieldsPagesAsTheyArrive(t *testing.T) {
	fetch, calls := fakeFetcher(seq(250), true)

	var yielded [][]int
	var totals []int
	err := client.EachPage(0, 100, 0, fetch, func(page client.Page[int]) error {
		yielded = append(yielded, page.Results)
		totals = append(totals, page.Total)
		return nil
	})

	require.NoError(t, err)
	require.Len(t, yielded, 3)
	assert.Equal(t, seq(100), yielded[0])
	assert.Equal(t, 100, len(yielded[1]))
	assert.Equal(t, 50, len(yielded[2]))
	assert.Equal(t, []int{250, 250, 250}, totals)
	assert.Equal(t, []call{{100, 0}, {100, 100}, {100, 200}}, *calls)
}

func TestEachPageStopsWhenYieldFails(t *testing.T) {
	fetch, calls := fakeFetcher(seq(500), true)
	boom := errors.New("closed")

	err := client.EachPage(0, 100, 0, fetch, func(client.Page[int]) error { return boom })

	assert.ErrorIs(t, err, boom)
	assert.Len(t, *calls, 1)
}

func TestEachPageYieldsEmptyFirstPage(t *testing.T) {
	fetch, _ := fakeFetcher(nil, true)

	yields := 0
	err := client.EachPage(25, 100, 0, fetch, func(page client.Page[int]) error {
		yields++
		assert.Empty(t, page.Results)
		assert.Equal(t, 0, page.Total)
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 1, yields)
}

func TestReportedTotal(t *testing.T) {
	withCount := func(count int) *client.JSONAPIMeta {
		meta := &client.JSONAPIMeta{}
		meta.Meta.Count = count
		return meta
	}

	tests := map[string]struct {
		meta    *client.JSONAPIMeta
		records int
		want    int
	}{
		"no meta block falls back to the records in hand": {meta: nil, records: 10, want: 10},
		"meta count is the total":                         {meta: withCount(90), records: 20, want: 90},
		// A meta block can arrive without a usable count. Reporting a total
		// below the records already held would be a count no caller can act
		// on, so it clamps up, as FetchPages does.
		"empty meta block clamps up to the records":    {meta: withCount(0), records: 10, want: 10},
		"count smaller than the records clamps up":     {meta: withCount(3), records: 10, want: 10},
		"no records and no count is zero":              {meta: nil, records: 0, want: 0},
		"count with no records is still the API total": {meta: withCount(90), records: 0, want: 90},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, client.ReportedTotal(tt.meta, tt.records))
		})
	}
}
