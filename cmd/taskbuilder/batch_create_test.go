package taskbuilder_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/taskbuilder"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewBatchCreateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := taskbuilder.NewBatchCreateCommand(c, os.Stdout)

	use := createCommandUse
	short := "Create a batch"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewBatchCreateCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchName := "Test Data Collection Batch"
	workspaceID := "6278acb09062db3b35bcbeb0"
	datasetID := "1234acb09999db4b99bcded1"

	taskName := "Sample Task"
	taskIntroduction := "This is a sample task for testing"
	taskSteps := "1. Review the data\n2. Provide your response"

	response := &client.CreateAITaskBuilderBatchResponse{
		ID:                    "497f6eca-6276-4993-bfeb-53cbbbba6f08",
		CreatedAt:             "2019-08-24T14:15:22Z",
		CreatedBy:             "6278cb09062dbb35bc4abebc",
		Name:                  batchName,
		Status:                "UNINITIALISED",
		TotalTaskCount:        0,
		TotalInstructionCount: 0,
		WorkspaceID:           workspaceID,
		Datasets: []model.Dataset{
			{
				ID:                  datasetID,
				TotalDatapointCount: 100,
			},
		},
		AutoSyncEnabled: false,
		TaskDetails: model.TaskDetails{
			TaskName:         taskName,
			TaskIntroduction: taskIntroduction,
			TaskSteps:        taskSteps,
		},
	}

	c.EXPECT().CreateAITaskBuilderBatch(client.CreateBatchParams{
		Name:             batchName,
		WorkspaceID:      workspaceID,
		DatasetID:        datasetID,
		TaskName:         taskName,
		TaskIntroduction: taskIntroduction,
		TaskSteps:        taskSteps,
	}).Return(response, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := taskbuilder.NewBatchCreateCommand(c, writer)
	cmd.SetArgs([]string{
		"--name", batchName,
		"--workspace-id", workspaceID,
		"--dataset-id", datasetID,
		"--task-name", taskName,
		"--task-introduction", taskIntroduction,
		"--task-steps", taskSteps,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected error to be nil; got %v", err)
	}

	writer.Flush()

	expected := fmt.Sprintf("Task Builder Batch Created Successfully:\nID: %s\nName: %s\nStatus: %s\nAuto Sync Enabled: %v\nTotal Task Count: %d\nTotal Instruction Count: %d\nWorkspace ID: %s\nCreated By: %s\nCreated At: %s\nDatasets: %d\n  Dataset 1: %s (100 datapoints)\n\nTask Details:\n  Name: %s\n  Introduction: %s\n  Steps: %s\n",
		response.ID, response.Name, response.Status, response.AutoSyncEnabled, response.TotalTaskCount, response.TotalInstructionCount, response.WorkspaceID, response.CreatedBy, response.CreatedAt, len(response.Datasets), datasetID, taskName, taskIntroduction, taskSteps)

	if b.String() != expected {
		t.Fatalf("expected output:\n%s\ngot output:\n%s", expected, b.String())
	}
}

func TestNewBatchCreateCommandAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchName := "Test Data Collection Batch"
	workspaceID := "6278acb09062db3b35bcbeb0"
	datasetID := "1234acb09999db4b99bcded1"
	taskName := "Sample Task"
	taskIntroduction := "This is a sample task for testing"
	taskSteps := "1. Review the data\n2. Provide your response"

	c.EXPECT().CreateAITaskBuilderBatch(client.CreateBatchParams{
		Name:             batchName,
		WorkspaceID:      workspaceID,
		DatasetID:        datasetID,
		TaskName:         taskName,
		TaskIntroduction: taskIntroduction,
		TaskSteps:        taskSteps,
	}).Return(nil, errors.New("API error"))

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := taskbuilder.NewBatchCreateCommand(c, writer)
	cmd.SetArgs([]string{
		"--name", batchName,
		"--workspace-id", workspaceID,
		"--dataset-id", datasetID,
		"--task-name", taskName,
		"--task-introduction", taskIntroduction,
		"--task-steps", taskSteps,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error; got nil")
	}

	expectedError := apiError
	if err.Error() != expectedError {
		t.Fatalf("expected error: %s; got %s", expectedError, err.Error())
	}
}

func TestNewBatchCreateCommandMissingRequiredFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	testCases := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "missing name flag",
			args:        []string{"--workspace-id", "6278acb09062db3b35bcbeb0", "--dataset-id", "1234acb09999db4b99bcded1", "--task-name", "Sample Task", "--task-introduction", "Introduction", "--task-steps", "Steps"},
			expectedErr: `required flag(s) "name" not set`,
		},
		{
			name:        "missing task-name flag",
			args:        []string{"--name", "Test Batch", "--workspace-id", "6278acb09062db3b35bcbeb0", "--dataset-id", "1234acb09999db4b99bcded1", "--task-introduction", "Introduction", "--task-steps", "Steps"},
			expectedErr: `required flag(s) "task-name" not set`,
		},
		{
			name:        "missing task-introduction flag",
			args:        []string{"--name", "Test Batch", "--workspace-id", "6278acb09062db3b35bcbeb0", "--dataset-id", "1234acb09999db4b99bcded1", "--task-name", "Sample Task", "--task-steps", "Steps"},
			expectedErr: `required flag(s) "task-introduction" not set`,
		},
		{
			name:        "missing task-steps flag",
			args:        []string{"--name", "Test Batch", "--workspace-id", "6278acb09062db3b35bcbeb0", "--dataset-id", "1234acb09999db4b99bcded1", "--task-name", "Sample Task", "--task-introduction", "Introduction"},
			expectedErr: `required flag(s) "task-steps" not set`,
		},
		{
			name:        "missing all flags",
			args:        []string{},
			expectedErr: `required flag(s) "name", "task-introduction", "task-name", "task-steps", "workspace-id" not set`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := taskbuilder.NewBatchCreateCommand(c, writer)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error; got nil")
			}

			if err.Error() != tc.expectedErr {
				t.Fatalf("expected error: %s; got %s", tc.expectedErr, err.Error())
			}
		})
	}
}

