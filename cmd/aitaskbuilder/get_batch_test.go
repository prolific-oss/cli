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

func TestNewGetBatchCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewGetBatchCommand(c, os.Stdout)

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

	batchID := "01954894-65b3-779e-aaf6-348698e23634"

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

	cmd := aitaskbuilder.NewGetBatchCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", batchID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Batch Details:
ID: 01954894-65b3-779e-aaf6-348698e23634
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

	cmd := aitaskbuilder.NewGetBatchCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", batchID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Batch Details:
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
	errorMessage := aitaskbuilder.ErrBatchNotFound

	c.
		EXPECT().
		GetAITaskBuilderBatch(gomock.Eq(batchID)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := aitaskbuilder.NewGetBatchCommand(c, os.Stdout)
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

	cmd := aitaskbuilder.NewGetBatchCommand(c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Fatal("expected error when batch-id is missing")
	}

	if !cmd.Flags().Changed("batch-id") {
		expected := aitaskbuilder.ErrBatchIDRequired
		if err.Error() != "error: "+expected {
			t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
		}
	}
}
