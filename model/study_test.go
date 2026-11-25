package model_test

import (
	"encoding/json"
	"testing"

	"github.com/prolific-oss/cli/model"
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

// The API merges data_collection_metadata, so an update that raises the annotators per task must send that key and
// nothing else: a places figure of 0 or an empty credential pool would be read as a change.
func TestUpdateStudySendsOnlyTheMetadataGiven(t *testing.T) {
	body, err := json.Marshal(model.UpdateStudy{DataCollectionMetadata: map[string]any{"annotators_per_task": 3}})
	if err != nil {
		t.Fatal(err)
	}

	expected := `{"data_collection_metadata":{"annotators_per_task":3}}`
	if string(body) != expected {
		t.Fatalf("expected %s; got %s", expected, body)
	}
}

func TestStudyIsATaskBuilderBatchStudy(t *testing.T) {
	for method, expected := range map[string]bool{
		"AI_TASK_BUILDER_BATCH": true,
		"AI_TASK_BUILDER":       true,
		"DC_TOOL":               true,
		"":                      false,
		"HUMAN_SIGNAL":          false,
		// A Task Builder collection is a different thing: its places are not calculated from a batch.
		"AI_TASK_BUILDER_COLLECTION": false,
	} {
		if actual := (model.Study{DataCollectionMethod: method}).IsTaskBuilderBatch(); actual != expected {
			t.Fatalf("expected %q to give %v; got %v", method, expected, actual)
		}
	}
}

func TestStudyIsPublished(t *testing.T) {
	for name, tc := range map[string]struct {
		publishedAt any
		expected    bool
	}{
		"published":       {publishedAt: "2026-01-01T00:00:00.000Z", expected: true},
		"never published": {publishedAt: nil, expected: false},
		"empty":           {publishedAt: "", expected: false},
		// published_at is a date or null, so anything else is not a published date.
		"not a date": {publishedAt: map[string]any{}, expected: false},
		"a number":   {publishedAt: 1.0, expected: false},
	} {
		if actual := (model.Study{PublishedAt: tc.publishedAt}).IsPublished(); actual != tc.expected {
			t.Fatalf("%s: expected %v; got %v", name, tc.expected, actual)
		}
	}
}
