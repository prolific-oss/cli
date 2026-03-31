package submission_test

import (
	"strings"
	"testing"
	"time"

	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui/submission"
)

func TestFilterSubmissionsByStatus(t *testing.T) {
	subs := []model.Submission{
		{ID: "1", ParticipantID: "p1", Status: "APPROVED"},
		{ID: "2", ParticipantID: "p2", Status: "REJECTED"},
		{ID: "3", ParticipantID: "p3", Status: "APPROVED"},
		{ID: "4", ParticipantID: "p4", Status: "ACTIVE"},
	}

	approved := submission.FilterSubmissionsByStatus(subs, "APPROVED")
	if len(approved) != 2 {
		t.Fatalf("expected 2 approved submissions, got %d", len(approved))
	}
	if approved[0].ID != "1" || approved[1].ID != "3" {
		t.Fatalf("expected IDs 1 and 3, got %s and %s", approved[0].ID, approved[1].ID)
	}

	rejected := submission.FilterSubmissionsByStatus(subs, "REJECTED")
	if len(rejected) != 1 {
		t.Fatalf("expected 1 rejected submission, got %d", len(rejected))
	}

	none := submission.FilterSubmissionsByStatus(subs, "RETURNED")
	if len(none) != 0 {
		t.Fatalf("expected 0 returned submissions, got %d", len(none))
	}
}

func TestFilterSubmissionsByStatusCaseInsensitive(t *testing.T) {
	subs := []model.Submission{
		{ID: "1", ParticipantID: "p1", Status: "APPROVED"},
	}

	result := submission.FilterSubmissionsByStatus(subs, "approved")
	if len(result) != 1 {
		t.Fatalf("expected case-insensitive match, got %d results", len(result))
	}
}

func TestRenderSubmission(t *testing.T) {
	sub := model.Submission{
		ID:            "sub-123",
		ParticipantID: "part-456",
		Status:        "APPROVED",
		StudyCode:     "STUDY-789",
		StartedAt:     time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
		CompletedAt:   time.Date(2026, 1, 15, 10, 45, 0, 0, time.UTC),
		TimeTaken:     900,
		Reward:        500,
	}

	output := submission.RenderSubmission(sub)

	expectedFields := []string{
		"Participant: part-456",
		"ID:          sub-123",
		"Status:      APPROVED",
		"Study Code:  STUDY-789",
		"Started At:  2026-01-15 10:30:00",
		"Completed:   2026-01-15 10:45:00",
		"Time Taken:  900s",
		"Reward:      500",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Fatalf("expected output to contain '%s', got:\n%s", field, output)
		}
	}
}

func TestRenderSubmissionOmitsZeroCompleted(t *testing.T) {
	sub := model.Submission{
		ID:            "sub-123",
		ParticipantID: "part-456",
		Status:        "ACTIVE",
		StartedAt:     time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	output := submission.RenderSubmission(sub)

	if strings.Contains(output, "Completed:") {
		t.Fatalf("expected no Completed line for zero time, got:\n%s", output)
	}
}

func TestSubmissionListItemInterface(t *testing.T) {
	sub := model.Submission{
		ParticipantID: "part-456",
		Status:        "APPROVED",
		StudyCode:     "STUDY-789",
		TimeTaken:     900,
	}

	if sub.FilterValue() != "part-456" {
		t.Fatalf("expected FilterValue 'part-456', got '%s'", sub.FilterValue())
	}
	if sub.Title() != "part-456" {
		t.Fatalf("expected Title 'part-456', got '%s'", sub.Title())
	}
	expected := "APPROVED - STUDY-789 - 900s"
	if sub.Description() != expected {
		t.Fatalf("expected Description '%s', got '%s'", expected, sub.Description())
	}
}