const batchItemsJSON = `[{"rows":[{"columns":[{"items":[{"type":"free_text","description":"Describe your observations."}]}]}]}]`

func baseCreateArgs() []string {
	return []string{
		"--name", "Test Batch",
		"--workspace-id", "6278acb09062db3b35bcbeb0",
		"--dataset-id", "1234acb09999db4b99bcded1",
		"--task-name", "Task",
		"--task-introduction", "Intro",
		"--task-steps", "Steps",
	}
}

func TestNewBatchCreateCommandWithBatchItemsJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := &client.CreateAITaskBuilderBatchResponse{
		ID:          "497f6eca-6276-4993-bfeb-53cbbbba6f08",
		CreatedAt:   "2019-08-24T14:15:22Z",
		CreatedBy:   "user-1",
		Name:        "Test Batch",
		Status:      "UNINITIALISED",
		WorkspaceID: "6278acb09062db3b35bcbeb0",
	}

	c.EXPECT().CreateAITaskBuilderBatch(client.CreateBatchParams{
		Name:             "Test Batch",
		WorkspaceID:      "6278acb09062db3b35bcbeb0",
		DatasetID:        "1234acb09999db4b99bcded1",
		TaskName:         "Task",
		TaskIntroduction: "Intro",
		TaskSteps:        "Steps",
		BatchItems:       json.RawMessage(batchItemsJSON),
	}).Return(response, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := taskbuilder.NewBatchCreateCommand(c, writer)
	cmd.SetArgs(append(baseCreateArgs(), "--batch-items-json", batchItemsJSON))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error; got %v", err)
	}
}

func TestNewBatchCreateCommandWithBatchItemsFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	tmpFile := filepath.Join(t.TempDir(), "batch-items.json")
	if err := os.WriteFile(tmpFile, []byte(batchItemsJSON), 0600); err != nil {
		t.Fatal(err)
	}

	response := &client.CreateAITaskBuilderBatchResponse{
		ID:          "497f6eca-6276-4993-bfeb-53cbbbba6f08",
		CreatedAt:   "2019-08-24T14:15:22Z",
		CreatedBy:   "user-1",
		Name:        "Test Batch",
		Status:      "UNINITIALISED",
		WorkspaceID: "6278acb09062db3b35bcbeb0",
	}

	c.EXPECT().CreateAITaskBuilderBatch(client.CreateBatchParams{
		Name:             "Test Batch",
		WorkspaceID:      "6278acb09062db3b35bcbeb0",
		DatasetID:        "1234acb09999db4b99bcded1",
		TaskName:         "Task",
		TaskIntroduction: "Intro",
		TaskSteps:        "Steps",
		BatchItems:       json.RawMessage(batchItemsJSON),
	}).Return(response, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := taskbuilder.NewBatchCreateCommand(c, writer)
	cmd.SetArgs(append(baseCreateArgs(), "--batch-items-file", tmpFile))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error; got %v", err)
	}
}

