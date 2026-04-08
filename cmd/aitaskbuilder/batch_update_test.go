package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

const updateBatchID = "497f6eca-6276-4993-bfeb-53cbbbba6f08"

func TestNewBatchUpdateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewBatchUpdateCommand(c, os.Stdout)

	use := "update"
	short := "Update a batch"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewBatchUpdateCommandUpdatesName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := updateBatchID
	batchName := "Updated Batch Name"

	createdAt, _ := time.Parse(time.RFC3339, "2025-02-27T18:03:59.795Z")
	response := &client.UpdateAITaskBuilderBatchResponse{
		AITaskBuilderBatch: model.AITaskBuilderBatch{
			ID:                    batchID,
			CreatedAt:             createdAt,
			CreatedBy:             "6139f0d1dc08858054c63b2c",
			Name:                  batchName,
			Status:                "UNINITIALISED",
			TotalTaskCount:        0,
			TotalInstructionCount: 5,
			WorkspaceID:           "6745ab669112d10b9b3afb48",
			SchemaVersion:         3,
			Datasets:              []model.Dataset{},
			TaskDetails:           model.TaskDetails{},
		},
	}

	c.EXPECT().UpdateAITaskBuilderBatch(client.UpdateBatchParams{
		BatchID: batchID,
		Name:    batchName,
	}).Return(response, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewBatchUpdateCommand(c, writer)
	cmd.SetArgs([]string{"--batch-id", batchID, "--name", batchName})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected error to be nil; got %v", err)
	}

	writer.Flush()

	expected := fmt.Sprintf("AI Task Builder Batch Updated Successfully:\nID: %s\nName: %s\nStatus: %s\nTotal Task Count: %d\nTotal Instruction Count: %d\nWorkspace ID: %s\nCreated By: %s\nCreated At: %s\nSchema Version: %d\n",
		response.ID, response.Name, response.Status, response.TotalTaskCount, response.TotalInstructionCount,
		response.WorkspaceID, response.CreatedBy, "2025-02-27 18:03:59", response.SchemaVersion)

	if b.String() != expected {
		t.Fatalf("expected output:\n%s\ngot output:\n%s", expected, b.String())
	}
}

