// Package audience renders participant counts and breakdowns.
package audience

import (
	"sort"
	"strconv"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/ui"
)

// CountFields is the default column set for a count, which is a single value
// and so a single row whichever format asked for it.
var CountFields = ui.FieldSet{CSV: "Count", Table: "Count"}

// BreakdownFields is the default column set for a breakdown. A breakdown is
// two columns, so a table has no reason to show fewer than the CSV.
var BreakdownFields = ui.FieldSet{CSV: "Value,Count", Table: "Value,Count"}

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

// NewBreakdownItems turns the API's value-keyed object into rows. The API
// returns an object, whose key order is meaningless, so the ordering is
// imposed here: buckets in order, with the N/A bucket last.
func NewBreakdownItems(breakdown map[string]int) []BreakdownItem {
	keys := make([]string, 0, len(breakdown))
	for key := range breakdown {
		if key != client.FilterBreakdownNAKey {
			keys = append(keys, key)
		}
	}
	sort.Slice(keys, func(i, j int) bool { return lessBucket(keys[i], keys[j]) })

	if _, ok := breakdown[client.FilterBreakdownNAKey]; ok {
		keys = append(keys, client.FilterBreakdownNAKey)
	}

	items := make([]BreakdownItem, 0, len(keys))
	for _, key := range keys {
		items = append(items, BreakdownItem{Value: key, Count: breakdown[key]})
	}
	return items
}

// lessBucket orders two bucket keys. Choice IDs are numeric strings, so
// sorting them as text would put 10 between 1 and 2 and scramble the axis of
// any chart or spreadsheet built from the output. Numbers therefore sort
// numerically, and anything else — a range filter's bucket label, say —
// falls back to text, after the numbers.
func lessBucket(a, b string) bool {
	numA, errA := strconv.Atoi(a)
	numB, errB := strconv.Atoi(b)

	switch {
	case errA == nil && errB == nil:
		return numA < numB
	case errA == nil:
		return true
	case errB == nil:
		return false
	default:
		return a < b
	}
}
