package collection_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func TestPublishCommandRequiresParticipantsOrTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err == nil {
		t.Fatalf("expected error for missing participants/template, got nil")
	}

	expectedErr := "please provide a valid number of participants using --participants or -p, or provide a template file using --template or -t"
	if err.Error() != expectedErr {
		t.Fatalf("expected error message %q, got: %s", expectedErr, err.Error())
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

func TestPublishCommandWithTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	// Create a temporary template file
	templateContent := `{
		"name": "Template Study",
		"internal_name": "Template Study Internal",
		"description": "A study from template",
		"reward": 100,
		"total_available_places": 200,
		"prolific_id_option": "question",
		"completion_code": "TEMPLATE01",
		"estimated_completion_time": 5,
		"device_compatibility": ["desktop", "mobile"]
	}`

	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "template.json")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0600); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	testCollection := &model.Collection{
		ID:        testCollectionID,
		Name:      "Test Collection",
		CreatedAt: time.Now(),
		CreatedBy: "test-user",
		ItemCount: 10,
	}

	testStudy := &model.Study{
		ID:                   "study-456",
		Name:                 "Template Study",
		Status:               "active",
		TotalAvailablePlaces: 200,
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
			// Verify the collection fields are set correctly
			if s.DataCollectionMethod != model.DataCollectionMethodAITBCollection {
				t.Errorf("expected DataCollectionMethod %s, got %s", model.DataCollectionMethodAITBCollection, s.DataCollectionMethod)
			}
			if s.DataCollectionID != testCollectionID {
				t.Errorf("expected DataCollectionID %s, got %s", testCollectionID, s.DataCollectionID)
			}
			// Verify template values are used
			if s.Name != "Template Study" {
				t.Errorf("expected Name 'Template Study', got %s", s.Name)
			}
			if s.TotalAvailablePlaces != 200 {
				t.Errorf("expected TotalAvailablePlaces 200, got %d", s.TotalAvailablePlaces)
			}
			if s.Reward != 100 {
				t.Errorf("expected Reward 100, got %f", s.Reward)
			}
			return testStudy, nil
		}).
		Times(1)

	mockClient.
		EXPECT().
		TransitionStudy(gomock.Eq("study-456"), gomock.Eq(model.TransitionStudyPublish)).
		Return(&client.TransitionStudyResponse{}, nil).
		Times(1)

	mockClient.
		EXPECT().
		GetStudy(gomock.Eq("study-456")).
		Return(testStudy, nil).
		Times(1)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)
	_ = cmd.Flags().Set("template", templatePath)

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected output, got empty string")
	}
}

func TestPublishCommandWithTemplateOverridesCollectionFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	// Create a template that has different data_collection_method, data_collection_id,
	// and an external_study_url. These should be overridden/cleared by the command.
	templateContent := `{
		"name": "Template Study",
		"description": "A study from template",
		"reward": 100,
		"total_available_places": 50,
		"external_study_url": "https://example.com/study",
		"data_collection_method": "SOME_OTHER_METHOD",
		"data_collection_id": "some-other-id",
		"prolific_id_option": "question",
		"completion_code": "TEST01",
		"estimated_completion_time": 5,
		"device_compatibility": ["desktop"]
	}`

	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "template.json")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0600); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	testCollection := &model.Collection{
		ID:        testCollectionID,
		Name:      "Test Collection",
		CreatedAt: time.Now(),
		CreatedBy: "test-user",
		ItemCount: 10,
	}

	testStudy := &model.Study{
		ID:                   "study-789",
		Name:                 "Template Study",
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
			// Verify the collection fields are OVERRIDDEN correctly
			if s.DataCollectionMethod != model.DataCollectionMethodAITBCollection {
				t.Errorf("expected DataCollectionMethod to be overridden to %s, got %s", model.DataCollectionMethodAITBCollection, s.DataCollectionMethod)
			}
			if s.DataCollectionID != testCollectionID {
				t.Errorf("expected DataCollectionID to be overridden to %s, got %s", testCollectionID, s.DataCollectionID)
			}
			// Verify external_study_url is cleared (incompatible with data collection method)
			if s.ExternalStudyURL != "" {
				t.Errorf("expected ExternalStudyURL to be cleared, got %s", s.ExternalStudyURL)
			}
			return testStudy, nil
		}).
		Times(1)

	mockClient.
		EXPECT().
		TransitionStudy(gomock.Eq("study-789"), gomock.Eq(model.TransitionStudyPublish)).
		Return(&client.TransitionStudyResponse{}, nil).
		Times(1)

	mockClient.
		EXPECT().
		GetStudy(gomock.Eq("study-789")).
		Return(testStudy, nil).
		Times(1)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)
	_ = cmd.Flags().Set("template", templatePath)

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPublishCommandWithInvalidTemplatePath(t *testing.T) {
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
	cmd := collection.NewPublishCommand(mockClient, &buf)
	_ = cmd.Flags().Set("template", "/nonexistent/path/template.json")

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err == nil {
		t.Fatal("expected error for invalid template path, got nil")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("failed to read template file")) {
		t.Errorf("expected error to contain 'failed to read template file', got: %s", err.Error())
	}
}

