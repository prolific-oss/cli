package client_test

import (
	"errors"
	"testing"

	"github.com/prolific-oss/cli/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEachNumberedPageWalksPagesUntilExhausted(t *testing.T) {
	var pages []int
	fetch := func(page int) (client.Page[string], error) {
		pages = append(pages, page)
		if page == 1 {
			return client.Page[string]{Results: []string{"a", "b"}, Total: 3}, nil
		}
		return client.Page[string]{Results: []string{"c"}, Total: 3}, nil
	}

	items, total, err := client.FetchNumberedPages(0, fetch)

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, pages)
	assert.Equal(t, []string{"a", "b", "c"}, items)
	assert.Equal(t, 3, total)
}

func TestEachNumberedPageStopsOnEmptyPageWhenTotalUnknown(t *testing.T) {
	var pages []int
	fetch := func(page int) (client.Page[string], error) {
		pages = append(pages, page)
		if page == 1 {
			return client.Page[string]{Results: []string{"a"}}, nil
		}
		return client.Page[string]{}, nil
	}

	items, total, err := client.FetchNumberedPages(0, fetch)

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, pages)
	assert.Equal(t, []string{"a"}, items)
	assert.Equal(t, 1, total)
}

func TestEachNumberedPageTrimsToWant(t *testing.T) {
	var pages []int
	fetch := func(page int) (client.Page[string], error) {
		pages = append(pages, page)
		return client.Page[string]{Results: []string{"a", "b", "c"}, Total: 9}, nil
	}

	items, total, err := client.FetchNumberedPages(2, fetch)

	require.NoError(t, err)
	assert.Equal(t, []int{1}, pages, "should not fetch a page it does not need")
	assert.Equal(t, []string{"a", "b"}, items)
	assert.Equal(t, 9, total)
}

func TestEachNumberedPagePropagatesFetchError(t *testing.T) {
	boom := errors.New("boom")
	_, _, err := client.FetchNumberedPages(0, func(int) (client.Page[string], error) {
		return client.Page[string]{}, boom
	})
	assert.ErrorIs(t, err, boom)
}

func TestEachNumberedPageStopsWhenYieldErrors(t *testing.T) {
	boom := errors.New("stop")
	pages := 0
	err := client.EachNumberedPage(0, func(int) (client.Page[string], error) {
		pages++
		return client.Page[string]{Results: []string{"a"}, Total: 100}, nil
	}, func(client.Page[string]) error {
		return boom
	})
	assert.ErrorIs(t, err, boom)
	assert.Equal(t, 1, pages)
}

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

func TestFetchPagesSinglePageWhenWantFitsInPageSize(t *testing.T) {
	fetch, calls := fakeFetcher(seq(500), true)

	items, total, err := client.FetchPages(25, 100, fetch)

	require.NoError(t, err)
	assert.Equal(t, seq(25), items)
	assert.Equal(t, 500, total)
	assert.Equal(t, []call{{25, 0}}, *calls)
}

func TestFetchPagesSpansMultiplePagesAndTrimsLastRequest(t *testing.T) {
	fetch, calls := fakeFetcher(seq(500), true)

	items, total, err := client.FetchPages(250, 100, fetch)

	require.NoError(t, err)
	assert.Equal(t, seq(250), items)
	assert.Equal(t, 500, total)
	assert.Equal(t, []call{{100, 0}, {100, 100}, {50, 200}}, *calls)
}

func TestFetchPagesAllStopsAtReportedTotal(t *testing.T) {
	fetch, calls := fakeFetcher(seq(230), true)

	items, total, err := client.FetchPages(0, 100, fetch)

	require.NoError(t, err)
	assert.Equal(t, seq(230), items)
	assert.Equal(t, 230, total)
	assert.Equal(t, []call{{100, 0}, {100, 100}, {100, 200}}, *calls)
}

func TestFetchPagesAllStopsOnShortPageWithoutTotal(t *testing.T) {
	fetch, calls := fakeFetcher(seq(230), false)

	items, total, err := client.FetchPages(0, 100, fetch)

	require.NoError(t, err)
	assert.Equal(t, seq(230), items)
	assert.Equal(t, 230, total)
	assert.Equal(t, []call{{100, 0}, {100, 100}, {100, 200}}, *calls)
}

func TestFetchPagesExactMultipleOfPageSizeDoesNotOverfetchWhenTotalKnown(t *testing.T) {
	fetch, calls := fakeFetcher(seq(200), true)

	items, _, err := client.FetchPages(0, 100, fetch)

	require.NoError(t, err)
	assert.Len(t, items, 200)
	assert.Equal(t, []call{{100, 0}, {100, 100}}, *calls)
}

func TestFetchPagesEmptyCollection(t *testing.T) {
	fetch, calls := fakeFetcher(nil, true)

	items, total, err := client.FetchPages(25, 100, fetch)

	require.NoError(t, err)
	assert.Empty(t, items)
	assert.Equal(t, 0, total)
	assert.Equal(t, []call{{25, 0}}, *calls)
}

func TestFetchPagesWantLargerThanCollection(t *testing.T) {
	fetch, _ := fakeFetcher(seq(7), true)

	items, total, err := client.FetchPages(50, 100, fetch)

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

	items, total, err := client.FetchPages(0, 100, fetch)

	assert.ErrorIs(t, err, boom)
	assert.Nil(t, items)
	assert.Equal(t, 0, total)
}

func TestFetchPagesTrimsWhenAPIIgnoresLimit(t *testing.T) {
	// Always return 100 items regardless of the requested limit.
	fetch := func(limit, offset int) (client.Page[int], error) {
		return client.Page[int]{Results: seq(100), Total: 500}, nil
	}

	items, total, err := client.FetchPages(30, 100, fetch)

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

	items, total, err := client.FetchPages(0, 100, fetch)

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

	_, _, err := client.FetchPages(0, 100, fetch)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "stopped after fetching 1000 pages")
	assert.Equal(t, 1000, calls)
}

func TestEachPageYieldsPagesAsTheyArrive(t *testing.T) {
	fetch, calls := fakeFetcher(seq(250), true)

	var yielded [][]int
	var totals []int
	err := client.EachPage(0, 100, fetch, func(page client.Page[int]) error {
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

	err := client.EachPage(0, 100, fetch, func(client.Page[int]) error { return boom })

	assert.ErrorIs(t, err, boom)
	assert.Len(t, *calls, 1)
}

func TestEachPageYieldsEmptyFirstPage(t *testing.T) {
	fetch, _ := fakeFetcher(nil, true)

	yields := 0
	err := client.EachPage(25, 100, fetch, func(page client.Page[int]) error {
		yields++
		assert.Empty(t, page.Results)
		assert.Equal(t, 0, page.Total)
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 1, yields)
}
