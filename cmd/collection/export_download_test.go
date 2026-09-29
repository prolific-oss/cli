package collection_test

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
	"github.com/prolific-oss/cli/cmd/collection"
	"github.com/prolific-oss/cli/mock_client"
)

const testCollectionExportID = "export-job-uuid-456"

func TestNewExportDownloadCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewExportDownloadCommand(mockClient, &buf)

	if cmd.Use != "download <collection-id> <export-id>" {
		t.Fatalf("expected use: download <collection-id> <export-id>; got %s", cmd.Use)
	}
}

func TestExportDownloadCommandRequiresArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewExportDownloadCommand(mockClient, &buf)

	for _, args := range [][]string{{}, {testCollectionID}} {
		if err := cmd.RunE(cmd, args); err == nil {
			t.Fatalf("expected error for args %v, got nil", args)
		}
	}
}

// TestExportDownloadCommandAlreadyComplete covers downloading a job that is
// already complete — no polling required.
func TestExportDownloadCommandAlreadyComplete(t *testing.T) {
	zipContent := []byte("PK\x03\x04fake zip content")
	srv := newZIPServer(t, zipContent)
	defer collection.SetDownloadClientForTesting(srv.Client())()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		GetCollectionExportStatus(gomock.Eq(testCollectionID), gomock.Eq(testCollectionExportID)).
		Return(&client.CollectionExportResponse{
			Status:    "complete",
			URL:       srv.URL + "/export.zip",
			ExpiresAt: "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	outputPath := filepath.Join(t.TempDir(), "download-complete.zip")

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportDownloadCommand(mockClient, w)
	if err := cmd.Flags().Set("output", outputPath); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{testCollectionID, testCollectionExportID})
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

// TestExportDownloadCommandPollsWhileGenerating covers downloading a job
// that is still generating — the command should poll until complete.
func TestExportDownloadCommandPollsWhileGenerating(t *testing.T) {
	defer collection.SetPollSleepForTesting(func(time.Duration) {})()

	zipContent := []byte("PK\x03\x04fake zip content")
	srv := newZIPServer(t, zipContent)
	defer collection.SetDownloadClientForTesting(srv.Client())()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	gomock.InOrder(
		mockClient.EXPECT().
			GetCollectionExportStatus(gomock.Eq(testCollectionID), gomock.Eq(testCollectionExportID)).
			Return(&client.CollectionExportResponse{Status: "generating"}, nil).
			Times(1),
		mockClient.EXPECT().
			GetCollectionExportStatus(gomock.Eq(testCollectionID), gomock.Eq(testCollectionExportID)).
			Return(&client.CollectionExportResponse{
				Status:    "complete",
				URL:       srv.URL + "/export.zip",
				ExpiresAt: "2099-01-01T00:00:00Z",
			}, nil).
			Times(1),
	)

	outputPath := filepath.Join(t.TempDir(), "download-polled.zip")

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportDownloadCommand(mockClient, w)
	if err := cmd.Flags().Set("output", outputPath); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{testCollectionID, testCollectionExportID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("expected output file to exist after polling")
	}
}

func TestExportDownloadCommandFailedStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		GetCollectionExportStatus(gomock.Eq(testCollectionID), gomock.Eq(testCollectionExportID)).
		Return(&client.CollectionExportResponse{Status: "failed"}, nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportDownloadCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testCollectionID, testCollectionExportID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error for failed status, got nil")
	}
	if !strings.Contains(err.Error(), "failed to generate") {
		t.Errorf("expected error to explain the job failed, got: %v", err)
	}
	if !strings.Contains(err.Error(), "collection export "+testCollectionID) {
		t.Errorf("expected error to suggest requesting a new export, got: %v", err)
	}
}

func TestExportDownloadCommandStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		GetCollectionExportStatus(gomock.Eq(testCollectionID), gomock.Eq(testCollectionExportID)).
		Return(nil, errors.New("network error")).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportDownloadCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testCollectionID, testCollectionExportID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error on client failure, got nil")
	}
}

func TestExportDownloadCommandFeatureNotEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		GetCollectionExportStatus(gomock.Eq(testCollectionID), gomock.Eq(testCollectionExportID)).
		Return(nil, errors.New("request failed: you do not currently have permission to access this feature")).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportDownloadCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testCollectionID, testCollectionExportID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error for feature-not-enabled, got: %v", err)
	}
}

func TestExportDownloadCommandDefaultOutputPath(t *testing.T) {
	zipContent := []byte("PK\x03\x04fake zip content")
	srv := newZIPServer(t, zipContent)
	defer collection.SetDownloadClientForTesting(srv.Client())()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		GetCollectionExportStatus(gomock.Eq(testCollectionID), gomock.Eq(testCollectionExportID)).
		Return(&client.CollectionExportResponse{
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
	cmd := collection.NewExportDownloadCommand(mockClient, w)

	err = cmd.RunE(cmd, []string{testCollectionID, testCollectionExportID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	wantPath := testCollectionID + "-export-" + testCollectionExportID + ".zip"
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("expected default output file %s to exist: %v", wantPath, err)
	}
}
