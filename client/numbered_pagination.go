package client

// NumberedPageFetcher fetches one page of results from an endpoint that
// paginates by page number rather than limit and offset. Page numbers start
// at 1.
type NumberedPageFetcher[T any] func(page int) (Page[T], error)

// EachNumberedPage walks a page-numbered endpoint in order and passes each
// page to yield as it arrives, so callers can stream output rather than
// waiting for the whole collection.
//
// want is the maximum number of items to deliver; zero or negative means all.
// Results are trimmed so yield never receives more than want items in total.
// Fetching stops when want is satisfied, a page comes back empty or short, the
// API's reported count has been reached, or yield returns an error.
//
// Each yielded page carries the largest Total seen so far, so a later page
// missing meta.count cannot lose a count established by an earlier one.
//
// Callers that paginate by limit and offset want EachPage instead. This exists
// so commands never have to know which of the two models an endpoint uses.
func EachNumberedPage[T any](want int, fetch NumberedPageFetcher[T], yield func(Page[T]) error) error {
	total := 0
	delivered := 0
	pageSize := 0

	for page := 1; ; page++ {
		if page > maxPages {
			return pageLimitError()
		}

		result, err := fetch(page)
		if err != nil {
			return err
		}

		fetched := len(result.Results)
		if page == 1 {
			pageSize = fetched
		}
		if want > 0 && delivered+fetched > want {
			result.Results = result.Results[:want-delivered]
		}
		if result.Total > total {
			total = result.Total
		}
		result.Total = total

		if err := yield(result); err != nil {
			return err
		}

		delivered += len(result.Results)

		if fetched == 0 {
			return nil
		}
		if want > 0 && delivered >= want {
			return nil
		}
		if total > 0 {
			// The count is authoritative: a page can come back short because
			// the server filtered an item out, so stopping on a short page
			// here would silently truncate the walk. An over-reported count
			// costs one extra request, which returns empty and stops above.
			if delivered >= total {
				return nil
			}
			continue
		}
		if fetched < pageSize {
			return nil
		}
	}
}

// FetchNumberedPages collects results across pages using EachNumberedPage, for
// callers that need the whole collection before rendering. The returned total
// is the API's reported count, or the number of items collected if the API did
// not report one.
func FetchNumberedPages[T any](want int, fetch NumberedPageFetcher[T]) ([]T, int, error) {
	var items []T
	total := 0

	err := EachNumberedPage(want, fetch, func(page Page[T]) error {
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
