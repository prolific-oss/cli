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
