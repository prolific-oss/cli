package collection_test

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/collection"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewExportDeleteCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewExportDeleteCommand(mockClient, &buf)

	if cmd.Use != "delete <collection-id> <export-id>" {
		t.Fatalf("expected use: delete <collection-id> <export-id>; got %s", cmd.Use)
	}
}

func TestExportDeleteCommandRequiresArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewExportDeleteCommand(mockClient, &buf)

	if err := cmd.RunE(cmd, []string{}); err == nil {
		t.Fatal("expected error for missing args, got nil")
	}

	if err := cmd.RunE(cmd, []string{testCollectionID}); err == nil {
		t.Fatal("expected error for missing export ID, got nil")
	}
}

func TestExportDeleteCommandHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		DeleteCollectionExport(gomock.Eq(testCollectionID), gomock.Eq(testExportID)).
		Return(nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportDeleteCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testCollectionID, testExportID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := b.String()
	want := "Export " + testExportID + " for collection " + testCollectionID + " deleted."
	if !strings.Contains(output, want) {
		t.Errorf("expected confirmation message %q, got: %s", want, output)
	}
}

func TestExportDeleteCommandError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		DeleteCollectionExport(gomock.Eq(testCollectionID), gomock.Eq(testExportID)).
		Return(errors.New("conflict: export still generating")).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportDeleteCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testCollectionID, testExportID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error on client failure, got nil")
	}
}

func TestExportDeleteCommandFeatureNotEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		DeleteCollectionExport(gomock.Eq(testCollectionID), gomock.Eq(testExportID)).
		Return(errors.New("request failed: you do not currently have permission to access this feature")).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := collection.NewExportDeleteCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testCollectionID, testExportID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error for feature-not-enabled, got: %v", err)
	}
}
