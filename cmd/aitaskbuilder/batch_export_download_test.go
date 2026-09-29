package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewBatchExportDownloadCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := aitaskbuilder.NewBatchExportDownloadCommand(mockClient, &buf)

	if cmd.Use != "download <batch-id> <export-id>" {
		t.Fatalf("expected use: download <batch-id> <export-id>; got %s", cmd.Use)
	}
}

func TestBatchExportDownloadCommandRequiresArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := aitaskbuilder.NewBatchExportDownloadCommand(mockClient, &buf)

	for _, args := range [][]string{{}, {testBatchID}} {
		if err := cmd.RunE(cmd, args); err == nil {
			t.Fatalf("expected error for args %v, got nil", args)
		}
	}
}

// TestBatchExportDownloadCommandAlreadyComplete covers downloading a job
// that is already complete — no polling required.
func TestBatchExportDownloadCommandAlreadyComplete(t *testing.T) {
	zipContent := []byte("PK\x03\x04fake zip content")
	srv := newBatchZIPServer(t, zipContent)
	defer aitaskbuilder.SetBatchExportDownloadClientForTesting(srv.Client())()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		GetBatchExportStatus(gomock.Eq(testBatchID), gomock.Eq(testBatchExportID)).
		Return(&client.BatchExportResponse{
			Status:    "complete",
			URL:       srv.URL + "/export.zip",
			ExpiresAt: "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	outputPath := filepath.Join(t.TempDir(), "download-complete.zip")

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportDownloadCommand(mockClient, w)
	if err := cmd.Flags().Set("output", outputPath); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{testBatchID, testBatchExportID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}
	if !bytes.Equal(data, zipContent) {
		t.Fatalf("expected file contents to match download, got: %v", data)
	}
}

// TestBatchExportDownloadCommandPollsWhileGenerating covers downloading a
// job that is still generating — the command should poll until complete.
func TestBatchExportDownloadCommandPollsWhileGenerating(t *testing.T) {
	defer aitaskbuilder.SetBatchExportPollSleepForTesting(func(time.Duration) {})()

	zipContent := []byte("PK\x03\x04fake zip content")
	srv := newBatchZIPServer(t, zipContent)
	defer aitaskbuilder.SetBatchExportDownloadClientForTesting(srv.Client())()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		GetBatchExportStatus(gomock.Eq(testBatchID), gomock.Eq(testBatchExportID)).
		Return(&client.BatchExportResponse{Status: "generating"}, nil).
		Times(1)

	gomock.InOrder(
		mockClient.EXPECT().
			GetBatchExportStatus(gomock.Eq(testBatchID), gomock.Eq(testBatchExportID)).
			Return(&client.BatchExportResponse{Status: "generating"}, nil).
			Times(1),
		mockClient.EXPECT().
			GetBatchExportStatus(gomock.Eq(testBatchID), gomock.Eq(testBatchExportID)).
			Return(&client.BatchExportResponse{
				Status:    "complete",
				URL:       srv.URL + "/export.zip",
				ExpiresAt: "2099-01-01T00:00:00Z",
			}, nil).
			Times(1),
	)

	outputPath := filepath.Join(t.TempDir(), "download-polled.zip")

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportDownloadCommand(mockClient, w)
	if err := cmd.Flags().Set("output", outputPath); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{testBatchID, testBatchExportID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("expected output file to exist after polling")
	}
}

func TestBatchExportDownloadCommandFailedStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		GetBatchExportStatus(gomock.Eq(testBatchID), gomock.Eq(testBatchExportID)).
		Return(&client.BatchExportResponse{Status: "failed"}, nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportDownloadCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID, testBatchExportID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error for failed status, got nil")
	}
	if !strings.Contains(err.Error(), "failed to generate") {
		t.Errorf("expected error to explain the job failed, got: %v", err)
	}
	if !strings.Contains(err.Error(), "aitaskbuilder batch export "+testBatchID) {
		t.Errorf("expected error to suggest requesting a new export, got: %v", err)
	}
}

func TestBatchExportDownloadCommandStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		GetBatchExportStatus(gomock.Eq(testBatchID), gomock.Eq(testBatchExportID)).
		Return(nil, errors.New("network error")).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportDownloadCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID, testBatchExportID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error on client failure, got nil")
	}
}

func TestBatchExportDownloadCommandDefaultOutputPath(t *testing.T) {
	zipContent := []byte("PK\x03\x04fake zip content")
	srv := newBatchZIPServer(t, zipContent)
	defer aitaskbuilder.SetBatchExportDownloadClientForTesting(srv.Client())()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		GetBatchExportStatus(gomock.Eq(testBatchID), gomock.Eq(testBatchExportID)).
		Return(&client.BatchExportResponse{
			Status:    "complete",
			URL:       srv.URL + "/export.zip",
			ExpiresAt: "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportDownloadCommand(mockClient, w)

	err = cmd.RunE(cmd, []string{testBatchID, testBatchExportID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	wantPath := testBatchID + "-export-" + testBatchExportID + ".zip"
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("expected default output file %s to exist: %v", wantPath, err)
	}
}
