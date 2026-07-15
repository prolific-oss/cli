package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

// noOpBrowserOpener is a no-op browser opener for testing.
func noOpBrowserOpener(url string) error {
	return nil
}

const testTaskGroupID = "task-group-1"

func TestNewBatchPreviewCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewBatchPreviewCommandWithOpener(c, os.Stdout, noOpBrowserOpener)

	use := "preview <batch-id>"
	short := "Preview a batch in the browser"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestBatchPreviewRequiresBatchID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewBatchPreviewCommandWithOpener(c, os.Stdout, noOpBrowserOpener)
	err := cmd.Args(cmd, []string{})

	if err == nil {
		t.Fatal("expected error when batch-id is missing")
	}
}

func TestBatchPreviewAcceptsPositionalBatchID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := testBatchUUID
	taskGroupID := testTaskGroupID

	response := client.GetAITaskBuilderBatchResponse{
		AITaskBuilderBatch: model.AITaskBuilderBatch{ID: batchID},
	}
	taskGroups := client.GetAITaskBuilderTaskGroupsResponse{taskGroupID}

	c.EXPECT().GetAITaskBuilderBatch(gomock.Eq(batchID)).Return(&response, nil).Times(1)
	c.EXPECT().GetAITaskBuilderTaskGroups(gomock.Eq(batchID)).Return(&taskGroups, nil).Times(1)

	cmd := aitaskbuilder.NewBatchPreviewCommandWithOpener(c, os.Stdout, noOpBrowserOpener)
	if err := cmd.RunE(cmd, []string{batchID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBatchPreviewCallsGetBatchAndTaskGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := testBatchUUID
	taskGroupID := testTaskGroupID

	response := client.GetAITaskBuilderBatchResponse{
		AITaskBuilderBatch: model.AITaskBuilderBatch{ID: batchID},
	}
	taskGroups := client.GetAITaskBuilderTaskGroupsResponse{taskGroupID}

	c.EXPECT().GetAITaskBuilderBatch(gomock.Eq(batchID)).Return(&response, nil).Times(1)
	c.EXPECT().GetAITaskBuilderTaskGroups(gomock.Eq(batchID)).Return(&taskGroups, nil).Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewBatchPreviewCommandWithOpener(c, writer, noOpBrowserOpener)
	err := cmd.RunE(cmd, []string{batchID})
	writer.Flush()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestBatchPreviewReturnsErrorOnBatchClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "an-invalid-batch-id"

	c.EXPECT().GetAITaskBuilderBatch(gomock.Eq(batchID)).Return(nil, errors.New("batch not found")).Times(1)

	cmd := aitaskbuilder.NewBatchPreviewCommandWithOpener(c, os.Stdout, noOpBrowserOpener)
	err := cmd.RunE(cmd, []string{batchID})

	expected := "error: failed to get batch: batch not found"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected error %q, got %v", expected, err)
	}
}

func TestBatchPreviewReturnsErrorWhenNoTaskGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := testBatchUUID

	response := client.GetAITaskBuilderBatchResponse{
		AITaskBuilderBatch: model.AITaskBuilderBatch{ID: batchID},
	}
	taskGroups := client.GetAITaskBuilderTaskGroupsResponse{}

	c.EXPECT().GetAITaskBuilderBatch(gomock.Eq(batchID)).Return(&response, nil).Times(1)
	c.EXPECT().GetAITaskBuilderTaskGroups(gomock.Eq(batchID)).Return(&taskGroups, nil).Times(1)

	cmd := aitaskbuilder.NewBatchPreviewCommandWithOpener(c, os.Stdout, noOpBrowserOpener)
	err := cmd.RunE(cmd, []string{batchID})

	expected := fmt.Sprintf("error: %s %s", aitaskbuilder.ErrNoTaskGroupsFound, batchID)
	if err == nil || err.Error() != expected {
		t.Fatalf("expected error %q, got %v", expected, err)
	}
}

func TestBatchPreviewOutputContainsURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := testBatchUUID
	taskGroupID := testTaskGroupID

	response := client.GetAITaskBuilderBatchResponse{
		AITaskBuilderBatch: model.AITaskBuilderBatch{ID: batchID},
	}
	taskGroups := client.GetAITaskBuilderTaskGroupsResponse{taskGroupID}

	c.EXPECT().GetAITaskBuilderBatch(gomock.Eq(batchID)).Return(&response, nil).Times(1)
	c.EXPECT().GetAITaskBuilderTaskGroups(gomock.Eq(batchID)).Return(&taskGroups, nil).Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewBatchPreviewCommandWithOpener(c, writer, noOpBrowserOpener)
	err := cmd.RunE(cmd, []string{batchID})
	writer.Flush()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := b.String()
	expectedStrings := []string{
		"Opening batch preview in browser",
		"data-collection-tool/batches/" + batchID + "/task-groups/" + taskGroupID,
		"preview=true",
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("expected output to contain %q, got: %s", expected, output)
		}
	}
}

func TestGetBatchPreviewPath(t *testing.T) {
	path := aitaskbuilder.GetBatchPreviewPath("batch-123", "task-group-456")

	expected := "data-collection-tool/batches/batch-123/task-groups/task-group-456?preview=true"
	if path != expected {
		t.Fatalf("expected path %q, got %q", expected, path)
	}
}

func TestGetBatchPreviewURL(t *testing.T) {
	url := aitaskbuilder.GetBatchPreviewURL("batch-123", "task-group-456")

	expected := "https://app.prolific.com/data-collection-tool/batches/batch-123/task-groups/task-group-456?preview=true"
	if url != expected {
		t.Fatalf("expected URL %q, got %q", expected, url)
	}
}
