package collection_test

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

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/collection"
	"github.com/prolific-oss/cli/mock_client"
)

const testExportID = "export-job-uuid-123"

func TestNewExportCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewExportCommand(mockClient, &buf)

	if cmd.Use != "export <collection-id>" {
		t.Fatalf("expected use: export <collection-id>; got %s", cmd.Use)
	}

	if cmd.Short != "Export a collection's responses as a ZIP archive" {
		t.Fatalf("unexpected short description: %s", cmd.Short)
	}
}

func TestExportCommandRequiresCollectionID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewExportCommand(mockClient, &buf)

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for missing collection ID, got nil")
	}
}

// TestExportCommandImmediateComplete covers the case where POST returns
// status "complete" immediately (server has a valid cached export).
func TestExportCommandImmediateComplete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	zipContent := []byte("PK\x03\x04fake zip content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(zipContent)
	}))
	defer srv.Close()

	mockClient.
		EXPECT().
		InitiateCollectionExport(gomock.Eq(testCollectionID)).
		Return(&client.CollectionExportResponse{
			Status:    "complete",
			URL:       srv.URL + "/export.zip",
			ExpiresAt: "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	outputPath := filepath.Join(t.TempDir(), "immediate-export.zip")

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportCommand(mockClient, w)
	if err := cmd.Flags().Set("output", outputPath); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{testCollectionID})
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

// TestExportCommandPollingToComplete covers the normal async flow:
// POST returns "generating", then GET eventually returns "complete".
func TestExportCommandPollingToComplete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	zipContent := []byte("PK\x03\x04fake zip content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(zipContent)
	}))
	defer srv.Close()

	mockClient.EXPECT().
		InitiateCollectionExport(gomock.Eq(testCollectionID)).
		Return(&client.CollectionExportResponse{
			Status:   "generating",
			ExportID: testExportID,
		}, nil).
		Times(1)

	gomock.InOrder(
		mockClient.EXPECT().
			GetCollectionExportStatus(gomock.Eq(testCollectionID), gomock.Eq(testExportID)).
			Return(&client.CollectionExportResponse{Status: "generating"}, nil).
			Times(1),
		mockClient.EXPECT().
			GetCollectionExportStatus(gomock.Eq(testCollectionID), gomock.Eq(testExportID)).
			Return(&client.CollectionExportResponse{
				Status:    "complete",
				URL:       srv.URL + "/export.zip",
				ExpiresAt: "2099-01-01T00:00:00Z",
			}, nil).
			Times(1),
	)

	outputPath := filepath.Join(t.TempDir(), "poll-export.zip")

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportCommand(mockClient, w)
	if err := cmd.Flags().Set("output", outputPath); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{testCollectionID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("expected output file to exist after polling")
	}
}

func TestExportCommandFailedStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		InitiateCollectionExport(gomock.Eq(testCollectionID)).
		Return(&client.CollectionExportResponse{
			Status:   "generating",
			ExportID: testExportID,
		}, nil).
		Times(1)

	mockClient.EXPECT().
		GetCollectionExportStatus(gomock.Eq(testCollectionID), gomock.Eq(testExportID)).
		Return(&client.CollectionExportResponse{Status: "failed"}, nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testCollectionID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error for failed status, got nil")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("expected error to mention 'failed', got: %v", err)
	}
}

func TestExportCommandInitiateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		InitiateCollectionExport(gomock.Eq(testCollectionID)).
		Return(nil, errors.New("network error")).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testCollectionID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error on client failure, got nil")
	}
}

func TestExportCommandDefaultOutputPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	zipContent := []byte("PK\x03\x04fake zip content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(zipContent)
	}))
	defer srv.Close()

	mockClient.EXPECT().
		InitiateCollectionExport(gomock.Eq(testCollectionID)).
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
	cmd := collection.NewExportCommand(mockClient, w)

	err = cmd.RunE(cmd, []string{testCollectionID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedPath := testCollectionID + "-export.zip"
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected default output file %q to be created", expectedPath)
	}
}