func TestNewBatchCreateCommandWithAutoSync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := &client.CreateAITaskBuilderBatchResponse{
		ID:          "497f6eca-6276-4993-bfeb-53cbbbba6f08",
		CreatedAt:   "2019-08-24T14:15:22Z",
		CreatedBy:   "user-1",
		Name:        "Test Batch",
		Status:      "UNINITIALISED",
		WorkspaceID: "6278acb09062db3b35bcbeb0",
	}

	c.EXPECT().CreateAITaskBuilderBatch(client.CreateBatchParams{
		Name:             "Test Batch",
		WorkspaceID:      "6278acb09062db3b35bcbeb0",
		DatasetID:        "1234acb09999db4b99bcded1",
		TaskName:         "Task",
		TaskIntroduction: "Intro",
		TaskSteps:        "Steps",
		AutoSync:         true,
	}).Return(response, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := taskbuilder.NewBatchCreateCommand(c, writer)
	cmd.SetArgs(append(baseCreateArgs(), "--auto-sync"))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error; got %v", err)
	}
}

func TestNewBatchCreateCommandBatchItemsValidationErrors(t *testing.T) {
	testCases := []struct {
		name        string
		extraArgs   []string
		expectedErr string
	}{
		{
			name:        "both file and json flags",
			extraArgs:   []string{"-f", "some-file.json", "-j", batchItemsJSON},
			expectedErr: "error: " + taskbuilder.ErrBothBatchItemsInputsProvided,
		},
		{
			name:        "invalid JSON",
			extraArgs:   []string{"-j", "not-json"},
			expectedErr: "error: " + taskbuilder.ErrBatchItemsMustBeArray,
		},
		{
			name:        "non-array JSON",
			extraArgs:   []string{"-j", `{"rows":[]}`},
			expectedErr: "error: " + taskbuilder.ErrBatchItemsMustBeArray,
		},
		{
			name:        "empty array",
			extraArgs:   []string{"-j", "[]"},
			expectedErr: "error: " + taskbuilder.ErrBatchItemsMustBeNonEmpty,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := taskbuilder.NewBatchCreateCommand(c, writer)
			cmd.SetArgs(append(baseCreateArgs(), tc.extraArgs...))

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error; got nil")
			}
			if err.Error() != tc.expectedErr {
				t.Fatalf("expected error: %q; got %q", tc.expectedErr, err.Error())
			}
		})
	}
}

const reviewRubricsJSON = `[{"description":"Is the narration accurate?","scoring_scheme":"pass_fail_unsure"},{"description":"How clear is it?","scoring_scheme":"scale_1_5"}]`

// reviewCreateArgs are the flags for a review batch: everything in baseCreateArgs but the dataset.
func reviewCreateArgs(extra ...string) []string {
	args := []string{
		"--name", "Test Batch",
		"--workspace-id", "6278acb09062db3b35bcbeb0",
		"--task-name", "Task",
		"--task-introduction", "Intro",
		"--task-steps", "Steps",
		"--review-source-batch-id", "source-batch-1",
		"--review-sample-rate-bps", "1000",
	}
	return append(args, extra...)
}

func TestNewBatchCreateCommandReview(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().CreateAITaskBuilderBatch(client.CreateBatchParams{
		Name:             "Test Batch",
		WorkspaceID:      "6278acb09062db3b35bcbeb0",
		TaskName:         "Task",
		TaskIntroduction: "Intro",
		TaskSteps:        "Steps",
		Review: &client.ReviewParams{
			SourceBatchID: "source-batch-1",
			SampleRateBps: 1000,
			Rubrics: []client.ReviewRubric{
				{Description: "Is the narration accurate?", ScoringScheme: "pass_fail_unsure"},
				{Description: "How clear is it?", ScoringScheme: "scale_1_5"},
			},
			Justification: true,
		},
	}).Return(&client.CreateAITaskBuilderBatchResponse{
		ID:     "review-batch-1",
		Name:   "Test Batch",
		Status: "UNINITIALISED",
		Review: &client.ReviewResult{Round: 1, SampleRateBps: 1000, ReviewDatasetID: "review-dataset-1", BackfillRequested: true},
	}, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := taskbuilder.NewBatchCreateCommand(c, writer)
	cmd.SetArgs(reviewCreateArgs("--review-notes", "--review-rubrics-json", reviewRubricsJSON))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected error to be nil; got %v", err)
	}
	writer.Flush()

	expected := "\nReview:\n  Round: 1\n  Sample Rate: 1000 bps\n  Review Dataset ID: review-dataset-1\n  Backfill Requested: true\n"
	if !strings.Contains(b.String(), expected) {
		t.Fatalf("expected output to contain:\n%s\ngot output:\n%s", expected, b.String())
	}
}

