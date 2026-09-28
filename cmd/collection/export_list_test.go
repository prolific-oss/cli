package collection_test

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/collection"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewExportListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewExportListCommand(mockClient, &buf)

	if cmd.Use != "list <collection-id>" {
		t.Fatalf("expected use: list <collection-id>; got %s", cmd.Use)
	}
}

func TestExportListCommandRequiresCollectionID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewExportListCommand(mockClient, &buf)

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for missing collection ID, got nil")
	}
}

func TestExportListCommandHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		ListCollectionExportJobs(gomock.Eq(testCollectionID)).
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
	cmd := collection.NewExportListCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testCollectionID})
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

func TestExportListCommandEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		ListCollectionExportJobs(gomock.Eq(testCollectionID)).
		Return([]client.ExportJobListItem{}, nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportListCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testCollectionID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := b.String()
	if !strings.Contains(output, "No export jobs found for collection "+testCollectionID) {
		t.Errorf("expected friendly empty message, got: %s", output)
	}
}

func TestExportListCommandError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		ListCollectionExportJobs(gomock.Eq(testCollectionID)).
		Return(nil, errors.New("network error")).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportListCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testCollectionID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error on client failure, got nil")
	}
}

func TestExportListCommandFeatureNotEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		ListCollectionExportJobs(gomock.Eq(testCollectionID)).
		Return(nil, errors.New("request failed: you do not currently have permission to access this feature")).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportListCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testCollectionID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error for feature-not-enabled, got: %v", err)
	}
}
