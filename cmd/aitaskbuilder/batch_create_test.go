package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewBatchCreateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewBatchCreateCommand(c, os.Stdout)

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
		Datasets: []client.DatasetReference{
			{
				ID:                  datasetID,
				TotalDatapointCount: 100,
			},
		},
		TaskDetails: client.TaskDetailsResponse{
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

	cmd := aitaskbuilder.NewBatchCreateCommand(c, writer)
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

	expected := fmt.Sprintf("AI Task Builder Batch Created Successfully:\nID: %s\nName: %s\nStatus: %s\nTotal Task Count: %d\nTotal Instruction Count: %d\nWorkspace ID: %s\nCreated By: %s\nCreated At: %s\nDatasets: %d\n  Dataset 1: %s (100 datapoints)\n\nTask Details:\n  Name: %s\n  Introduction: %s\n  Steps: %s\n",
		response.ID, response.Name, response.Status, response.TotalTaskCount, response.TotalInstructionCount, response.WorkspaceID, response.CreatedBy, response.CreatedAt, len(response.Datasets), datasetID, taskName, taskIntroduction, taskSteps)

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

	cmd := aitaskbuilder.NewBatchCreateCommand(c, writer)
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
			expectedErr: `required flag(s) "dataset-id", "name", "task-introduction", "task-name", "task-steps", "workspace-id" not set`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := aitaskbuilder.NewBatchCreateCommand(c, writer)
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

	cmd := aitaskbuilder.NewBatchCreateCommand(c, writer)
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

	cmd := aitaskbuilder.NewBatchCreateCommand(c, writer)
	cmd.SetArgs(append(baseCreateArgs(), "--batch-items-file", tmpFile))

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
			expectedErr: "error: " + aitaskbuilder.ErrBothBatchItemsInputsProvided,
		},
		{
			name:        "invalid JSON",
			extraArgs:   []string{"-j", "not-json"},
			expectedErr: "error: " + aitaskbuilder.ErrBatchItemsMustBeArray,
		},
		{
			name:        "non-array JSON",
			extraArgs:   []string{"-j", `{"rows":[]}`},
			expectedErr: "error: " + aitaskbuilder.ErrBatchItemsMustBeArray,
		},
		{
			name:        "empty array",
			extraArgs:   []string{"-j", "[]"},
			expectedErr: "error: " + aitaskbuilder.ErrBatchItemsMustBeNonEmpty,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := aitaskbuilder.NewBatchCreateCommand(c, writer)
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
