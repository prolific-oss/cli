package taskbuilder_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/taskbuilder"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

// testBatchUUID is a shared batch ID used across multiple taskbuilder tests.
const testBatchUUID = "01954894-65b3-779e-aaf6-348698e23634"

func TestNewGetBatchCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := taskbuilder.NewGetBatchCommand(c, os.Stdout)

	use := "view"
	short := "Get batch details"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewGetBatchCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := testBatchUUID

	createdAt, _ := time.Parse(time.RFC3339, "2025-02-27T18:03:59.795Z")
	response := client.GetAITaskBuilderBatchResponse{
		AITaskBuilderBatch: model.AITaskBuilderBatch{
			ID:                    batchID,
			CreatedAt:             createdAt,
			CreatedBy:             "6139f0d1dc08858054c63b2c",
			Name:                  "Test Batch",
			Status:                "UNINITIALISED",
			TotalTaskCount:        0,
			TotalInstructionCount: 5,
			WorkspaceID:           "6745ab669112d10b9b3afb48",
			SchemaVersion:         3,
			Datasets: []model.Dataset{
				{
					ID:                  "01954894-562f-71be-b2e0-adc7fdd7b3ea",
					TotalDatapointCount: 10,
				},
			},
			TaskDetails: model.TaskDetails{
				TaskName:         "Data Collection Task",
				TaskIntroduction: "This is a test task",
				TaskSteps:        "Step 1: Do something",
			},
		},
	}

	c.
		EXPECT().
		GetAITaskBuilderBatch(gomock.Eq(batchID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := taskbuilder.NewGetBatchCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", batchID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `Task Builder Batch Details:
ID: ` + batchID + `
Name: Test Batch
Status: UNINITIALISED
Total Task Count: 0
Total Instruction Count: 5
Workspace ID: 6745ab669112d10b9b3afb48
Created By: 6139f0d1dc08858054c63b2c
Created At: 2025-02-27 18:03:59
Schema Version: 3
Datasets: 1
  Dataset 1: 01954894-562f-71be-b2e0-adc7fdd7b3ea (10 datapoints)

Task Details:
  Name: Data Collection Task
  Introduction: This is a test task
  Steps: Step 1: Do something
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewGetBatchCommandCallsAPIWithoutOptionalFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23699"

	createdAt, _ := time.Parse(time.RFC3339, "2025-02-27T18:03:59.795Z")
	response := client.GetAITaskBuilderBatchResponse{
		AITaskBuilderBatch: model.AITaskBuilderBatch{
			ID:                    batchID,
			CreatedAt:             createdAt,
			CreatedBy:             "6139f0d1dc08858054c63b2c",
			Name:                  "Simple Batch",
			Status:                "ACTIVE",
			TotalTaskCount:        5,
			TotalInstructionCount: 0,
			WorkspaceID:           "6745ab669112d10b9b3afb48",
			SchemaVersion:         1,
			Datasets:              []model.Dataset{},   // Empty datasets
			TaskDetails:           model.TaskDetails{}, // Empty task details
		},
	}

	c.
		EXPECT().
		GetAITaskBuilderBatch(gomock.Eq(batchID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := taskbuilder.NewGetBatchCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", batchID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `Task Builder Batch Details:
ID: 01954894-65b3-779e-aaf6-348698e23699
Name: Simple Batch
Status: ACTIVE
Total Task Count: 5
Total Instruction Count: 0
Workspace ID: 6745ab669112d10b9b3afb48
Created By: 6139f0d1dc08858054c63b2c
Created At: 2025-02-27 18:03:59
Schema Version: 1
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewGetBatchCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "an-invalid-batch-id"
	errorMessage := taskbuilder.ErrBatchNotFound

	c.
		EXPECT().
		GetAITaskBuilderBatch(gomock.Eq(batchID)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := taskbuilder.NewGetBatchCommand(c, os.Stdout)
	_ = cmd.Flags().Set("batch-id", batchID)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewGetBatchCommandRequiresBatchID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := taskbuilder.NewGetBatchCommand(c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Fatal("expected error when batch-id is missing")
	}

	if !cmd.Flags().Changed("batch-id") {
		expected := taskbuilder.ErrBatchIDRequired
		if err.Error() != "error: "+expected {
			t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
		}
	}
}

// reviewBatchJSON is a review batch as the API returns it: the form is the stored items, so the rubrics and the notes
// box are read back from them.
const reviewBatchJSON = `{
	"id": "review-batch-1",
	"created_at": "2025-02-27T18:03:59.795Z",
	"created_by": "6139f0d1dc08858054c63b2c",
	"name": "Review Batch",
	"status": "UNINITIALISED",
	"workspace_id": "6745ab669112d10b9b3afb48",
	"schema_version": 6,
	"review_source_batch_id": "source-batch-1",
	"batch_items": [{"rows": [
		{"columns": [{"items": [{"type": "rubric_score", "description": "Is the narration accurate?", "scoring_scheme": "pass_fail_unsure"}]}]},
		{"columns": [{"items": [{"type": "rubric_score", "description": "How clear is it?", "scoring_scheme": "scale_1_5"}]}]},
		{"columns": [{"items": [{"type": "free_text", "description": "<p>Add notes (optional)</p>", "optional": true}]}]}
	]}]
}`

func TestNewGetBatchCommandShowsAReviewBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var batch model.AITaskBuilderBatch
	if err := json.Unmarshal([]byte(reviewBatchJSON), &batch); err != nil {
		t.Fatal(err)
	}
	c.EXPECT().GetAITaskBuilderBatch("review-batch-1").Return(&client.GetAITaskBuilderBatchResponse{AITaskBuilderBatch: batch}, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := taskbuilder.NewGetBatchCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", "review-batch-1")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("expected error to be nil; got %v", err)
	}
	writer.Flush()

	expected := `
Review:
  Source Batch ID: source-batch-1
  Notes Box: true
  Rubrics: 2
    1. Is the narration accurate? (pass_fail_unsure)
    2. How clear is it? (scale_1_5)
`
	if !strings.Contains(b.String(), expected) {
		t.Fatalf("expected output to contain:\n%s\ngot output:\n%s", expected, b.String())
	}
}

func TestAITaskBuilderBatchReviewReadsOnlyTheNotesBoxTheAPIWrites(t *testing.T) {
	var batch model.AITaskBuilderBatch
	err := json.Unmarshal([]byte(`{
		"review_source_batch_id": "source-batch-1",
		"batch_items": [{"rows": [{"columns": [{"items": [{"type": "free_text", "description": "<p>Anything else?</p>"}]}]}]}]
	}`), &batch)
	if err != nil {
		t.Fatal(err)
	}

	review := batch.Review()
	if review == nil || review.Notes || len(review.Rubrics) != 0 {
		t.Fatalf("expected a review with no notes box and no rubrics; got %+v", review)
	}

	if (model.AITaskBuilderBatch{ID: "not-a-review"}).Review() != nil {
		t.Fatal("expected no review for a batch with no source batch")
	}
}

// batch_items is only read for a review batch. A batch of any other shape must still load and show as it always did, so
// the form is not decoded until a review batch asks for it.
func TestAITaskBuilderBatchLoadsWhateverItsBatchItemsAre(t *testing.T) {
	var batch model.AITaskBuilderBatch
	err := json.Unmarshal([]byte(`{
		"id": "batch-1",
		"batch_items": [{"rows": [{"columns": [{"items": [{"type": "free_text", "description": {"not": "a string"}}]}]}]}]
	}`), &batch)
	if err != nil {
		t.Fatalf("expected a batch with an unusual form to load; got %v", err)
	}

	if batch.Review() != nil {
		t.Fatal("expected no review for a batch with no source batch")
	}
}

func TestNewGetBatchCommandSaysWhenAReviewFormCannotBeRead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var batch model.AITaskBuilderBatch
	err := json.Unmarshal([]byte(`{"id": "review-batch-1", "review_source_batch_id": "source-batch-1", "batch_items": "not a form"}`), &batch)
	if err != nil {
		t.Fatalf("expected the batch to load; got %v", err)
	}
	c.EXPECT().GetAITaskBuilderBatch("review-batch-1").Return(&client.GetAITaskBuilderBatchResponse{AITaskBuilderBatch: batch}, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := taskbuilder.NewGetBatchCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", "review-batch-1")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("expected error to be nil; got %v", err)
	}
	writer.Flush()

	expected := "Review:\n  Source Batch ID: source-batch-1\n  Rubrics: could not be read from the batch's form\n"
	if !strings.Contains(b.String(), expected) {
		t.Fatalf("expected output to contain:\n%s\ngot output:\n%s", expected, b.String())
	}
	if strings.Contains(b.String(), "Notes Box") {
		t.Fatalf("expected no notes box line when the form could not be read; got\n%s", b.String())
	}
}
