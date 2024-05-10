package ui_test

import (
	"testing"

	"github.com/benmatselby/prolificli/ui"
)

func TestRenderMoney(t *testing.T) {
	tt := []struct {
		name     string
		amount   float64
		currency string
		expected string
	}{{
		name:     "Pound sterling",
		amount:   10.00,
		currency: "GBP",
		expected: "£10.00",
	}, {
		name:     "Default sterling if nothing passed",
		amount:   1.99,
		currency: "",
		expected: "£1.99",
	}, {
		name:     "Dollar",
		amount:   80001.01,
		currency: "USD",
		expected: "$80001.01",
	}}

	for _, tc := range tt {
		actual := ui.RenderMoney(tc.amount, tc.currency)

		if tc.expected != actual {
			t.Fatalf("expected %v got %v", tc.expected, actual)
		}
	}
}

func TestRenderRecordCounter(t *testing.T) {
	tt := []struct {
		name     string
		count    int
		total    int
		expected string
	}{
		{
			name:     "Single record count",
			count:    1,
			total:    1,
			expected: "Showing 1 record of 1",
		},
		{
			name:     "Showing more than one record",
			count:    2,
			total:    10,
			expected: "Showing 2 records of 10",
		},
	}

	for _, tc := range tt {
		actual := ui.RenderRecordCounter(tc.count, tc.total)

		if tc.expected != actual {
			t.Fatalf("expected '%v' got '%v'", tc.expected, actual)
		}
	}
}