func TestNewBatchCreateCommandReviewLimitedToOneStudy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	const studyID = "6ac6ce260218ce7b7b0e05d3"
	c.EXPECT().CreateAITaskBuilderBatch(gomock.Any()).DoAndReturn(func(p client.CreateBatchParams) (*client.CreateAITaskBuilderBatchResponse, error) {
		if p.Review == nil || p.Review.StudyID != studyID {
			t.Fatalf("expected the review to be limited to study %s; got %+v", studyID, p.Review)
		}
		return &client.CreateAITaskBuilderBatchResponse{
			Review: &client.ReviewResult{Round: 1, StudyID: studyID, SampleRateBps: 1000, ReviewDatasetID: "rd-1"},
		}, nil
	})

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	cmd := taskbuilder.NewBatchCreateCommand(c, writer)
	cmd.SetArgs(reviewCreateArgs("--review-study-id", studyID, "--review-rubrics-json", reviewRubricsJSON))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected error to be nil; got %v", err)
	}
	writer.Flush()

	if !strings.Contains(b.String(), "  Round: 1\n  Study ID: "+studyID+"\n") {
		t.Fatalf("expected the study to be shown in the review block; got\n%s", b.String())
	}
	if strings.Contains(b.String(), "Warning") {
		t.Fatalf("expected no warning when the API confirmed the study; got\n%s", b.String())
	}
}

func TestNewBatchCreateCommandWarnsWhenTheStudyWasNotConfirmed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// An API that predates study_id ignores it and does not echo it back.
	c.EXPECT().CreateAITaskBuilderBatch(gomock.Any()).Return(&client.CreateAITaskBuilderBatchResponse{
		Review: &client.ReviewResult{Round: 1},
	}, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	cmd := taskbuilder.NewBatchCreateCommand(c, writer)
	cmd.SetArgs(reviewCreateArgs("--review-study-id", "6ac6ce260218ce7b7b0e05d3", "--review-rubrics-json", reviewRubricsJSON))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected error to be nil; got %v", err)
	}
	writer.Flush()

	if !strings.Contains(b.String(), "Warning: the API did not confirm study 6ac6ce260218ce7b7b0e05d3") {
		t.Fatalf("expected a warning; got\n%s", b.String())
	}
}

func TestNewBatchCreateCommandReviewWithoutAStudyCoversTheBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().CreateAITaskBuilderBatch(gomock.Any()).DoAndReturn(func(p client.CreateBatchParams) (*client.CreateAITaskBuilderBatchResponse, error) {
		if p.Review == nil || p.Review.StudyID != "" {
			t.Fatalf("expected no study; got %+v", p.Review)
		}
		return &client.CreateAITaskBuilderBatchResponse{Review: &client.ReviewResult{Round: 1}}, nil
	})

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	cmd := taskbuilder.NewBatchCreateCommand(c, writer)
	cmd.SetArgs(reviewCreateArgs("--review-rubrics-json", reviewRubricsJSON))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected error to be nil; got %v", err)
	}
	writer.Flush()

	if strings.Contains(b.String(), "Study ID") {
		t.Fatalf("expected no study line when the review covers the whole batch; got\n%s", b.String())
	}
}

func TestNewBatchCreateCommandReviewRubricsFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	path := filepath.Join(t.TempDir(), "rubrics.json")
	if err := os.WriteFile(path, []byte(reviewRubricsJSON), 0o600); err != nil {
		t.Fatal(err)
	}

	c.EXPECT().CreateAITaskBuilderBatch(gomock.Any()).DoAndReturn(func(p client.CreateBatchParams) (*client.CreateAITaskBuilderBatchResponse, error) {
		if p.Review == nil || len(p.Review.Rubrics) != 2 || p.Review.Justification {
			t.Fatalf("expected two rubrics and no notes box; got %+v", p.Review)
		}
		return &client.CreateAITaskBuilderBatchResponse{}, nil
	})

	var b bytes.Buffer
	cmd := taskbuilder.NewBatchCreateCommand(c, bufio.NewWriter(&b))
	cmd.SetArgs(reviewCreateArgs("--review-rubrics-file", path))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected error to be nil; got %v", err)
	}
}

func TestNewBatchCreateCommandReviewValidationErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl) // no call is expected: each of these stops before the API

	withRubrics := []string{"--review-rubrics-json", reviewRubricsJSON}
	without := func(flag string, args []string) []string {
		var kept []string
		for i := 0; i < len(args); i++ {
			if args[i] == flag {
				i++
				continue
			}
			kept = append(kept, args[i])
		}
		return kept
	}

	testCases := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "no source batch",
			args:        append(without("--review-source-batch-id", reviewCreateArgs()), withRubrics...),
			expectedErr: "error: " + taskbuilder.ErrReviewSourceBatchIDRequired,
		},
		{
			name:        "no sample rate",
			args:        append(without("--review-sample-rate-bps", reviewCreateArgs()), withRubrics...),
			expectedErr: "error: " + taskbuilder.ErrReviewSampleRateInvalid,
		},
		{
			name:        "sample rate over every submission",
			args:        reviewCreateArgs(append([]string{"--review-sample-rate-bps", "10001"}, withRubrics...)...),
			expectedErr: "error: " + taskbuilder.ErrReviewSampleRateInvalid,
		},
		{
			name:        "no rubrics",
			args:        reviewCreateArgs(),
			expectedErr: "error: " + taskbuilder.ErrReviewRubricsRequired,
		},
		{
			name:        "rubrics not an array",
			args:        reviewCreateArgs("--review-rubrics-json", `{"description":"x"}`),
			expectedErr: "error: " + taskbuilder.ErrReviewRubricsMustBeArray,
		},
		{
			name:        "empty rubrics",
			args:        reviewCreateArgs("--review-rubrics-json", `[]`),
			expectedErr: "error: " + taskbuilder.ErrReviewRubricsCount,
		},
		{
			name:        "rubric without a scheme",
			args:        reviewCreateArgs("--review-rubrics-json", `[{"description":"x"}]`),
			expectedErr: "error: review rubric 1 needs a description and a scoring_scheme",
		},
		{
			name:        "both rubrics inputs",
			args:        reviewCreateArgs(append([]string{"--review-rubrics-file", "r.json"}, withRubrics...)...),
			expectedErr: "error: " + taskbuilder.ErrBothReviewRubricsInputsProvided,
		},
		{
			name:        "dataset given with a review",
			args:        reviewCreateArgs(append([]string{"--dataset-id", "1234acb09999db4b99bcded1"}, withRubrics...)...),
			expectedErr: "error: " + taskbuilder.ErrReviewDatasetNotAllowed,
		},
		{
			name:        "batch items given with a review",
			args:        reviewCreateArgs(append([]string{"--batch-items-json", batchItemsJSON}, withRubrics...)...),
			expectedErr: "error: " + taskbuilder.ErrReviewBatchItemsNotAllowed,
		},
		{
			name:        "a study id that is not a study id",
			args:        reviewCreateArgs(append([]string{"--review-study-id", "not-an-id"}, withRubrics...)...),
			expectedErr: "error: " + taskbuilder.ErrReviewStudyIDInvalid,
		},
		{
			name:        "a study id in capitals, which the API refuses",
			args:        reviewCreateArgs(append([]string{"--review-study-id", "6AC6CE260218CE7B7B0E05D3"}, withRubrics...)...),
			expectedErr: "error: " + taskbuilder.ErrReviewStudyIDInvalid,
		},
		{
			name:        "a study id alone is still a review, and needs its source",
			args:        append(baseCreateArgs(), "--review-study-id", "6ac6ce260218ce7b7b0e05d3"),
			expectedErr: "error: " + taskbuilder.ErrReviewSourceBatchIDRequired,
		},
		{
			name:        "a review flag alone is still a review",
			args:        append(baseCreateArgs(), "--review-notes"),
			expectedErr: "error: " + taskbuilder.ErrReviewSourceBatchIDRequired,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			cmd := taskbuilder.NewBatchCreateCommand(c, bufio.NewWriter(&b))
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if err == nil || err.Error() != tc.expectedErr {
				t.Fatalf("expected error: %s; got %v", tc.expectedErr, err)
			}
		})
	}
}

func TestNewBatchCreateCommandDatasetStillRequiredWithoutReview(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var args []string
	for i := 0; i < len(baseCreateArgs()); i++ {
		if baseCreateArgs()[i] == "--dataset-id" {
			i++
			continue
		}
		args = append(args, baseCreateArgs()[i])
	}

	var b bytes.Buffer
	cmd := taskbuilder.NewBatchCreateCommand(c, bufio.NewWriter(&b))
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err == nil || err.Error() != "error: "+taskbuilder.ErrDatasetIDRequired {
		t.Fatalf("expected the dataset ID error; got %v", err)
	}
}
