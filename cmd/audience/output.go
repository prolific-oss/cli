package audience

import (
	"sort"

	"github.com/prolific-oss/cli/client"
)

// CountFields is the default column set for a count, which is a single value
// and so a single row.
const CountFields = "Count"

// BreakdownFields is the default column set for a breakdown.
const BreakdownFields = "Value,Count"

// CountItem is a count as a row, so the shared table and CSV renderers can
// emit it like any other record.
type CountItem struct {
	Count int
}

// BreakdownItem is one bucket of a breakdown: the value participants gave,
// and how many of them gave it.
type BreakdownItem struct {
	Value string
	Count int
}

// NewBreakdownItems turns the API's value-keyed object into rows, sorted
// alphabetically with the N/A bucket last. The API returns an object, whose
// key order is meaningless, so the ordering has to be imposed here for the
// output to be stable between runs.
func NewBreakdownItems(breakdown map[string]int) []BreakdownItem {
	keys := make([]string, 0, len(breakdown))
	for key := range breakdown {
		if key != client.FilterBreakdownNAKey {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if _, ok := breakdown[client.FilterBreakdownNAKey]; ok {
		keys = append(keys, client.FilterBreakdownNAKey)
	}

	items := make([]BreakdownItem, 0, len(keys))
	for _, key := range keys {
		items = append(items, BreakdownItem{Value: key, Count: breakdown[key]})
	}
	return items
}
