package client

import "fmt"

// FilterSearchPageSize is the maximum page size accepted by the filter search
// endpoint. Larger requests are satisfied by fetching several pages.
const FilterSearchPageSize = 100

// maxPages bounds how many pages a single fetch-all may request, so an API
// that keeps returning full pages without a count cannot loop forever.
const maxPages = 1000

// Page is a single page of results returned by a paginated fetch.
type Page[T any] struct {
	// Results is the items on this page.
	Results []T
	// Total is the total number of matching items across all pages, as
	// reported by the API's meta.count. Zero if the API did not report it.
	Total int
}

// PageFetcher fetches one page of results using the given limit and offset.
type PageFetcher[T any] func(limit, offset int) (Page[T], error)

// EachPage fetches pages in order and passes each to yield as it arrives, so
// callers can stream output rather than waiting for the whole collection.
//
// want is the maximum number of items to deliver; zero or negative means all.
// pageSize is the maximum page size the API accepts. Results are trimmed so
// yield never receives more than want items in total, even if the API
// ignores limit. Fetching stops when want is satisfied, the API reports no
// more items, a page comes back short, or yield returns an error.
//
// Each yielded page carries the largest Total seen so far, so a later page
// missing meta.count cannot lose a count established by an earlier one.
func EachPage[T any](want, pageSize int, fetch PageFetcher[T], yield func(Page[T]) error) error {
	if pageSize < 1 {
		pageSize = 1
	}

	total := 0
	offset := 0
	delivered := 0

	for pages := 0; ; pages++ {
		if pages >= maxPages {
			return fmt.Errorf("stopped after fetching %d pages without reaching the end of the collection", maxPages)
		}

		limit := pageSize
		if want > 0 && want-delivered < limit {
			limit = want - delivered
		}

		page, err := fetch(limit, offset)
		if err != nil {
			return err
		}

		fetched := len(page.Results)
		if want > 0 && delivered+fetched > want {
			page.Results = page.Results[:want-delivered]
		}
		if page.Total > total {
			total = page.Total
		}
		page.Total = total

		if err := yield(page); err != nil {
			return err
		}

		delivered += len(page.Results)
		offset += fetched

		if fetched == 0 || fetched < limit {
			return nil
		}
		if want > 0 && delivered >= want {
			return nil
		}
		if total > 0 && offset >= total {
			return nil
		}
	}
}

// FetchPages collects results across multiple pages using EachPage, for
// callers that need the whole collection before rendering. The returned total
// is the API's reported count, or the number of items collected if the API
// did not report one.
func FetchPages[T any](want, pageSize int, fetch PageFetcher[T]) ([]T, int, error) {
	var items []T
	total := 0

	err := EachPage(want, pageSize, fetch, func(page Page[T]) error {
		items = append(items, page.Results...)
		total = page.Total
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	if total < len(items) {
		total = len(items)
	}
	return items, total, nil
}
