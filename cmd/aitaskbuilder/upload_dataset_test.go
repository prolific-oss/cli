package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
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
	"github.com/prolific-oss/cli/model"
)

func setDatasetUploadPollSleepForTesting(f func(time.Duration)) func() {
	prev := aitaskbuilder.DatasetUploadPollSleep
	aitaskbuilder.DatasetUploadPollSleep = f
	return func() { aitaskbuilder.DatasetUploadPollSleep = prev }
}

func TestNewDatasetUploadCommand(t *testing.T) {
	c := setupMockClient(t)

	cmd := aitaskbuilder.NewDatasetUploadCommand(c, os.Stdout)

	if cmd.Use != "upload" {
		t.Fatalf("expected use: upload; got %s", cmd.Use)
	}

	if cmd.Short != "Upload data to a dataset" {
		t.Fatalf("expected short: Upload data to a dataset; got %s", cmd.Short)
	}

	if !cmd.SilenceErrors {
		t.Fatal("expected upload command to silence cobra error output")
	}

	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("expected format flag to be registered")
	}

	timeoutFlag := cmd.Flags().Lookup("timeout")
	if timeoutFlag == nil {
		t.Fatal("expected timeout flag to be registered")
	}

	if !strings.Contains(cmd.Long, "auto-detects CSV and JSONL formats") {
		t.Fatalf("expected long description to mention format detection, got %s", cmd.Long)
	}

	if !strings.Contains(cmd.Long, "requires a schema") {
		t.Fatalf("expected long description to mention schema outcome, got %s", cmd.Long)
	}

	if !strings.Contains(cmd.Example, "docs/examples/aitb-model-evaluation.csv") {
		t.Fatalf("expected example to include CSV upload, got %s", cmd.Example)
	}

	if !strings.Contains(cmd.Example, "--format jsonl") {
		t.Fatalf("expected example to include explicit jsonl format, got %s", cmd.Example)
	}
}

