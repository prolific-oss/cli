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

func TestNewGetBatchesCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewGetBatchesListCommand(c, os.Stdout)

	use := "list"
	short := "List batches in a workspace"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewGetBatchesCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := "workspace-123"
	createdAt, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")

	response := client.GetAITaskBuilderBatchesResponse{
		Results: []model.AITaskBuilderBatch{
			{
				ID:                    "batch-1",
				CreatedAt:             createdAt,
				CreatedBy:             "user-123",
				Name:                  "Test Batch 1",
				Status:                "READY",
				TasksPerGroup:         5,
				TotalTaskCount:        100,
				TotalInstructionCount: 10,
				WorkspaceID:           workspaceID,
				Datasets: []model.Dataset{
					{
						ID:                  "dataset-1",
						TotalDatapointCount: 50,
					},
				},
			},
			{
				ID:                    "batch-2",
				CreatedAt:             createdAt,
				CreatedBy:             "user-123",
				Name:                  "Test Batch 2",
				Status:                "PROCESSING",
				TasksPerGroup:         3,
				TotalTaskCount:        75,
				TotalInstructionCount: 8,
				WorkspaceID:           workspaceID,
				Datasets: []model.Dataset{
					{
						ID:                  "dataset-2",
						TotalDatapointCount: 25,
					},
				},
			},
		},
	}

	c.
		EXPECT().
		GetAITaskBuilderBatches(gomock.Eq(workspaceID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewGetBatchesListCommand(c, writer)
	_ = cmd.Flags().Set("workspace-id", workspaceID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Batches by Workspace:
Workspace ID: workspace-123
Batches: 2
  Batch 1: batch-1 | Name: Test Batch 1 | Status: READY
  Batch 2: batch-2 | Name: Test Batch 2 | Status: PROCESSING
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewGetBatchesCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := "invalid-workspace-id"

	c.
		EXPECT().
		GetAITaskBuilderBatches(gomock.Eq(workspaceID)).
		Return(nil, errors.New(workspaceNotFoundError)).
		AnyTimes()

	cmd := aitaskbuilder.NewGetBatchesListCommand(c, os.Stdout)
	_ = cmd.Flags().Set("workspace-id", workspaceID)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", workspaceNotFoundError)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewGetBatchesCommandRequiresWorkspaceID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewGetBatchesListCommand(c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Fatal("expected error when workspace-id is missing")
	}

	if !cmd.Flags().Changed("workspace-id") {
		expected := "workspace ID is required"
		if err.Error() != "error: "+expected {
			t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
		}
	}
}

func TestNewGetBatchesCommandWithNoBatches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := "empty-workspace"

	response := client.GetAITaskBuilderBatchesResponse{
		Results: []model.AITaskBuilderBatch{},
	}

	c.
		EXPECT().
		GetAITaskBuilderBatches(gomock.Eq(workspaceID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewGetBatchesListCommand(c, writer)
	_ = cmd.Flags().Set("workspace-id", workspaceID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Batches by Workspace:
Workspace ID: empty-workspace
Batches: 0
No batches found for workspace empty-workspace
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}
