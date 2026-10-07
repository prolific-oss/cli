package client_test

import (
	"testing"

	"github.com/prolific-oss/cli/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A page can come back shorter than the first one without being the last page,
// for example when the server filters an item out of it. The reported count is
// what says whether more remain.
func TestFetchNumberedPagesDoesNotStopOnAShortMiddlePage(t *testing.T) {
	pages := map[int][]int{1: seq(100), 2: seq(99), 3: seq(51)}
	var requested []int

	items, total, err := client.FetchNumberedPages(0, func(page int) (client.Page[int], error) {
		requested = append(requested, page)
		return client.Page[int]{Results: pages[page], Total: 250}, nil
	})

	require.NoError(t, err)
	assert.Len(t, items, 250)
	assert.Equal(t, 250, total)
	assert.Equal(t, []int{1, 2, 3}, requested, "a short middle page must not end the walk")
}

// When the count over-reports, the walk costs one extra request and stops on
// the empty page rather than looping.
func TestFetchNumberedPagesStopsOnAnEmptyPageWhenTheCountOverReports(t *testing.T) {
	var requested []int

	items, total, err := client.FetchNumberedPages(0, func(page int) (client.Page[int], error) {
		requested = append(requested, page)
		if page == 1 {
			return client.Page[int]{Results: seq(10), Total: 500}, nil
		}
		return client.Page[int]{Total: 500}, nil
	})

	require.NoError(t, err)
	assert.Len(t, items, 10)
	assert.Equal(t, 500, total)
	assert.Equal(t, []int{1, 2}, requested)
}
