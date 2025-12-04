package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewBatchTasksCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewBatchTasksCommand(c, os.Stdout)

	use := "tasks"
	short := "Get AI Task Builder batch tasks"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewBatchTasksCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "5cf3ea63-3980-4149-9ea9-bea243489cc8"

	response := client.GetAITaskBuilderTasksResponse{
		"task-123",
		"task-456",
		"task-789",
	}

	c.
		EXPECT().
		GetAITaskBuilderTasks(gomock.Eq(batchID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewBatchTasksCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", batchID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Batch Tasks:
Batch ID: 5cf3ea63-3980-4149-9ea9-bea243489cc8
Total Tasks: 3

Task IDs:
  1. task-123
  2. task-456
  3. task-789
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewBatchTasksCommandHandlesEmptyResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "5d883286-9480-463a-a738-9ddcfae65b8b"

	response := client.GetAITaskBuilderTasksResponse{}

	c.
		EXPECT().
		GetAITaskBuilderTasks(gomock.Eq(batchID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewBatchTasksCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", batchID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Batch Tasks:
Batch ID: 5d883286-9480-463a-a738-9ddcfae65b8b
Total Tasks: 0

No tasks found for batch 5d883286-9480-463a-a738-9ddcfae65b8b
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewBatchTasksCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "big-invalid-batch-id"
	errorMessage := aitaskbuilder.ErrBatchNotFound

	c.
		EXPECT().
		GetAITaskBuilderTasks(gomock.Eq(batchID)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := aitaskbuilder.NewBatchTasksCommand(c, os.Stdout)
	_ = cmd.Flags().Set("batch-id", batchID)
	err := cmd.RunE(cmd, nil)

	expected := errorMessage

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewBatchTasksCommandRequiresBatchID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewBatchTasksCommand(c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Fatal("expected error when batch-id is missing")
	}

	if !cmd.Flags().Changed("batch-id") {
		expected := aitaskbuilder.ErrBatchIDRequired
		if err.Error() != ""+expected {
			t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
		}
	}
}

func TestNewBatchTasksCommandHandlesTasksWithSingleTaskID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "e0d498d3-09cc-4f11-b3ed-99b6753b0a2c"

	response := client.GetAITaskBuilderTasksResponse{
		"single-task-id",
	}

	c.
		EXPECT().
		GetAITaskBuilderTasks(gomock.Eq(batchID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewBatchTasksCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", batchID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Batch Tasks:
Batch ID: e0d498d3-09cc-4f11-b3ed-99b6753b0a2c
Total Tasks: 1

Task IDs:
  1. single-task-id
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}
