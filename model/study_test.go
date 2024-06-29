package model_test

import (
	"testing"

	"github.com/benmatselby/prolificli/model"
)

func TestFilterValueReturnsName(t *testing.T) {
	name := "Patterns of migratory birds"
	study := model.Study{
		ID:   "study-id",
		Name: name,
	}

	if study.FilterValue() != name {
		t.Fatalf("expected filter value to be %s, got %s", name, study.FilterValue())
	}
}

func TestTitleIsTheStudyName(t *testing.T) {
	name := "Patterns of migratory birds"
	study := model.Study{
		ID:   "study-id",
		Name: name,
	}

	if study.Title() != name {
		t.Fatalf("expected filter value to be %s, got %s", name, study.Title())
	}
}

func TestDescriptionReturnsADescriptiveString(t *testing.T) {
	study := model.Study{
		ID:                   "study-id",
		Name:                 "Patterns of migratory birds",
		Status:               model.StatusActive,
		StudyType:            "single",
		TotalAvailablePlaces: 515,
	}

	expected := "active - single - 515 places available"
	if study.Description() != expected {
		t.Fatalf("expected filter value to be %s, got %s", expected, study.Description())
	}
}

func TestGetCurrencyCodeCanFigureOutWhichCurrencyToUse(t *testing.T) {
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
