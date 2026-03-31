package model_test

import (
	"testing"

	"github.com/prolific-oss/cli/model"
)

func TestSubmissionCountsToItems(t *testing.T) {
	counts := &model.SubmissionCounts{
		Active:            5,
		Approved:          10,
		AwaitingReview:    0,
		Rejected:          2,
		Reserved:          0,
		Returned:          0,
		TimedOut:          0,
		PartiallyApproved: 0,
		ScreenedOut:       0,
		Total:             17,
	}

	items := counts.ToItems()

	if len(items) != 3 {
		t.Fatalf("expected 3 items (non-zero, excluding Total), got %d", len(items))
	}

	expected := []struct {
		label string
		key   string
		count int
	}{
		{"Active", "ACTIVE", 5},
		{"Approved", "APPROVED", 10},
		{"Rejected", "REJECTED", 2},
	}

	for i, e := range expected {
		if items[i].StatusLabel != e.label {
			t.Fatalf("item %d: expected label %s, got %s", i, e.label, items[i].StatusLabel)
		}
		if items[i].StatusKey != e.key {
			t.Fatalf("item %d: expected key %s, got %s", i, e.key, items[i].StatusKey)
		}
		if items[i].Count != e.count {
			t.Fatalf("item %d: expected count %d, got %d", i, e.count, items[i].Count)
		}
	}
}

func TestSubmissionCountsToItemsAllZero(t *testing.T) {
	counts := &model.SubmissionCounts{}
	items := counts.ToItems()

	if len(items) != 0 {
		t.Fatalf("expected 0 items for all-zero counts, got %d", len(items))
	}
}

func TestSubmissionCountsToItemsExcludesTotal(t *testing.T) {
	counts := &model.SubmissionCounts{
		Total: 100,
	}
	items := counts.ToItems()

	if len(items) != 0 {
		t.Fatalf("expected 0 items (Total should be excluded), got %d", len(items))
	}
}

func TestSubmissionCountItemListInterface(t *testing.T) {
	item := model.SubmissionCountItem{
		StatusLabel: "Approved",
		StatusKey:   "APPROVED",
		Count:       10,
	}

	if item.FilterValue() != "Approved" {
		t.Fatalf("expected FilterValue 'Approved', got '%s'", item.FilterValue())
	}
	if item.Title() != "Approved" {
		t.Fatalf("expected Title 'Approved', got '%s'", item.Title())
	}
	if item.Description() != "10 submissions" {
		t.Fatalf("expected Description '10 submissions', got '%s'", item.Description())
	}
}

func TestSubmissionCountsToItemsPreservesOrder(t *testing.T) {
	counts := &model.SubmissionCounts{
		Active:            1,
		Approved:          1,
		AwaitingReview:    1,
		Rejected:          1,
		Reserved:          1,
		Returned:          1,
		TimedOut:          1,
		PartiallyApproved: 1,
		ScreenedOut:       1,
		Total:             9,
	}

	items := counts.ToItems()

	if len(items) != 9 {
		t.Fatalf("expected 9 items, got %d", len(items))
	}

	expectedOrder := []string{
		"Active", "Approved", "Awaiting Review", "Rejected", "Reserved",
		"Returned", "Timed Out", "Partially Approved", "Screened Out",
	}

	for i, label := range expectedOrder {
		if items[i].StatusLabel != label {
			t.Fatalf("item %d: expected %s, got %s", i, label, items[i].StatusLabel)
		}
	}
}
