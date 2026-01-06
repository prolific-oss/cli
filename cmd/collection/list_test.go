package collection_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/collection"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

const testWorkspaceID = "6655b8281cc82a88996f0bbb"

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	cmd := collection.NewListCommand(mockClient, os.Stdout)

	use := "list"
	short := "List all collections in a workspace"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestListCommandRequiresWorkspaceID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	cmd := collection.NewListCommand(mockClient, os.Stdout)

	_ = cmd.Flags().Set("non-interactive", "true")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatalf("expected error for missing workspace ID, got nil")
	}

	expectedErr := "workspace ID is required"
	if err.Error() != expectedErr {
		t.Fatalf("expected error: %s; got %s", expectedErr, err.Error())
	}
}

func TestListCommandPassesWorkspaceIDToGetCollections(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	collectionResponse := client.ListCollectionsResponse{
		Results: []model.Collection{},
	}

	mockClient.
		EXPECT().
		GetCollections(gomock.Eq(testWorkspaceID), gomock.Eq(client.DefaultRecordLimit), gomock.Eq(client.DefaultRecordOffset)).
		Return(&collectionResponse, nil).
		Times(1)

	cmd := collection.NewListCommand(mockClient, os.Stdout)

	_ = cmd.Flags().Set("workspace", testWorkspaceID)
	_ = cmd.Flags().Set("non-interactive", "true")

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestListCommandPassesLimitAndOffsetToGetCollections(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	limit := 50
	offset := 10

	collectionResponse := client.ListCollectionsResponse{
		Results: []model.Collection{},
	}

	mockClient.
		EXPECT().
		GetCollections(gomock.Eq(testWorkspaceID), gomock.Eq(limit), gomock.Eq(offset)).
		Return(&collectionResponse, nil).
		Times(1)

	cmd := collection.NewListCommand(mockClient, os.Stdout)

	_ = cmd.Flags().Set("workspace", testWorkspaceID)
	_ = cmd.Flags().Set("limit", "50")
	_ = cmd.Flags().Set("offset", "10")
	_ = cmd.Flags().Set("non-interactive", "true")

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestListCommandUsesJSONRendererWithJsonFlag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	collectionResponse := client.ListCollectionsResponse{
		Results: []model.Collection{},
	}

	mockClient.
		EXPECT().
		GetCollections(gomock.Eq(testWorkspaceID), gomock.Eq(client.DefaultRecordLimit), gomock.Eq(client.DefaultRecordOffset)).
		Return(&collectionResponse, nil).
		Times(1)

	cmd := collection.NewListCommand(mockClient, os.Stdout)

	_ = cmd.Flags().Set("workspace", testWorkspaceID)
	_ = cmd.Flags().Set("json", "true")

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}
