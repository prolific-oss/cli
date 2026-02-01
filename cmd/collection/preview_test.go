package collection_test

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/collection"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

// noOpBrowserOpener is a no-op browser opener for testing
func noOpBrowserOpener(url string) error {
	return nil
}

func TestNewPreviewCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewPreviewCommandWithOpener(mockClient, &buf, noOpBrowserOpener)

	use := "preview <collection-id>"
	short := "Preview a collection in the browser"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestPreviewCommandRequiresCollectionID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewPreviewCommandWithOpener(mockClient, &buf, noOpBrowserOpener)

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatalf("expected error for missing collection ID, got nil")
	}

	expected := "please provide a collection ID"
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

func TestPreviewCommandRequiresNonEmptyCollectionID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewPreviewCommandWithOpener(mockClient, &buf, noOpBrowserOpener)

	err := cmd.RunE(cmd, []string{""})
	if err == nil {
		t.Fatalf("expected error for empty collection ID, got nil")
	}

	expected := "please provide a collection ID"
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

func TestPreviewCommandCallsGetCollection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	testCollection := &model.Collection{
		ID:        testCollectionID,
		Name:      "Test Collection",
		CreatedAt: time.Now(),
		CreatedBy: "test-user",
		ItemCount: 10,
	}

	mockClient.
		EXPECT().
		GetCollection(gomock.Eq(testCollectionID)).
		Return(testCollection, nil).
		Times(1)

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	cmd := collection.NewPreviewCommandWithOpener(mockClient, writer, noOpBrowserOpener)

	err := cmd.RunE(cmd, []string{testCollectionID})
	writer.Flush()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPreviewCommandReturnsErrorOnClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.
		EXPECT().
		GetCollection(gomock.Eq(testCollectionID)).
		Return(nil, errors.New("collection not found")).
		Times(1)

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	cmd := collection.NewPreviewCommandWithOpener(mockClient, writer, noOpBrowserOpener)

	err := cmd.RunE(cmd, []string{testCollectionID})
	writer.Flush()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expected := "error: collection not found"
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

func TestPreviewCommandHandlesFeatureNotEnabledError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	// The feature not enabled error must contain "request failed", "permission", and "feature"
	featureError := errors.New("request failed: you do not currently have permission to access this feature")

	mockClient.
		EXPECT().
		GetCollection(gomock.Eq(testCollectionID)).
		Return(nil, featureError).
		Times(1)

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	cmd := collection.NewPreviewCommandWithOpener(mockClient, writer, noOpBrowserOpener)

	err := cmd.RunE(cmd, []string{testCollectionID})
	writer.Flush()

	// When feature is not enabled, the command should not return an error
	// but should display a feature access message
	if err != nil {
		t.Fatalf("expected no error for feature not enabled, got: %v", err)
	}
}

func TestPreviewCommandOutputContainsURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	testCollection := &model.Collection{
		ID:        testCollectionID,
		Name:      "My Test Collection",
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		CreatedBy: "user@example.com",
		ItemCount: 25,
	}

	mockClient.
		EXPECT().
		GetCollection(gomock.Eq(testCollectionID)).
		Return(testCollection, nil).
		Times(1)

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	cmd := collection.NewPreviewCommandWithOpener(mockClient, writer, noOpBrowserOpener)

	err := cmd.RunE(cmd, []string{testCollectionID})
	writer.Flush()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Check that the output contains the expected URL components
	expectedStrings := []string{
		"Opening collection preview in browser",
		"data-collection-tool/collections/" + testCollectionID,
		"preview=true",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got: %s", expected, output)
		}
	}
}
