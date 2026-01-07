package collection_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/collection"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

const testCollectionID = "123456789"

func TestNewGetCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewGetCommand(mockClient, &buf)

	use := "get <collection-id>"
	short := "Get details of a specific collection"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestGetCommandRequiresCollectionID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewGetCommand(mockClient, &buf)

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatalf("expected error for missing collection ID, got nil")
	}
}

func TestGetCommandCallsGetCollection(t *testing.T) {
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
	cmd := collection.NewGetCommand(mockClient, &buf)

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected output, got empty string")
	}
}

func TestGetCommandReturnsErrorOnClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.
		EXPECT().
		GetCollection(gomock.Eq(testCollectionID)).
		Return(nil, errors.New("collection not found")).
		Times(1)

	var buf bytes.Buffer
	cmd := collection.NewGetCommand(mockClient, &buf)

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetCommandOutputContainsCollectionDetails(t *testing.T) {
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
	cmd := collection.NewGetCommand(mockClient, &buf)

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	expectedStrings := []string{
		"My Test Collection",
		testCollectionID,
		"user@example.com",
		"25",
	}

	for _, expected := range expectedStrings {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("expected output to contain %q, got: %s", expected, output)
		}
	}
}
