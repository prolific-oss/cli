package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

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
		Return([]client.ExportJobListItem{
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
		}, nil).
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
