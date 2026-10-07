package client_test

import (
	"errors"
	"testing"

	"github.com/prolific-oss/cli/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeNumberedFetcher serves a fixed collection of ints in pages of pageSize,
// recording the page number of each request. Page numbers start at 1.
func fakeNumberedFetcher(collection []int, pageSize int, reportTotal bool) (client.NumberedPageFetcher[int], *[]int) {
	pages := &[]int{}
	return func(page int) (client.Page[int], error) {
		*pages = append(*pages, page)
		start := (page - 1) * pageSize
		if start > len(collection) {
			start = len(collection)
		}
		end := start + pageSize
		if end > len(collection) {
			end = len(collection)
		}
		result := client.Page[int]{Results: collection[start:end]}
		if reportTotal {
			result.Total = len(collection)
		}
		return result, nil
	}, pages
}

func TestFetchNumberedPagesStopsOnceWantIsSatisfied(t *testing.T) {
	fetch, pages := fakeNumberedFetcher(seq(500), 100, true)

	items, total, err := client.FetchNumberedPages(25, fetch)

	require.NoError(t, err)
	assert.Equal(t, seq(25), items)
	assert.Equal(t, 500, total)
	assert.Equal(t, []int{1}, *pages, "must not request a second page once want is met")
}

func TestFetchNumberedPagesSpansPagesAndTrimsToWant(t *testing.T) {
	fetch, pages := fakeNumberedFetcher(seq(500), 100, true)

	items, total, err := client.FetchNumberedPages(250, fetch)

	require.NoError(t, err)
	assert.Equal(t, seq(250), items)
	assert.Equal(t, 500, total)
	assert.Equal(t, []int{1, 2, 3}, *pages)
}

func TestFetchNumberedPagesAllStopsAtReportedTotal(t *testing.T) {
	fetch, pages := fakeNumberedFetcher(seq(230), 100, true)

	items, total, err := client.FetchNumberedPages(0, fetch)

	require.NoError(t, err)
	assert.Equal(t, seq(230), items)
	assert.Equal(t, 230, total)
	assert.Equal(t, []int{1, 2, 3}, *pages)
}

func TestFetchNumberedPagesAllStopsOnEmptyPageWithoutTotal(t *testing.T) {
	fetch, pages := fakeNumberedFetcher(seq(150), 100, false)

	items, total, err := client.FetchNumberedPages(0, fetch)

	require.NoError(t, err)
	assert.Equal(t, seq(150), items)
	assert.Equal(t, 150, total, "total falls back to the number of items collected")
	assert.Equal(t, []int{1, 2}, *pages, "a short page ends the walk")
}

func TestFetchNumberedPagesEmptyCollection(t *testing.T) {
	fetch, pages := fakeNumberedFetcher(nil, 100, true)

	items, total, err := client.FetchNumberedPages(0, fetch)

	require.NoError(t, err)
	assert.Empty(t, items)
	assert.Equal(t, 0, total)
	assert.Equal(t, []int{1}, *pages)
}

func TestFetchNumberedPagesPropagatesFetchError(t *testing.T) {
	boom := errors.New("boom")
	fetch := func(page int) (client.Page[int], error) { return client.Page[int]{}, boom }

	_, _, err := client.FetchNumberedPages(0, fetch)

	require.ErrorIs(t, err, boom)
}

func TestFetchNumberedPagesStopsAtMaxPages(t *testing.T) {
	// An API that never reports a total and always returns a full page would
	// otherwise loop forever.
	fetch := func(page int) (client.Page[int], error) {
		return client.Page[int]{Results: seq(100)}, nil
	}

	_, _, err := client.FetchNumberedPages(0, fetch)

	require.ErrorContains(t, err, "without reaching the end of the collection")
}

func TestEachNumberedPageYieldsPagesAsTheyArrive(t *testing.T) {
	fetch, _ := fakeNumberedFetcher(seq(250), 100, true)

	var sizes []int
	var totals []int
	err := client.EachNumberedPage(0, fetch, func(page client.Page[int]) error {
		sizes = append(sizes, len(page.Results))
		totals = append(totals, page.Total)
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, []int{100, 100, 50}, sizes)
	assert.Equal(t, []int{250, 250, 250}, totals)
}

func TestEachNumberedPageStopsWhenYieldFails(t *testing.T) {
	boom := errors.New("stop")
	fetch, pages := fakeNumberedFetcher(seq(500), 100, true)

	err := client.EachNumberedPage(0, fetch, func(client.Page[int]) error { return boom })

	require.ErrorIs(t, err, boom)
	assert.Equal(t, []int{1}, *pages)
}
