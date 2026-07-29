package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
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

const testBatchExportID = "export-job-uuid-456"
const testBatchID = "batch-id-123"

// newBatchZIPServer creates a TLS test server that returns the given bytes as a
// download response. Tests must inject srv.Client() via SetBatchExportDownloadClientForTesting
// so that the HTTPS scheme check passes and the self-signed cert is trusted.
func newBatchZIPServer(t *testing.T, content []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestNewBatchExportCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := aitaskbuilder.NewBatchExportCommand(mockClient, &buf)

	if cmd.Use != "export <batch-id>" {
		t.Fatalf("expected use: export <batch-id>; got %s", cmd.Use)
	}

	if cmd.Short != "Export a batch's responses as a ZIP archive" {
		t.Fatalf("unexpected short description: %s", cmd.Short)
	}
}

func TestBatchExportCommandRequiresBatchID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := aitaskbuilder.NewBatchExportCommand(mockClient, &buf)

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for missing batch ID, got nil")
	}
}

// TestBatchExportCommandImmediateComplete covers the case where POST returns
// status "complete" immediately (server has a valid cached export).
func TestBatchExportCommandImmediateComplete(t *testing.T) {
	zipContent := []byte("PK\x03\x04fake zip content")
	srv := newBatchZIPServer(t, zipContent)
	defer aitaskbuilder.SetBatchExportDownloadClientForTesting(srv.Client())()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.
		EXPECT().
		InitiateBatchExport(gomock.Eq(testBatchID)).
		Return(&client.BatchExportResponse{
			Status:    "complete",
			URL:       srv.URL + "/export.zip",
			ExpiresAt: "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	outputPath := filepath.Join(t.TempDir(), "immediate-export.zip")

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportCommand(mockClient, w)
	if err := cmd.Flags().Set("output", outputPath); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{testBatchID})
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

	output := b.String()
	if !strings.Contains(output, outputPath) {
		t.Errorf("expected output to mention file path %q, got: %s", outputPath, output)
	}
}

// TestBatchExportCommandPollingToComplete covers the normal async flow:
// POST returns "generating", then GET eventually returns "complete".
func TestBatchExportCommandPollingToComplete(t *testing.T) {
	defer aitaskbuilder.SetBatchExportPollSleepForTesting(func(time.Duration) {})()

	zipContent := []byte("PK\x03\x04fake zip content")
	srv := newBatchZIPServer(t, zipContent)
	defer aitaskbuilder.SetBatchExportDownloadClientForTesting(srv.Client())()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		InitiateBatchExport(gomock.Eq(testBatchID)).
		Return(&client.BatchExportResponse{
			Status:   "generating",
			ExportID: testBatchExportID,
		}, nil).
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

	outputPath := filepath.Join(t.TempDir(), "poll-export.zip")

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportCommand(mockClient, w)
	if err := cmd.Flags().Set("output", outputPath); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("expected output file to exist after polling")
	}
}

func TestBatchExportCommandFailedStatus(t *testing.T) {
	defer aitaskbuilder.SetBatchExportPollSleepForTesting(func(time.Duration) {})()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		InitiateBatchExport(gomock.Eq(testBatchID)).
		Return(&client.BatchExportResponse{
			Status:   "generating",
			ExportID: testBatchExportID,
		}, nil).
		Times(1)

	mockClient.EXPECT().
		GetBatchExportStatus(gomock.Eq(testBatchID), gomock.Eq(testBatchExportID)).
		Return(&client.BatchExportResponse{Status: "failed"}, nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error for failed status, got nil")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("expected error to mention 'failed', got: %v", err)
	}
}

func TestBatchExportCommandInitiateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		InitiateBatchExport(gomock.Eq(testBatchID)).
		Return(nil, errors.New("network error")).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error on client failure, got nil")
	}
}

func TestBatchExportCommandDefaultOutputPath(t *testing.T) {
	zipContent := []byte("PK\x03\x04fake zip content")
	srv := newBatchZIPServer(t, zipContent)
	defer aitaskbuilder.SetBatchExportDownloadClientForTesting(srv.Client())()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		InitiateBatchExport(gomock.Eq(testBatchID)).
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
	cmd := aitaskbuilder.NewBatchExportCommand(mockClient, w)

	err = cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// The default filename includes a timestamp, so match by prefix/suffix.
	matches, err := filepath.Glob(testBatchID + "-export-*.zip")
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("expected a default output file matching %s-export-*.zip to be created", testBatchID)
	}
}