func TestNewBatchUpdateCommandUpdatesAllTaskDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := updateBatchID
	taskName := "Updated Task"
	taskIntroduction := "Updated introduction"
	taskSteps := "1. Updated step"

	createdAt, _ := time.Parse(time.RFC3339, "2025-02-27T18:03:59.795Z")
	response := &client.UpdateAITaskBuilderBatchResponse{
		AITaskBuilderBatch: model.AITaskBuilderBatch{
			ID:            batchID,
			CreatedAt:     createdAt,
			CreatedBy:     "6139f0d1dc08858054c63b2c",
			Name:          "Existing Name",
			Status:        "UNINITIALISED",
			WorkspaceID:   "6745ab669112d10b9b3afb48",
			SchemaVersion: 3,
			Datasets:      []model.Dataset{},
			TaskDetails: model.TaskDetails{
				TaskName:         taskName,
				TaskIntroduction: taskIntroduction,
				TaskSteps:        taskSteps,
			},
		},
	}

	// All three task detail flags provided — no GET call expected
	c.EXPECT().UpdateAITaskBuilderBatch(client.UpdateBatchParams{
		BatchID:          batchID,
		TaskName:         taskName,
		TaskIntroduction: taskIntroduction,
		TaskSteps:        taskSteps,
	}).Return(response, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewBatchUpdateCommand(c, writer)
	cmd.SetArgs([]string{
		"--batch-id", batchID,
		"--task-name", taskName,
		"--task-introduction", taskIntroduction,
		"--task-steps", taskSteps,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected error to be nil; got %v", err)
	}

	writer.Flush()

	expected := fmt.Sprintf("AI Task Builder Batch Updated Successfully:\nID: %s\nName: %s\nStatus: %s\nTotal Task Count: %d\nTotal Instruction Count: %d\nWorkspace ID: %s\nCreated By: %s\nCreated At: %s\nSchema Version: %d\n\nTask Details:\n  Name: %s\n  Introduction: %s\n  Steps: %s\n",
		response.ID, response.Name, response.Status, response.TotalTaskCount, response.TotalInstructionCount,
		response.WorkspaceID, response.CreatedBy, "2025-02-27 18:03:59", response.SchemaVersion,
		taskName, taskIntroduction, taskSteps)

	if b.String() != expected {
		t.Fatalf("expected output:\n%s\ngot output:\n%s", expected, b.String())
	}
}

func TestNewBatchUpdateCommandMergesPartialTaskDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := updateBatchID
	newTaskName := "Updated Task Name"
	existingIntroduction := "Existing introduction"
	existingSteps := "Existing steps"

	createdAt, _ := time.Parse(time.RFC3339, "2025-02-27T18:03:59.795Z")

	existingBatch := &client.GetAITaskBuilderBatchResponse{
		AITaskBuilderBatch: model.AITaskBuilderBatch{
			ID:        batchID,
			CreatedAt: createdAt,
			TaskDetails: model.TaskDetails{
				TaskName:         "Old Task Name",
				TaskIntroduction: existingIntroduction,
				TaskSteps:        existingSteps,
			},
		},
	}

	// Partial task details trigger a GET to fetch existing values
	c.EXPECT().GetAITaskBuilderBatch(batchID).Return(existingBatch, nil)

	updateResponse := &client.UpdateAITaskBuilderBatchResponse{
		AITaskBuilderBatch: model.AITaskBuilderBatch{
			ID:            batchID,
			CreatedAt:     createdAt,
			CreatedBy:     "6139f0d1dc08858054c63b2c",
			Name:          "Existing Name",
			Status:        "UNINITIALISED",
			WorkspaceID:   "6745ab669112d10b9b3afb48",
			SchemaVersion: 3,
			Datasets:      []model.Dataset{},
			TaskDetails: model.TaskDetails{
				TaskName:         newTaskName,
				TaskIntroduction: existingIntroduction,
				TaskSteps:        existingSteps,
			},
		},
	}

	// Merged params: new task name + existing introduction and steps
	c.EXPECT().UpdateAITaskBuilderBatch(client.UpdateBatchParams{
		BatchID:          batchID,
		TaskName:         newTaskName,
		TaskIntroduction: existingIntroduction,
		TaskSteps:        existingSteps,
	}).Return(updateResponse, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewBatchUpdateCommand(c, writer)
	cmd.SetArgs([]string{"--batch-id", batchID, "--task-name", newTaskName})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected error to be nil; got %v", err)
	}

	writer.Flush()

	if b.Len() == 0 {
		t.Fatal("expected output; got none")
	}
}

func TestNewBatchUpdateCommandGetErrorOnPartialTaskDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := updateBatchID

	c.EXPECT().GetAITaskBuilderBatch(batchID).Return(nil, errors.New("batch not found"))

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewBatchUpdateCommand(c, writer)
	cmd.SetArgs([]string{"--batch-id", batchID, "--task-name", "New Task"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error; got nil")
	}

	expectedError := "error: batch not found"
	if err.Error() != expectedError {
		t.Fatalf("expected error: %s; got %s", expectedError, err.Error())
	}
}

func TestNewBatchUpdateCommandAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := updateBatchID

	c.EXPECT().UpdateAITaskBuilderBatch(client.UpdateBatchParams{
		BatchID: batchID,
		Name:    "New Name",
	}).Return(nil, errors.New("API error"))

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewBatchUpdateCommand(c, writer)
	cmd.SetArgs([]string{"--batch-id", batchID, "--name", "New Name"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error; got nil")
	}

	expectedError := apiError
	if err.Error() != expectedError {
		t.Fatalf("expected error: %s; got %s", expectedError, err.Error())
	}
}

func TestNewBatchUpdateCommandMissingBatchID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewBatchUpdateCommand(c, writer)
	cmd.SetArgs([]string{"--name", "New Name"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error; got nil")
	}

	expectedError := `required flag(s) "batch-id" not set`
	if err.Error() != expectedError {
		t.Fatalf("expected error: %s; got %s", expectedError, err.Error())
	}
}

func TestNewBatchUpdateCommandNoFieldsProvided(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewBatchUpdateCommand(c, writer)
	cmd.SetArgs([]string{"--batch-id", "497f6eca-6276-4993-bfeb-53cbbbba6f08"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error; got nil")
	}

	expectedError := "error: " + aitaskbuilder.ErrAtLeastOneUpdateFieldRequired
	if err.Error() != expectedError {
		t.Fatalf("expected error: %s; got %s", expectedError, err.Error())
	}
}
