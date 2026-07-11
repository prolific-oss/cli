package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

const testSyncID = "sync-job-uuid-789"

func TestNewBatchSyncCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := aitaskbuilder.NewBatchSyncCommand(mockClient, &buf)

	if cmd.Use != "sync <batch-id>" {
		t.Fatalf("expected use: sync <batch-id>; got %s", cmd.Use)
	}
	if cmd.Short != "Sync a batch with datapoints appended to its dataset" {
		t.Fatalf("unexpected short description: %s", cmd.Short)
	}
}

func TestBatchSyncCommandRequiresBatchID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := aitaskbuilder.NewBatchSyncCommand(mockClient, &buf)

	if err := cmd.RunE(cmd, []string{}); err == nil {
		t.Fatal("expected error for missing batch ID, got nil")
	}
}

// TestBatchSyncCommandPollingToComplete covers the normal async flow: POST
// returns "queued", then GET returns "processing" and finally "complete".
func TestBatchSyncCommandPollingToComplete(t *testing.T) {
	defer aitaskbuilder.SetBatchSyncPollSleepForTesting(func(time.Duration) {})()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		SyncAITaskBuilderBatch(gomock.Eq(testBatchID)).
		Return(&client.AITaskBuilderBatchSyncResponse{Status: "queued", SyncID: testSyncID}, nil).
		Times(1)

	gomock.InOrder(
		mockClient.EXPECT().
			GetAITaskBuilderBatchSyncStatus(gomock.Eq(testBatchID), gomock.Eq(testSyncID)).
			Return(&client.AITaskBuilderBatchSyncResponse{Status: "processing"}, nil).
			Times(1),
		mockClient.EXPECT().
			GetAITaskBuilderBatchSyncStatus(gomock.Eq(testBatchID), gomock.Eq(testSyncID)).
			Return(&client.AITaskBuilderBatchSyncResponse{
				Status:              "complete",
				TasksCreated:        5,
				DatapointsProcessed: 5,
				GroupsCreated:       2,
				GroupsExpanded:      1,
			}, nil).
			Times(1),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchSyncCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := b.String()
	for _, want := range []string{"Sync complete", "Tasks created:", "5", "Groups expanded:", "1"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got: %s", want, output)
		}
	}
}

// TestBatchSyncCommandImmediateComplete covers the POST returning a terminal
// "complete" status straight away, so no polling happens.
func TestBatchSyncCommandImmediateComplete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		SyncAITaskBuilderBatch(gomock.Eq(testBatchID)).
		Return(&client.AITaskBuilderBatchSyncResponse{
			Status:              "complete",
			TasksCreated:        3,
			DatapointsProcessed: 3,
			GroupsCreated:       1,
			GroupsExpanded:      0,
		}, nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchSyncCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.Contains(b.String(), "Sync complete") {
		t.Errorf("expected completion summary, got: %s", b.String())
	}
}

// TestBatchSyncCommandImmediateFailed covers the POST returning a terminal
// "failed" status straight away.
func TestBatchSyncCommandImmediateFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		SyncAITaskBuilderBatch(gomock.Eq(testBatchID)).
		Return(&client.AITaskBuilderBatchSyncResponse{
			Status: "failed",
			Reason: "batch must be in READY status to sync",
		}, nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchSyncCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error for failed status, got nil")
	}
	if !strings.Contains(err.Error(), "batch must be in READY status to sync") {
		t.Errorf("expected error to include the failure reason, got: %v", err)
	}
}

func TestBatchSyncCommandFailedStatus(t *testing.T) {
	defer aitaskbuilder.SetBatchSyncPollSleepForTesting(func(time.Duration) {})()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		SyncAITaskBuilderBatch(gomock.Eq(testBatchID)).
		Return(&client.AITaskBuilderBatchSyncResponse{Status: "queued", SyncID: testSyncID}, nil).
		Times(1)

	mockClient.EXPECT().
		GetAITaskBuilderBatchSyncStatus(gomock.Eq(testBatchID), gomock.Eq(testSyncID)).
		Return(&client.AITaskBuilderBatchSyncResponse{
			Status: "failed",
			Reason: "dataset has an import in progress",
		}, nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchSyncCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error for failed status, got nil")
	}
	if !strings.Contains(err.Error(), "dataset has an import in progress") {
		t.Errorf("expected error to include the failure reason, got: %v", err)
	}
}

func TestBatchSyncCommandInitiateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		SyncAITaskBuilderBatch(gomock.Eq(testBatchID)).
		Return(nil, errors.New("network error")).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchSyncCommand(mockClient, w)

	if err := cmd.RunE(cmd, []string{testBatchID}); err == nil {
		t.Fatal("expected error on client failure, got nil")
	}
}

// TestBatchSyncCommandTransientPollErrorRecovers verifies the poll loop tolerates
// a transient error and continues to a successful terminal state.
func TestBatchSyncCommandTransientPollErrorRecovers(t *testing.T) {
	defer aitaskbuilder.SetBatchSyncPollSleepForTesting(func(time.Duration) {})()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		SyncAITaskBuilderBatch(gomock.Eq(testBatchID)).
		Return(&client.AITaskBuilderBatchSyncResponse{Status: "queued", SyncID: testSyncID}, nil).
		Times(1)

	gomock.InOrder(
		mockClient.EXPECT().
			GetAITaskBuilderBatchSyncStatus(gomock.Eq(testBatchID), gomock.Eq(testSyncID)).
			Return(nil, errors.New("temporary network blip")).
			Times(1),
		mockClient.EXPECT().
			GetAITaskBuilderBatchSyncStatus(gomock.Eq(testBatchID), gomock.Eq(testSyncID)).
			Return(&client.AITaskBuilderBatchSyncResponse{Status: "complete"}, nil).
			Times(1),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchSyncCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected transient error to be tolerated, got: %v", err)
	}
	if !strings.Contains(b.String(), "Sync complete") {
		t.Errorf("expected sync to complete after transient error, got: %s", b.String())
	}
}

func TestBatchSyncCommandTimeout(t *testing.T) {
	defer aitaskbuilder.SetBatchSyncPollSleepForTesting(func(time.Duration) {})()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		SyncAITaskBuilderBatch(gomock.Eq(testBatchID)).
		Return(&client.AITaskBuilderBatchSyncResponse{Status: "queued", SyncID: testSyncID}, nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchSyncCommand(mockClient, w)
	// A zero timeout means the deadline is already in the past by the first poll
	// check, so the loop exits before calling GET.
	if err := cmd.Flags().Set("timeout", "0s"); err != nil {
		t.Fatalf("failed to set timeout flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error, got: %v", err)
	}
}
