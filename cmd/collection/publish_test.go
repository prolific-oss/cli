package collection_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/collection"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewPublishCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)

	use := "publish <collection-id>"
	short := "Publish a collection as a study"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestPublishCommandRequiresCollectionID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatalf("expected error for missing collection ID, got nil")
	}
}

func TestPublishCommandRequiresParticipants(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err == nil {
		t.Fatalf("expected error for missing participants, got nil")
	}

	expectedErr := "please provide a valid number of participants"
	if err.Error() != "please provide a valid number of participants using --participants or -p" {
		t.Fatalf("expected error message containing %q, got: %s", expectedErr, err.Error())
	}
}

func TestPublishCommandSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	testCollection := &model.Collection{
		ID:        testCollectionID,
		Name:      "Test Collection",
		CreatedAt: time.Now(),
		CreatedBy: "test-user",
		ItemCount: 10,
		TaskDetails: &model.TaskDetails{
			TaskName:         "Test Task Name",
			TaskIntroduction: "Test task introduction",
		},
	}

	testStudy := &model.Study{
		ID:                   "study-123",
		Name:                 "Test Task Name",
		Status:               "active",
		TotalAvailablePlaces: 100,
	}

	mockClient.
		EXPECT().
		GetCollection(gomock.Eq(testCollectionID)).
		Return(testCollection, nil).
		Times(1)

	mockClient.
		EXPECT().
		CreateStudy(gomock.Any()).
		DoAndReturn(func(s model.CreateStudy) (*model.Study, error) {
			if s.DataCollectionMethod != model.DataCollectionMethodAITBCollection {
				t.Errorf("expected DataCollectionMethod %s, got %s", model.DataCollectionMethodAITBCollection, s.DataCollectionMethod)
			}
			if s.DataCollectionID != testCollectionID {
				t.Errorf("expected DataCollectionID %s, got %s", testCollectionID, s.DataCollectionID)
			}
			if s.TotalAvailablePlaces != 100 {
				t.Errorf("expected TotalAvailablePlaces 100, got %d", s.TotalAvailablePlaces)
			}
			return testStudy, nil
		}).
		Times(1)

	mockClient.
		EXPECT().
		TransitionStudy(gomock.Eq("study-123"), gomock.Eq(model.TransitionStudyPublish)).
		Return(&client.TransitionStudyResponse{}, nil).
		Times(1)

	mockClient.
		EXPECT().
		GetStudy(gomock.Eq("study-123")).
		Return(testStudy, nil).
		Times(1)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)
	_ = cmd.Flags().Set("participants", "100")

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected output, got empty string")
	}
}

func TestPublishCommandCollectionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	mockClient.
		EXPECT().
		GetCollection(gomock.Eq(testCollectionID)).
		Return(nil, errors.New("collection not found")).
		Times(1)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)
	_ = cmd.Flags().Set("participants", "100")

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("failed to get collection")) {
		t.Errorf("expected error to contain 'failed to get collection', got: %s", err.Error())
	}
}

func TestPublishCommandCreateStudyError(t *testing.T) {
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

	mockClient.
		EXPECT().
		CreateStudy(gomock.Any()).
		Return(nil, errors.New("validation error")).
		Times(1)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)
	_ = cmd.Flags().Set("participants", "100")

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("failed to create study")) {
		t.Errorf("expected error to contain 'failed to create study', got: %s", err.Error())
	}
}

func TestPublishCommandTransitionError(t *testing.T) {
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

	testStudy := &model.Study{
		ID:                   "study-123",
		Name:                 "Test Collection",
		Status:               "unpublished",
		TotalAvailablePlaces: 100,
	}

	mockClient.
		EXPECT().
		GetCollection(gomock.Eq(testCollectionID)).
		Return(testCollection, nil).
		Times(1)

	mockClient.
		EXPECT().
		CreateStudy(gomock.Any()).
		Return(testStudy, nil).
		Times(1)

	mockClient.
		EXPECT().
		TransitionStudy(gomock.Eq("study-123"), gomock.Eq(model.TransitionStudyPublish)).
		Return(nil, errors.New("insufficient funds")).
		Times(1)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)
	_ = cmd.Flags().Set("participants", "100")

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("failed to publish study")) {
		t.Errorf("expected error to contain 'failed to publish study', got: %s", err.Error())
	}
}

func TestPublishCommandUsesCustomNameAndDescription(t *testing.T) {
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

	testStudy := &model.Study{
		ID:                   "study-123",
		Name:                 "Custom Study Name",
		Status:               "active",
		TotalAvailablePlaces: 50,
	}

	mockClient.
		EXPECT().
		GetCollection(gomock.Eq(testCollectionID)).
		Return(testCollection, nil).
		Times(1)

	mockClient.
		EXPECT().
		CreateStudy(gomock.Any()).
		DoAndReturn(func(s model.CreateStudy) (*model.Study, error) {
			if s.Name != "Custom Study Name" {
				t.Errorf("expected Name 'Custom Study Name', got %s", s.Name)
			}
			if s.Description != "Custom description" {
				t.Errorf("expected Description 'Custom description', got %s", s.Description)
			}
			return testStudy, nil
		}).
		Times(1)

	mockClient.
		EXPECT().
		TransitionStudy(gomock.Eq("study-123"), gomock.Eq(model.TransitionStudyPublish)).
		Return(&client.TransitionStudyResponse{}, nil).
		Times(1)

	mockClient.
		EXPECT().
		GetStudy(gomock.Eq("study-123")).
		Return(testStudy, nil).
		Times(1)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)
	_ = cmd.Flags().Set("participants", "50")
	_ = cmd.Flags().Set("name", "Custom Study Name")
	_ = cmd.Flags().Set("description", "Custom description")

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}