func TestDatasetUploadCommandUploadsDetectedCSV(t *testing.T) {
	defer setDatasetUploadPollSleepForTesting(func(time.Duration) {})()

	fileContent := []byte("name,score\nalice,10\n")
	filePath := filepath.Join(t.TempDir(), "dataset.csv")
	if err := os.WriteFile(filePath, fileContent, 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var gotMethod string
	var gotContentType string
	var gotContentLength int64
	var gotTransferEncoding []string
	var gotBody []byte
	var gotUserAgent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read upload body: %v", err)
		}

		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		gotContentLength = r.ContentLength
		gotTransferEncoding = r.TransferEncoding
		gotBody = body
		gotUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := setupMockClient(t)
	c.EXPECT().
		GetAITaskBuilderDatasetUploadURL(
			gomock.Eq("dataset-123"),
			gomock.Eq("dataset.csv"),
		).
		Return(&client.GetAITaskBuilderDatasetUploadURLResponse{
			ImportID:    "import-123",
			UploadURL:   srv.URL,
			HTTPMethod:  http.MethodPut,
			ContentType: "text/csv",
			ExpiresAt:   "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	acceptedCount := 3
	rejectedCount := 0
	c.EXPECT().
		GetAITaskBuilderDatasetImportStatus(
			gomock.Eq("dataset-123"),
			gomock.Eq("import-123"),
		).
		Return(&client.GetAITaskBuilderDatasetImportStatusResponse{
			DatasetImportJob: model.DatasetImportJob{
				DatasetID:     "dataset-123",
				ImportID:      "import-123",
				Status:        model.DatasetImportJobStatusComplete,
				AcceptedCount: &acceptedCount,
				RejectedCount: &rejectedCount,
			},
		}, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewDatasetUploadCommand(c, writer)
	_ = cmd.Flags().Set("dataset-id", "dataset-123")
	_ = cmd.Flags().Set("file", filePath)

	err := cmd.RunE(cmd, nil)
	writer.Flush()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Fatalf("expected upload method %s, got %s", http.MethodPut, gotMethod)
	}

	if gotContentType != "text/csv" {
		t.Fatalf("expected content type text/csv, got %s", gotContentType)
	}

	if gotContentLength != int64(len(fileContent)) {
		t.Fatalf("expected content length %d, got %d", len(fileContent), gotContentLength)
	}

	if len(gotTransferEncoding) != 0 {
		t.Fatalf("expected no transfer encoding, got %v", gotTransferEncoding)
	}

	if !bytes.Equal(gotBody, fileContent) {
		t.Fatalf("expected uploaded file contents to match")
	}

	if gotUserAgent != testUserAgent {
		t.Fatalf("expected User-Agent %q, got %q", testUserAgent, gotUserAgent)
	}

	output := b.String()
	if !strings.Contains(output, "Import ID: import-123") {
		t.Fatalf("expected output to contain import ID, got %s", output)
	}

	if !strings.Contains(output, "Import complete.") {
		t.Fatalf("expected success summary, got %s", output)
	}

	if !strings.Contains(output, "Format: csv") || !strings.Contains(output, "Imported: 3") {
		t.Fatalf("expected import summary details, got %s", output)
	}
}

func TestDatasetUploadCommandFormatOverrideAppendsExtension(t *testing.T) {
	defer setDatasetUploadPollSleepForTesting(func(time.Duration) {})()

	fileContent := []byte("{\"name\":\"alice\"}\n")
	filePath := filepath.Join(t.TempDir(), "records.data")
	if err := os.WriteFile(filePath, fileContent, 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := setupMockClient(t)
	c.EXPECT().
		GetAITaskBuilderDatasetUploadURL(
			gomock.Eq("dataset-456"),
			gomock.Eq("records.data.jsonl"),
		).
		Return(&client.GetAITaskBuilderDatasetUploadURLResponse{
			ImportID:    "import-456",
			UploadURL:   srv.URL,
			HTTPMethod:  http.MethodPut,
			ContentType: "application/x-ndjson",
			ExpiresAt:   "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	acceptedCount := 1
	rejectedCount := 0
	c.EXPECT().
		GetAITaskBuilderDatasetImportStatus(
			gomock.Eq("dataset-456"),
			gomock.Eq("import-456"),
		).
		Return(&client.GetAITaskBuilderDatasetImportStatusResponse{
			DatasetImportJob: model.DatasetImportJob{
				DatasetID:     "dataset-456",
				ImportID:      "import-456",
				Status:        model.DatasetImportJobStatusComplete,
				AcceptedCount: &acceptedCount,
				RejectedCount: &rejectedCount,
			},
		}, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewDatasetUploadCommand(c, writer)
	_ = cmd.Flags().Set("dataset-id", "dataset-456")
	_ = cmd.Flags().Set("file", filePath)
	_ = cmd.Flags().Set("format", "jsonl")

	err := cmd.RunE(cmd, nil)
	writer.Flush()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if gotContentType != "application/x-ndjson" {
		t.Fatalf("expected content type application/x-ndjson, got %s", gotContentType)
	}

	if !strings.Contains(b.String(), "Format: jsonl") {
		t.Fatalf("expected output to contain jsonl format, got %s", b.String())
	}
}

func TestDatasetUploadCommandRejectsUnsupportedCSVMediaExtension(t *testing.T) {
	tests := []struct {
		name                string
		fieldType           string
		datasetID           string
		unsupportedURL      string
		supportedExtensions string
	}{
		{
			name:                "audio",
			fieldType:           "audio_url",
			datasetID:           "dataset-audio",
			unsupportedURL:      "https://example.com/audio.txt",
			supportedExtensions: ".aac, .m4a, .mp3, .wav",
		},
		{
			name:                "video",
			fieldType:           "video_url",
			datasetID:           "dataset-video",
			unsupportedURL:      "https://example.com/video.txt",
			supportedExtensions: ".mp4, .webm, .mov",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(t.TempDir(), "dataset.csv")
			csvContents := fmt.Sprintf("question,clip\nhello,%s\n", tt.unsupportedURL)
			if err := os.WriteFile(filePath, []byte(csvContents), 0o600); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)
			c.EXPECT().
				GetAITaskBuilderDataset(gomock.Eq(tt.datasetID)).
				Return(&client.GetAITaskBuilderDatasetResponse{
					Schema: &client.DatasetSchema{
						Fields: map[string]client.DatasetSchemaField{
							"clip": {Type: tt.fieldType},
						},
					},
				}, nil).
				Times(1)

			cmd := aitaskbuilder.NewDatasetUploadCommand(c, os.Stdout)
			_ = cmd.Flags().Set("dataset-id", tt.datasetID)
			_ = cmd.Flags().Set("file", filePath)

			err := cmd.RunE(cmd, nil)
			if err == nil {
				t.Fatal("expected invalid media URL extension error")
			}

			if !strings.Contains(err.Error(), fmt.Sprintf("must end with one of %s", tt.supportedExtensions)) {
				t.Fatalf("expected supported extensions in error, got %v", err)
			}
		})
	}
}

func TestValidateMediaURLFieldsInJSONLRejectsUnsupportedExtension(t *testing.T) {
	tests := []struct {
		name           string
		validate       func(string, map[string]struct{}) error
		unsupportedURL string
	}{
		{
			name:           "audio",
			validate:       aitaskbuilder.ValidateAudioURLFieldsInJSONL,
			unsupportedURL: "https://example.com/audio.mov",
		},
		{
			name:           "video",
			validate:       aitaskbuilder.ValidateVideoURLFieldsInJSONL,
			unsupportedURL: "https://example.com/video.mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(t.TempDir(), "dataset.jsonl")
			content := fmt.Sprintf("{\"clip\":%q}\n", tt.unsupportedURL)
			if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			mediaFields := map[string]struct{}{
				"clip": {},
			}

			err := tt.validate(filePath, mediaFields)
			if err == nil {
				t.Fatal("expected invalid media URL extension error")
			}

			if !strings.Contains(err.Error(), "record 1 field clip") {
				t.Fatalf("expected record location in error, got %v", err)
			}
		})
	}
}

func TestDatasetUploadCommandRendersPartialImportSummary(t *testing.T) {
	defer setDatasetUploadPollSleepForTesting(func(time.Duration) {})()

	filePath := filepath.Join(t.TempDir(), "dataset.csv")
	if err := os.WriteFile(filePath, []byte("name,score\nalice,10\n"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := setupMockClient(t)
	c.EXPECT().
		GetAITaskBuilderDatasetUploadURL(
			gomock.Eq("dataset-partial"),
			gomock.Eq("dataset.csv"),
		).
		Return(&client.GetAITaskBuilderDatasetUploadURLResponse{
			ImportID:    "import-partial",
			UploadURL:   srv.URL,
			HTTPMethod:  http.MethodPut,
			ContentType: "text/csv",
			ExpiresAt:   "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	acceptedCount := 4
	duplicateCount := 2
	rejectedCount := 25
	c.EXPECT().
		GetAITaskBuilderDatasetImportStatus(
			gomock.Eq("dataset-partial"),
			gomock.Eq("import-partial"),
		).
		Return(&client.GetAITaskBuilderDatasetImportStatusResponse{
			DatasetImportJob: model.DatasetImportJob{
				DatasetID:      "dataset-partial",
				ImportID:       "import-partial",
				Status:         model.DatasetImportJobStatusPartial,
				AcceptedCount:  &acceptedCount,
				DuplicateCount: &duplicateCount,
				RejectedCount:  &rejectedCount,
				Errors: []model.DatasetImportJobRecordError{
					{RecordIndex: 0, Field: "email", Reason: "invalid email"},
					{RecordIndex: 1, Field: "name", Reason: "missing value"},
				},
				Reason: "Some rows were rejected.",
			},
		}, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewDatasetUploadCommand(c, writer)
	_ = cmd.Flags().Set("dataset-id", "dataset-partial")
	_ = cmd.Flags().Set("file", filePath)

	err := cmd.RunE(cmd, nil)
	writer.Flush()
	if err != nil {
		t.Fatalf("expected no error for partial import, got %v", err)
	}

	output := b.String()
	for _, expected := range []string{
		"Import completed with warnings.",
		"Accepted: 4",
		"Newly Imported: 2",
		"Duplicates: 2",
		"Rejected: 25",
		"record 1 field email: invalid email",
		"record 2 field name: missing value",
		"Detailed errors are a partial view: 25 records were rejected but only 2 detailed errors were returned.",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q, got %s", expected, output)
		}
	}
}

func TestDatasetUploadCommandRendersDuplicateOnlySuccess(t *testing.T) {
	defer setDatasetUploadPollSleepForTesting(func(time.Duration) {})()

	filePath := filepath.Join(t.TempDir(), "dataset.csv")
	if err := os.WriteFile(filePath, []byte("name,score\nalice,10\n"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := setupMockClient(t)
	c.EXPECT().
		GetAITaskBuilderDatasetUploadURL(
			gomock.Eq("dataset-duplicates"),
			gomock.Eq("dataset.csv"),
		).
		Return(&client.GetAITaskBuilderDatasetUploadURLResponse{
			ImportID:    "import-duplicates",
			UploadURL:   srv.URL,
			HTTPMethod:  http.MethodPut,
			ContentType: "text/csv",
			ExpiresAt:   "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	acceptedCount := 3
	duplicateCount := 3
	rejectedCount := 0
	c.EXPECT().
		GetAITaskBuilderDatasetImportStatus(
			gomock.Eq("dataset-duplicates"),
			gomock.Eq("import-duplicates"),
		).
		Return(&client.GetAITaskBuilderDatasetImportStatusResponse{
			DatasetImportJob: model.DatasetImportJob{
				DatasetID:      "dataset-duplicates",
				ImportID:       "import-duplicates",
				Status:         model.DatasetImportJobStatusComplete,
				AcceptedCount:  &acceptedCount,
				DuplicateCount: &duplicateCount,
				RejectedCount:  &rejectedCount,
			},
		}, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewDatasetUploadCommand(c, writer)
	_ = cmd.Flags().Set("dataset-id", "dataset-duplicates")
	_ = cmd.Flags().Set("file", filePath)

	err := cmd.RunE(cmd, nil)
	writer.Flush()
	if err != nil {
		t.Fatalf("expected no error for duplicate-only import, got %v", err)
	}

	output := b.String()
	for _, expected := range []string{
		"Import complete.",
		"Accepted: 3",
		"Newly Imported: 0",
		"Duplicates: 3",
		"Note: all accepted records were duplicates; no new records were imported.",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q, got %s", expected, output)
		}
	}
}

func TestDatasetUploadCommandRejectsUndeterminedFormat(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "records")
	if err := os.WriteFile(filePath, []byte("payload"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	c := setupMockClient(t)

	cmd := aitaskbuilder.NewDatasetUploadCommand(c, os.Stdout)
	_ = cmd.Flags().Set("dataset-id", "dataset-789")
	_ = cmd.Flags().Set("file", filePath)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error for missing file format")
	}

	if !strings.Contains(err.Error(), aitaskbuilder.ErrDatasetUploadFormatRequired) {
		t.Fatalf("expected format guidance error, got %v", err)
	}
}

func TestDatasetUploadCommandRejectsEmptyFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "empty.csv")
	if err := os.WriteFile(filePath, nil, 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	c := setupMockClient(t)

	cmd := aitaskbuilder.NewDatasetUploadCommand(c, os.Stdout)
	_ = cmd.Flags().Set("dataset-id", "dataset-999")
	_ = cmd.Flags().Set("file", filePath)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error for empty file")
	}

	if !strings.Contains(err.Error(), aitaskbuilder.ErrDatasetUploadFileEmpty) {
		t.Fatalf("expected empty-file error, got %v", err)
	}
}

func TestDatasetUploadCommandDoesNotPrintUsageForRuntimeError(t *testing.T) {
	defer setDatasetUploadPollSleepForTesting(func(time.Duration) {})()

	filePath := filepath.Join(t.TempDir(), "dataset.csv")
	if err := os.WriteFile(filePath, []byte("name,score\nalice,10\n"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := setupMockClient(t)
	c.EXPECT().
		GetAITaskBuilderDatasetUploadURL(
			gomock.Eq("dataset-schema"),
			gomock.Eq("dataset.csv"),
		).
		Return(&client.GetAITaskBuilderDatasetUploadURLResponse{
			ImportID:    "import-schema",
			UploadURL:   srv.URL,
			HTTPMethod:  http.MethodPut,
			ContentType: "text/csv",
			ExpiresAt:   "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	c.EXPECT().
		GetAITaskBuilderDatasetImportStatus(
			gomock.Eq("dataset-schema"),
			gomock.Eq("import-schema"),
		).
		Return(&client.GetAITaskBuilderDatasetImportStatusResponse{
			DatasetImportJob: model.DatasetImportJob{
				DatasetID: "dataset-schema",
				ImportID:  "import-schema",
				Status:    model.DatasetImportJobStatusFailed,
				Reason:    "processing failed",
			},
		}, nil).
		Times(1)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	writer := bufio.NewWriter(&stdout)
	cmd := aitaskbuilder.NewDatasetUploadCommand(c, writer)
	cmd.SetOut(io.Discard)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"-d", "dataset-schema", "-f", filePath})

	err := cmd.Execute()
	writer.Flush()
	if err == nil {
		t.Fatal("expected runtime error")
	}

	if strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("expected no usage output for runtime error, got %s", stderr.String())
	}

	if strings.Contains(stderr.String(), "waiting for a dataset schema") {
		t.Fatalf("expected no cobra error output for runtime error, got %s", stderr.String())
	}

	if !strings.Contains(err.Error(), "dataset import failed") {
		t.Fatalf("expected failed import error, got %v", err)
	}
}

func TestDatasetUploadCommandTimesOutWhilePolling(t *testing.T) {
	defer setDatasetUploadPollSleepForTesting(func(time.Duration) {})()

	filePath := filepath.Join(t.TempDir(), "dataset.csv")
	if err := os.WriteFile(filePath, []byte("name\nalice\n"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := setupMockClient(t)
	c.EXPECT().
		GetAITaskBuilderDatasetUploadURL(
			gomock.Eq("dataset-timeout"),
			gomock.Eq("dataset.csv"),
		).
		Return(&client.GetAITaskBuilderDatasetUploadURLResponse{
			ImportID:    "import-timeout",
			UploadURL:   srv.URL,
			HTTPMethod:  http.MethodPut,
			ContentType: "text/csv",
			ExpiresAt:   "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	c.EXPECT().
		GetAITaskBuilderDatasetImportStatus(
			gomock.Eq("dataset-timeout"),
			gomock.Eq("import-timeout"),
		).
		Return(&client.GetAITaskBuilderDatasetImportStatusResponse{
			DatasetImportJob: model.DatasetImportJob{
				Status: model.DatasetImportJobStatusProcessing,
			},
		}, nil).
		MinTimes(1)

	cmd := aitaskbuilder.NewDatasetUploadCommand(c, os.Stdout)
	_ = cmd.Flags().Set("dataset-id", "dataset-timeout")
	_ = cmd.Flags().Set("file", filePath)
	_ = cmd.Flags().Set("timeout", "1ms")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !strings.Contains(err.Error(), "processing may still continue server-side") {
		t.Fatalf("expected timeout guidance, got %v", err)
	}
}

func TestDatasetUploadCommandFailsAfterConsecutivePollingErrors(t *testing.T) {
	defer setDatasetUploadPollSleepForTesting(func(time.Duration) {})()

	filePath := filepath.Join(t.TempDir(), "dataset.csv")
	if err := os.WriteFile(filePath, []byte("name\nalice\n"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := setupMockClient(t)
	c.EXPECT().
		GetAITaskBuilderDatasetUploadURL(
			gomock.Eq("dataset-errors"),
			gomock.Eq("dataset.csv"),
		).
		Return(&client.GetAITaskBuilderDatasetUploadURLResponse{
			ImportID:    "import-errors",
			UploadURL:   srv.URL,
			HTTPMethod:  http.MethodPut,
			ContentType: "text/csv",
			ExpiresAt:   "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	c.EXPECT().
		GetAITaskBuilderDatasetImportStatus(
			gomock.Eq("dataset-errors"),
			gomock.Eq("import-errors"),
		).
		Return(nil, errors.New("temporary status failure")).
		Times(3)

	cmd := aitaskbuilder.NewDatasetUploadCommand(c, os.Stdout)
	_ = cmd.Flags().Set("dataset-id", "dataset-errors")
	_ = cmd.Flags().Set("file", filePath)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected polling error")
	}

	if !strings.Contains(err.Error(), "after 3 consecutive attempts") {
		t.Fatalf("expected consecutive polling error, got %v", err)
	}
}

func TestDatasetUploadCommandWarnsForPendingSchema(t *testing.T) {
	defer setDatasetUploadPollSleepForTesting(func(time.Duration) {})()

	filePath := filepath.Join(t.TempDir(), "dataset.csv")
	if err := os.WriteFile(filePath, []byte("name\nalice\n"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := setupMockClient(t)
	c.EXPECT().
		GetAITaskBuilderDatasetUploadURL(
			gomock.Eq("dataset-schema"),
			gomock.Eq("dataset.csv"),
		).
		Return(&client.GetAITaskBuilderDatasetUploadURLResponse{
			ImportID:    "import-schema",
			UploadURL:   srv.URL,
			HTTPMethod:  http.MethodPut,
			ContentType: "text/csv",
			ExpiresAt:   "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	c.EXPECT().
		GetAITaskBuilderDatasetImportStatus(
			gomock.Eq("dataset-schema"),
			gomock.Eq("import-schema"),
		).
		Return(&client.GetAITaskBuilderDatasetImportStatusResponse{
			DatasetImportJob: model.DatasetImportJob{
				Status: model.DatasetImportJobStatusPendingSchema,
			},
		}, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewDatasetUploadCommand(c, writer)
	_ = cmd.Flags().Set("dataset-id", "dataset-schema")
	_ = cmd.Flags().Set("file", filePath)

	err := cmd.RunE(cmd, nil)
	writer.Flush()
	if err != nil {
		t.Fatalf("expected pending-schema warning, got %v", err)
	}

	output := b.String()
	if !strings.Contains(output, "Warning: upload was received for dataset dataset-schema import import-schema") {
		t.Fatalf("expected pending-schema warning output, got %s", output)
	}

	if !strings.Contains(output, "define a schema for the dataset rather than re-uploading the file") {
		t.Fatalf("expected pending-schema guidance, got %s", output)
	}
}

func TestDatasetUploadCommandReturnsFailedImportError(t *testing.T) {
	defer setDatasetUploadPollSleepForTesting(func(time.Duration) {})()

	filePath := filepath.Join(t.TempDir(), "dataset.csv")
	if err := os.WriteFile(filePath, []byte("name\nalice\n"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := setupMockClient(t)
	c.EXPECT().
		GetAITaskBuilderDatasetUploadURL(
			gomock.Eq("dataset-failed"),
			gomock.Eq("dataset.csv"),
		).
		Return(&client.GetAITaskBuilderDatasetUploadURLResponse{
			ImportID:    "import-failed",
			UploadURL:   srv.URL,
			HTTPMethod:  http.MethodPut,
			ContentType: "text/csv",
			ExpiresAt:   "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	rejectedCount := 25
	recordErrors := make([]model.DatasetImportJobRecordError, 0, 21)
	for i := 0; i < 21; i++ {
		recordErrors = append(recordErrors, model.DatasetImportJobRecordError{
			RecordIndex: i,
			Field:       "email",
			Reason:      "invalid email",
		})
	}

	c.EXPECT().
		GetAITaskBuilderDatasetImportStatus(
			gomock.Eq("dataset-failed"),
			gomock.Eq("import-failed"),
		).
		Return(&client.GetAITaskBuilderDatasetImportStatusResponse{
			DatasetImportJob: model.DatasetImportJob{
				DatasetID:     "dataset-failed",
				ImportID:      "import-failed",
				Status:        model.DatasetImportJobStatusFailed,
				RejectedCount: &rejectedCount,
				Errors:        recordErrors,
				Reason:        "Schema validation failed.",
			},
		}, nil).
		Times(1)

	cmd := aitaskbuilder.NewDatasetUploadCommand(c, os.Stdout)
	_ = cmd.Flags().Set("dataset-id", "dataset-failed")
	_ = cmd.Flags().Set("file", filePath)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected failed import error")
	}

	errorText := err.Error()
	for _, expected := range []string{
		"dataset import failed for dataset dataset-failed import import-failed: Schema validation failed.",
		"Record errors:",
		"record 1 field email: invalid email",
		"record 20 field email: invalid email",
		"Detailed errors are a partial view: 25 records were rejected but only 20 detailed errors were returned.",
	} {
		if !strings.Contains(errorText, expected) {
			t.Fatalf("expected error to contain %q, got %s", expected, errorText)
		}
	}

	if strings.Contains(errorText, "record 21 field email: invalid email") {
		t.Fatalf("expected error output to stop at 20 record errors, got %s", errorText)
	}
}

func TestDatasetUploadCommandRejectsNonPositiveTimeout(t *testing.T) {
	c := setupMockClient(t)

	cmd := aitaskbuilder.NewDatasetUploadCommand(c, os.Stdout)
	_ = cmd.Flags().Set("dataset-id", "dataset-timeout-invalid")
	_ = cmd.Flags().Set("file", "ignored.csv")
	_ = cmd.Flags().Set("timeout", "0s")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected non-positive timeout error")
	}

	if !strings.Contains(err.Error(), "timeout must be greater than 0") {
		t.Fatalf("expected timeout validation error, got %v", err)
	}
}

func TestDatasetUploadCommandReturnsUploadStatusError(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "dataset.csv")
	if err := os.WriteFile(filePath, []byte("name\nalice\n"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upload denied", http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)

	c := setupMockClient(t)
	c.EXPECT().
		GetAITaskBuilderDatasetUploadURL(
			gomock.Eq("dataset-321"),
			gomock.Eq("dataset.csv"),
		).
		Return(&client.GetAITaskBuilderDatasetUploadURLResponse{
			ImportID:    "import-321",
			UploadURL:   srv.URL,
			HTTPMethod:  http.MethodPut,
			ContentType: "text/csv",
			ExpiresAt:   "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	cmd := aitaskbuilder.NewDatasetUploadCommand(c, os.Stdout)
	_ = cmd.Flags().Set("dataset-id", "dataset-321")
	_ = cmd.Flags().Set("file", filePath)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected upload status error")
	}

	if !strings.Contains(err.Error(), "upload request returned status 403") {
		t.Fatalf("expected upload status code in error, got %v", err)
	}
}

func TestDatasetUploadCommandRejectsMissingImportID(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "dataset.csv")
	if err := os.WriteFile(filePath, []byte("name\nalice\n"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	c := setupMockClient(t)
	c.EXPECT().
		GetAITaskBuilderDatasetUploadURL(
			gomock.Eq("dataset-missing-import"),
			gomock.Eq("dataset.csv"),
		).
		Return(&client.GetAITaskBuilderDatasetUploadURLResponse{
			UploadURL:   "https://example.com/upload",
			HTTPMethod:  http.MethodPut,
			ContentType: "text/csv",
			ExpiresAt:   "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	cmd := aitaskbuilder.NewDatasetUploadCommand(c, os.Stdout)
	_ = cmd.Flags().Set("dataset-id", "dataset-missing-import")
	_ = cmd.Flags().Set("file", filePath)

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected missing import ID error")
	}

	if !strings.Contains(err.Error(), "import_id is missing in response") {
		t.Fatalf("expected missing import ID error, got %v", err)
	}
}

func TestDatasetUploadCommandCapsPollSleepToRemainingTimeout(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "dataset.csv")
	if err := os.WriteFile(filePath, []byte("name\nalice\n"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	sleepDurations := make([]time.Duration, 0, 1)
	restoreSleep := setDatasetUploadPollSleepForTesting(func(d time.Duration) {
		sleepDurations = append(sleepDurations, d)
		time.Sleep(d)
	})
	defer restoreSleep()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := setupMockClient(t)
	c.EXPECT().
		GetAITaskBuilderDatasetUploadURL(
			gomock.Eq("dataset-short-timeout"),
			gomock.Eq("dataset.csv"),
		).
		Return(&client.GetAITaskBuilderDatasetUploadURLResponse{
			ImportID:    "import-short-timeout",
			UploadURL:   srv.URL,
			HTTPMethod:  http.MethodPut,
			ContentType: "text/csv",
			ExpiresAt:   "2099-01-01T00:00:00Z",
		}, nil).
		Times(1)

	c.EXPECT().
		GetAITaskBuilderDatasetImportStatus(
			gomock.Eq("dataset-short-timeout"),
			gomock.Eq("import-short-timeout"),
		).
		Return(&client.GetAITaskBuilderDatasetImportStatusResponse{
			DatasetImportJob: model.DatasetImportJob{
				Status: model.DatasetImportJobStatusProcessing,
			},
		}, nil).
		MinTimes(1)

	cmd := aitaskbuilder.NewDatasetUploadCommand(c, os.Stdout)
	_ = cmd.Flags().Set("dataset-id", "dataset-short-timeout")
	_ = cmd.Flags().Set("file", filePath)
	_ = cmd.Flags().Set("timeout", "20ms")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if len(sleepDurations) == 0 {
		t.Fatal("expected at least one poll sleep")
	}

	if sleepDurations[0] >= 3*time.Second {
		t.Fatalf("expected capped poll sleep below %s, got %s", 3*time.Second, sleepDurations[0])
	}
}
