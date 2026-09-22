package client

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

// FetchPages collects results across multiple pages so callers can think in
// "how many do I want" rather than offsets.
//
// want is the maximum number of items to return; zero or negative means all.
// pageSize is the maximum page size the API accepts. Fetching stops when want
// is satisfied, the API reports no more items, or a page comes back short.
// The returned total is the API's meta.count from the last page fetched, or
// the number of items collected if the API did not report a count.
func FetchPages[T any](want, pageSize int, fetch PageFetcher[T]) ([]T, int, error) {
	if pageSize < 1 {
		pageSize = 1
	}

	var items []T
	total := 0
	offset := 0

	for {
		limit := pageSize
		if want > 0 && want-len(items) < limit {
			limit = want - len(items)
		}

		page, err := fetch(limit, offset)
		if err != nil {
			return nil, 0, err
		}

		items = append(items, page.Results...)
		total = page.Total
		offset += len(page.Results)

		if len(page.Results) == 0 || len(page.Results) < limit {
			break
		}
		if want > 0 && len(items) >= want {
			break
		}
		if total > 0 && offset >= total {
			break
		}
	}

	if total < len(items) {
		total = len(items)
	}

	return items, total, nil
}
