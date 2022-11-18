package model_test

import (
	"testing"

	"github.com/benmatselby/prolificli/model"
)

func TestGetCurrencyCode(t *testing.T) {
	tt := []struct {
		Name                    string
		PresentmentCurrencyCode string
		CurrencyCode            string
		Expected                string
	}{
		{"Using presentment", "USD", "GBP", "USD"},
		{"Using currency", "", "EUR", "EUR"},
		{"Default", "", "", "GBP"},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			study := model.Study{
				PresentmentCurrencyCode: tc.PresentmentCurrencyCode,
				CurrencyCode:            tc.CurrencyCode,
			}

			actual := study.GetCurrencyCode()

			if actual != tc.Expected {
				t.Fatalf("expected %v, got %v", tc.Expected, actual)
			}
		})
	}
}