func TestCreateStudyOmitsEmptyExternalStudyURL(t *testing.T) {
	// This test verifies that when ExternalStudyURL is empty, it is omitted
	// from the JSON serialization entirely (not sent as "external_study_url": "").
	// This is important because the API rejects requests that include
	// external_study_url when using data_collection_method.

	study := model.CreateStudy{
		Name:                 "Test Study",
		Description:          "Test description",
		TotalAvailablePlaces: 100,
		DataCollectionMethod: model.DataCollectionMethodAITBCollection,
		DataCollectionID:     "collection-123",
		ExternalStudyURL:     "", // Empty - should be omitted from JSON
	}

	jsonBytes, err := json.Marshal(study)
	if err != nil {
		t.Fatalf("failed to marshal CreateStudy: %v", err)
	}

	jsonStr := string(jsonBytes)

	if strings.Contains(jsonStr, "external_study_url") {
		t.Errorf("expected external_study_url to be omitted from JSON when empty, but got: %s", jsonStr)
	}

	// Also verify that the data collection fields ARE present
	if !strings.Contains(jsonStr, "data_collection_method") {
		t.Errorf("expected data_collection_method to be present in JSON, but got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, "data_collection_id") {
		t.Errorf("expected data_collection_id to be present in JSON, but got: %s", jsonStr)
	}
}

func TestPublishCommandWithTemplateUsesCollectionDescriptionFallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	// Create a template WITHOUT a description
	templateContent := `{
		"name": "Template Study",
		"reward": 100,
		"total_available_places": 50,
		"prolific_id_option": "question",
		"completion_code": "TEST01",
		"estimated_completion_time": 5,
		"device_compatibility": ["desktop"]
	}`

	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "template.json")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0600); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	testCollection := &model.Collection{
		ID:        testCollectionID,
		Name:      "Test Collection",
		CreatedAt: time.Now(),
		CreatedBy: "test-user",
		ItemCount: 10,
		TaskDetails: &model.TaskDetails{
			TaskName:         "Collection Task Name",
			TaskIntroduction: "This is the task introduction from the collection",
		},
	}

	testStudy := &model.Study{
		ID:                   "study-desc-test",
		Name:                 "Template Study",
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
			// Verify the description falls back to collection's task introduction
			expectedDescription := "This is the task introduction from the collection"
			if s.Description != expectedDescription {
				t.Errorf("expected Description to fall back to %q, got %q", expectedDescription, s.Description)
			}
			return testStudy, nil
		}).
		Times(1)

	mockClient.
		EXPECT().
		TransitionStudy(gomock.Eq("study-desc-test"), gomock.Eq(model.TransitionStudyPublish)).
		Return(&client.TransitionStudyResponse{}, nil).
		Times(1)

	mockClient.
		EXPECT().
		GetStudy(gomock.Eq("study-desc-test")).
		Return(testStudy, nil).
		Times(1)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)
	_ = cmd.Flags().Set("template", templatePath)

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPublishCommandWithTemplateAndParticipantsFlag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	// Create a template with total_available_places set to 50
	templateContent := `{
		"name": "Template Study",
		"description": "Template description",
		"reward": 100,
		"total_available_places": 50,
		"prolific_id_option": "question",
		"completion_code": "TEST01",
		"estimated_completion_time": 5,
		"device_compatibility": ["desktop"]
	}`

	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "template.json")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0600); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	testCollection := &model.Collection{
		ID:        testCollectionID,
		Name:      "Test Collection",
		CreatedAt: time.Now(),
		CreatedBy: "test-user",
		ItemCount: 10,
	}

	testStudy := &model.Study{
		ID:                   "study-override-test",
		Name:                 "Template Study",
		Status:               "active",
		TotalAvailablePlaces: 150,
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
			// Verify -p flag overrides template's total_available_places
			if s.TotalAvailablePlaces != 150 {
				t.Errorf("expected TotalAvailablePlaces to be overridden to 150, got %d", s.TotalAvailablePlaces)
			}
			// Verify other template values are still used
			if s.Reward != 100 {
				t.Errorf("expected Reward from template (100), got %f", s.Reward)
			}
			return testStudy, nil
		}).
		Times(1)

	mockClient.
		EXPECT().
		TransitionStudy(gomock.Eq("study-override-test"), gomock.Eq(model.TransitionStudyPublish)).
		Return(&client.TransitionStudyResponse{}, nil).
		Times(1)

	mockClient.
		EXPECT().
		GetStudy(gomock.Eq("study-override-test")).
		Return(testStudy, nil).
		Times(1)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)
	_ = cmd.Flags().Set("template", templatePath)
	_ = cmd.Flags().Set("participants", "150") // Override template's 50

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPublishCommandWithTemplateAndNameDescriptionFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	// Create a template with name and description set
	templateContent := `{
		"name": "Template Name",
		"description": "Template description",
		"reward": 100,
		"total_available_places": 50,
		"prolific_id_option": "question",
		"completion_code": "TEST01",
		"estimated_completion_time": 5,
		"device_compatibility": ["desktop"]
	}`

	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "template.json")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0600); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	testCollection := &model.Collection{
		ID:        testCollectionID,
		Name:      "Test Collection",
		CreatedAt: time.Now(),
		CreatedBy: "test-user",
		ItemCount: 10,
	}

	testStudy := &model.Study{
		ID:                   "study-name-desc-test",
		Name:                 "Flag Override Name",
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
			// Verify -n and -d flags override template values
			if s.Name != "Flag Override Name" {
				t.Errorf("expected Name to be overridden to 'Flag Override Name', got %q", s.Name)
			}
			if s.InternalName != "Flag Override Name" {
				t.Errorf("expected InternalName to be overridden to 'Flag Override Name', got %q", s.InternalName)
			}
			if s.Description != "Flag override description" {
				t.Errorf("expected Description to be overridden to 'Flag override description', got %q", s.Description)
			}
			// Verify other template values are still used
			if s.Reward != 100 {
				t.Errorf("expected Reward from template (100), got %f", s.Reward)
			}
			return testStudy, nil
		}).
		Times(1)

	mockClient.
		EXPECT().
		TransitionStudy(gomock.Eq("study-name-desc-test"), gomock.Eq(model.TransitionStudyPublish)).
		Return(&client.TransitionStudyResponse{}, nil).
		Times(1)

	mockClient.
		EXPECT().
		GetStudy(gomock.Eq("study-name-desc-test")).
		Return(testStudy, nil).
		Times(1)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)
	_ = cmd.Flags().Set("template", templatePath)
	_ = cmd.Flags().Set("name", "Flag Override Name")
	_ = cmd.Flags().Set("description", "Flag override description")

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPublishCommandWithTemplateUsesDefaultDescriptionFallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	// Create a template WITHOUT a description
	templateContent := `{
		"name": "Template Study",
		"reward": 100,
		"total_available_places": 50,
		"prolific_id_option": "question",
		"completion_code": "TEST01",
		"estimated_completion_time": 5,
		"device_compatibility": ["desktop"]
	}`

	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "template.json")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0600); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	// Collection WITHOUT TaskDetails
	testCollection := &model.Collection{
		ID:          testCollectionID,
		Name:        "My Test Collection",
		CreatedAt:   time.Now(),
		CreatedBy:   "test-user",
		ItemCount:   10,
		TaskDetails: nil, // No task details
	}

	testStudy := &model.Study{
		ID:                   "study-fallback-test",
		Name:                 "Template Study",
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
			// Verify the description falls back to the default format
			expectedDescription := "Study for collection: My Test Collection"
			if s.Description != expectedDescription {
				t.Errorf("expected Description to fall back to %q, got %q", expectedDescription, s.Description)
			}
			return testStudy, nil
		}).
		Times(1)

	mockClient.
		EXPECT().
		TransitionStudy(gomock.Eq("study-fallback-test"), gomock.Eq(model.TransitionStudyPublish)).
		Return(&client.TransitionStudyResponse{}, nil).
		Times(1)

	mockClient.
		EXPECT().
		GetStudy(gomock.Eq("study-fallback-test")).
		Return(testStudy, nil).
		Times(1)

	var buf bytes.Buffer
	cmd := collection.NewPublishCommand(mockClient, &buf)
	_ = cmd.Flags().Set("template", templatePath)

	err := cmd.RunE(cmd, []string{testCollectionID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}
