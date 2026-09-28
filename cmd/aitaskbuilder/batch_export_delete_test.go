package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewBatchExportDeleteCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := aitaskbuilder.NewBatchExportDeleteCommand(mockClient, &buf)

	if cmd.Use != "delete <batch-id> <export-id>" {
		t.Fatalf("expected use: delete <batch-id> <export-id>; got %s", cmd.Use)
	}
}

func TestBatchExportDeleteCommandRequiresArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := aitaskbuilder.NewBatchExportDeleteCommand(mockClient, &buf)

	if err := cmd.RunE(cmd, []string{}); err == nil {
		t.Fatal("expected error for missing args, got nil")
	}

	if err := cmd.RunE(cmd, []string{testBatchID}); err == nil {
		t.Fatal("expected error for missing export ID, got nil")
	}
}

func TestBatchExportDeleteCommandHappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		DeleteBatchExport(gomock.Eq(testBatchID), gomock.Eq(testBatchExportID)).
		Return(nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportDeleteCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID, testBatchExportID})
	w.Flush()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := b.String()
	want := "Export " + testBatchExportID + " for batch " + testBatchID + " deleted."
	if !strings.Contains(output, want) {
		t.Errorf("expected confirmation message %q, got: %s", want, output)
	}
}

func TestBatchExportDeleteCommandError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.EXPECT().
		DeleteBatchExport(gomock.Eq(testBatchID), gomock.Eq(testBatchExportID)).
		Return(errors.New("conflict: export still generating")).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	cmd := aitaskbuilder.NewBatchExportDeleteCommand(mockClient, w)

	err := cmd.RunE(cmd, []string{testBatchID, testBatchExportID})
	w.Flush()
	if err == nil {
		t.Fatal("expected error on client failure, got nil")
	}
}
