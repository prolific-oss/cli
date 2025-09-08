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

func TestNewGetCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewGetCommand(c, os.Stdout)

	use := "get"
	short := "Get an AI task builder batch"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewGetCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23608"
	name := "Test Batch"
	workspaceID := "6745ab669112d10b9b3afb48"

	createdAt, _ := time.Parse(time.RFC3339, "2025-02-27T18:03:59.795Z")
	response := client.GetAITaskBuilderBatchResponse{
		AITaskBuilderBatch: model.AITaskBuilderBatch{
			ID:                    batchID,
			CreatedAt:             createdAt,
			CreatedBy:             "6139f0d1dc08858054c63b2c",
			Name:                  name,
			Status:                "UNINITIALISED",
			TotalTaskCount:        0,
			TotalInstructionCount: 5,
			WorkspaceID:           workspaceID,
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
		GetAITaskBuilderBatch(gomock.Eq(batchID), gomock.Eq(name), gomock.Eq(workspaceID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewGetCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", batchID)
	_ = cmd.Flags().Set("name", name)
	_ = cmd.Flags().Set("workspace", workspaceID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Batch Details:
ID: 01954894-65b3-779e-aaf6-348698e23608
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

func TestNewGetCommandCallsAPIWithoutOptionalFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23608"
	name := "Simple Batch"
	workspaceID := "6745ab669112d10b9b3afb48"

	createdAt, _ := time.Parse(time.RFC3339, "2025-02-27T18:03:59.795Z")
	response := client.GetAITaskBuilderBatchResponse{
		AITaskBuilderBatch: model.AITaskBuilderBatch{
			ID:                    batchID,
			CreatedAt:             createdAt,
			CreatedBy:             "6139f0d1dc08858054c63b2c",
			Name:                  name,
			Status:                "ACTIVE",
			TotalTaskCount:        5,
			TotalInstructionCount: 0,
			WorkspaceID:           workspaceID,
			SchemaVersion:         1,
			Datasets:              []model.Dataset{},   // Empty datasets
			TaskDetails:           model.TaskDetails{}, // Empty task details
		},
	}

	c.
		EXPECT().
		GetAITaskBuilderBatch(gomock.Eq(batchID), gomock.Eq(name), gomock.Eq(workspaceID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewGetCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", batchID)
	_ = cmd.Flags().Set("name", name)
	_ = cmd.Flags().Set("workspace", workspaceID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Batch Details:
ID: 01954894-65b3-779e-aaf6-348698e23608
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

func TestNewGetCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "invalid-batch-id"
	name := "Test Batch"
	workspaceID := "invalid-workspace"
	errorMessage := "batch not found"

	c.
		EXPECT().
		GetAITaskBuilderBatch(gomock.Eq(batchID), gomock.Eq(name), gomock.Eq(workspaceID)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := aitaskbuilder.NewGetCommand(c, os.Stdout)
	_ = cmd.Flags().Set("batch-id", batchID)
	_ = cmd.Flags().Set("name", name)
	_ = cmd.Flags().Set("workspace", workspaceID)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewGetCommandRequiresBatchID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewGetCommand(c, os.Stdout)
	_ = cmd.Flags().Set("name", "Test Batch")
	_ = cmd.Flags().Set("workspace", "workspace-id")
	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Fatal("expected error when batch-id is missing")
	}

	if !cmd.Flags().Changed("batch-id") {
		expected := "batch ID is required"
		if err.Error() != "error: "+expected {
			t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
		}
	}
}

func TestNewGetCommandRequiresName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewGetCommand(c, os.Stdout)
	_ = cmd.Flags().Set("batch-id", "batch-id")
	_ = cmd.Flags().Set("workspace", "workspace-id")
	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Fatal("expected error when name is missing")
	}

	if !cmd.Flags().Changed("name") {
		expected := "batch name is required"
		if err.Error() != "error: "+expected {
			t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
		}
	}
}

func TestNewGetCommandRequiresWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewGetCommand(c, os.Stdout)
	_ = cmd.Flags().Set("batch-id", "batch-id")
	_ = cmd.Flags().Set("name", "Test Batch")
	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Fatal("expected error when workspace is missing")
	}

	if !cmd.Flags().Changed("workspace") {
		expected := "workspace ID is required"
		if err.Error() != "error: "+expected {
			t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
		}
	}
}
