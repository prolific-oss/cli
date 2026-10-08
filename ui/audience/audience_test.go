package audience_test

import (
	"testing"

	"github.com/prolific-oss/cli/client"
	uiaudience "github.com/prolific-oss/cli/ui/audience"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Choice IDs are numeric strings. Sorting them as text puts 10 between 1 and
// 2, which scrambles the axis of any chart or spreadsheet built from the CSV.
func TestNewBreakdownItemsOrdersNumericBucketsNumerically(t *testing.T) {
	items := uiaudience.NewBreakdownItems(map[string]int{
		"0": 1, "1": 2, "2": 3, "9": 4, "10": 5, "11": 6, "12": 7,
		client.FilterBreakdownNAKey: 8,
	})

	var order []string
	for _, item := range items {
		order = append(order, item.Value)
	}

	assert.Equal(t, []string{"0", "1", "2", "9", "10", "11", "12", client.FilterBreakdownNAKey}, order)
}

// A range filter labels its buckets, so not every key is a number.
func TestNewBreakdownItemsOrdersLabelledBucketsAsText(t *testing.T) {
	items := uiaudience.NewBreakdownItems(map[string]int{
		"25-34": 2, "18-24": 1, client.FilterBreakdownNAKey: 3,
	})

	require.Len(t, items, 3)
	assert.Equal(t, "18-24", items[0].Value)
	assert.Equal(t, "25-34", items[1].Value)
	assert.Equal(t, client.FilterBreakdownNAKey, items[2].Value)
}

// Numbers and labels can arrive together; numbers lead, and the ordering has
// to stay total so the output does not shift between runs.
func TestNewBreakdownItemsOrdersMixedBuckets(t *testing.T) {
	items := uiaudience.NewBreakdownItems(map[string]int{
		"other": 1, "2": 2, "10": 3, "a label": 4,
	})

	var order []string
	for _, item := range items {
		order = append(order, item.Value)
	}

	assert.Equal(t, []string{"2", "10", "a label", "other"}, order)
}

func TestNewBreakdownItemsCarriesCounts(t *testing.T) {
	items := uiaudience.NewBreakdownItems(map[string]int{"0": 4, "1": 3, client.FilterBreakdownNAKey: 5})

	assert.Equal(t, []uiaudience.BreakdownItem{
		{Value: "0", Count: 4},
		{Value: "1", Count: 3},
		{Value: client.FilterBreakdownNAKey, Count: 5},
	}, items)
}

// A breakdown with no N/A bucket must not invent one.
func TestNewBreakdownItemsWithoutNA(t *testing.T) {
	assert.Equal(t, []uiaudience.BreakdownItem{{Value: "1", Count: 2}}, uiaudience.NewBreakdownItems(map[string]int{"1": 2}))
}

func TestNewBreakdownItemsEmpty(t *testing.T) {
	assert.Empty(t, uiaudience.NewBreakdownItems(nil))
	assert.NotNil(t, uiaudience.NewBreakdownItems(nil))
}
