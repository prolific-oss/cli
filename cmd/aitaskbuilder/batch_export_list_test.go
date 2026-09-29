package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

func batchExportListSampleJobs() []client.ExportJobListItem {
	return []client.ExportJobListItem{
		{
			ExportID:  "export-unfiltered",
			Filter:    nil,
			Status:    "complete",
			CreatedAt: "2024-01-01T00:00:00Z",
		},
		{
			ExportID: "export-filtered",
			Filter: &client.ExportJobFilter{
				StudyID: "study-id-789",
				From:    "2024-01-01T00:00:00Z",
				To:      "2024-02-01T00:00:00Z",
			},
			Status:    "generating",
			CreatedAt: "2024-02-01T00:00:00Z",
		},
	}
}

func TestNewBatchExportListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := aitaskbuilder.NewBatchExportListCommand(mockClient, &buf)

	if cmd.Use != "list <batch-id>" {
		t.Fatalf("expected use: list <batch-id>; got %s", cmd.Use)
	}
}

func TestBatchExportListCommandRegistersOutputFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := aitaskbuilder.NewBatchExportListCommand(mockClient, &buf)

	for _, name := range []string{"json", "csv", "table", "non-interactive"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected --%s flag to be registered", name)
		}
	}
}

func TestBatchExportListCommandRequiresBatchID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := aitaskbuilder.NewBatchExportListCommand(mockClient, &buf)

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for missing batch ID, got nil")
	}
}

func TestBatchExportListCommandHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		ListBatchExportJobs(gomock.Eq(testBatchID)).
		Return(batchExportListSampleJobs(), nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportListCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := b.String()
	for _, want := range []string{
		"export-unfiltered",
		"export-filtered",
		"complete",
		"generating",
		"study_id=study-id-789",
		"from=2024-01-01T00:00:00Z",
		"to=2024-02-01T00:00:00Z",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got: %s", want, output)
		}
	}
}

func TestBatchExportListCommandEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		ListBatchExportJobs(gomock.Eq(testBatchID)).
		Return([]client.ExportJobListItem{}, nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportListCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := b.String()
	if !strings.Contains(output, "No export jobs found for batch "+testBatchID) {
		t.Errorf("expected friendly empty message, got: %s", output)
	}
}

func TestBatchExportListCommandError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		ListBatchExportJobs(gomock.Eq(testBatchID)).
		Return(nil, errors.New("network error")).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportListCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error on client failure, got nil")
	}
}

func TestBatchExportListCommandJSONOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		ListBatchExportJobs(gomock.Eq(testBatchID)).
		Return(batchExportListSampleJobs(), nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportListCommand(mockClient, w)
	if err := cmd.Flags().Set("json", "true"); err != nil {
		t.Fatalf("failed to set json flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var decoded []client.ExportJobListItem
	if err := json.Unmarshal(b.Bytes(), &decoded); err != nil {
		t.Fatalf("expected valid JSON output, got error %v; output: %s", err, b.String())
	}
	if len(decoded) != 2 {
		t.Fatalf("expected 2 decoded jobs, got %d", len(decoded))
	}
	if decoded[0].Filter != nil {
		t.Errorf("expected first job's filter to be nil, got: %+v", decoded[0].Filter)
	}
	if decoded[1].Filter == nil || decoded[1].Filter.StudyID != "study-id-789" {
		t.Errorf("expected second job's filter to be preserved with study_id, got: %+v", decoded[1].Filter)
	}
}

func TestBatchExportListCommandCSVOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		ListBatchExportJobs(gomock.Eq(testBatchID)).
		Return(batchExportListSampleJobs(), nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportListCommand(mockClient, w)
	if err := cmd.Flags().Set("csv", "true"); err != nil {
		t.Fatalf("failed to set csv flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(b.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("expected valid CSV output, got error %v; output: %s", err, b.String())
	}
	if len(records) != 3 { // header + 2 rows
		t.Fatalf("expected 3 CSV records (header + 2 rows), got %d: %v", len(records), records)
	}
	if records[0][0] != "ExportID" {
		t.Errorf("expected CSV header to start with ExportID, got: %v", records[0])
	}
	if !strings.Contains(records[2][1], "study_id=study-id-789") {
		t.Errorf("expected filtered row's Filter column to contain study_id, got: %v", records[2])
	}
}

func TestBatchExportListCommandTableFlagOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		ListBatchExportJobs(gomock.Eq(testBatchID)).
		Return(batchExportListSampleJobs(), nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportListCommand(mockClient, w)
	// -n/--non-interactive is a hidden alias for --table; both should render
	// the same non-interactive table output.
	if err := cmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatalf("failed to set non-interactive flag: %v", err)
	}

	err := cmd.RunE(cmd, []string{testBatchID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := b.String()
	if !strings.Contains(output, "export-filtered") || !strings.Contains(output, "study_id=study-id-789") {
		t.Errorf("expected -n/--non-interactive to render a table, got: %s", output)
	}
}
