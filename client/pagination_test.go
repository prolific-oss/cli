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
