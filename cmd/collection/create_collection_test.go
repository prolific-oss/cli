package collection_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/collection"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

const collectionItems = `[
    {
      "order": 0,
      "page_items": [
        {
          "order": 0,
          "type": "free_text",
          "description": "How was your experience completing this task?"
        },
        {
          "order": 1,
          "type": "multiple_choice",
          "description": "Which option do you prefer?",
          "options": [
            {
              "label": "Response 1",
              "value": "response1"
            },
            {
              "label": "Response 2",
              "value": "response2"
            }
          ],
          "answer_limit": -1
        }
      ]
	}
]`

const taskDetails = `{
    "task_name": "Test Task Name",
    "task_introduction": "Test task introduction",
    "task_steps": "Test task steps"
  }`

const collectionItemsWithContentBlocks = `[
    {
      "order": 0,
      "page_items": [
        {
          "order": 0,
          "type": "rich_text",
          "content": "Welcome to this task. Please read the following instructions carefully."
        },
        {
          "order": 1,
          "type": "image",
          "url": "https://example.com/image.png",
          "alt_text": "Example image",
          "caption": "This is an example image"
        }
      ]
    },
    {
      "order": 1,
      "page_items": [
        {
          "order": 0,
          "type": "free_text",
          "description": "How was your experience completing this task?"
        }
      ]
    }
]`

func TestNewCreateCollectionCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := collection.NewCreateCollectionCommand(c, os.Stdout)

	use := "create"
	short := "Create a new collection"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewCreateCollectionCommandCallsAPIWithJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// Create temporary test file
	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "collection.json")
	templateContent := fmt.Sprintf(`{
  "workspace_id": "6716028cd934ced9bac18658",
  "name": "test-collection",
  "task_details": %s,
  "collection_items": %s
}`, taskDetails, collectionItems)

	err := os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	// Expected response
	response := client.CreateAITaskBuilderCollectionResponse{
		ID:            "collection-123",
		Name:          "test-collection",
		WorkspaceID:   "6716028cd934ced9bac18658",
		SchemaVersion: 1,
		CreatedBy:     "user-456",
		CollectionItems: []model.CollectionPage{
			{
				Order: 0,
				PageItems: []model.CollectionInstruction{
					{
						Order:       0,
						Type:        "free_text",
						Description: "How was your experience completing this task?",
					},
				},
			},
		},
	}

	// Set up mock expectation - use Any since exact payload matching is complex
	c.EXPECT().
		CreateAITaskBuilderCollection(gomock.Any()).
		Return(&response, nil)

	// Execute command with buffer to capture output
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := collection.NewCreateCollectionCommand(c, writer)
	cmd.SetArgs([]string{
		"-t", templateFile,
	})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error; got %s", err.Error())
	}

	writer.Flush()

	// Verify output contains expected strings
	output := b.String()
	expectedStrings := []string{
		"Collection created successfully!",
		"ID:              collection-123",
		"Name:            test-collection",
		"Workspace ID:    6716028cd934ced9bac18658",
		"Schema Version:  1",
		"Created By:      user-456",
		"Pages:           1",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain '%s'; got:\n%s", expected, output)
		}
	}
}

func TestNewCreateCollectionCommandCallsAPIWithYAML(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// Create temporary YAML test file
	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "collection.yaml")
	templateContent := `workspace_id: 6716028cd934ced9bac18658
name: yaml-test-collection
task_details:
  task_name: YAML Task Name
  task_introduction: Welcome to the YAML test task
  task_steps: Follow the steps carefully
collection_items:
  - order: 0
    page_items:
      - order: 0
        type: rich_text
        content: This is a rich text content block from YAML
      - order: 1
        type: image
        url: https://example.com/yaml-image.png
        alt_text: YAML test image
        caption: An image from YAML config
  - order: 1
    page_items:
      - order: 0
        type: free_text
        description: YAML test description
      - order: 1
        type: multiple_choice
        description: Which option do you prefer?
        disable_dropdown: true
        answer_limit: 1
        options:
          - label: Option A
            value: option_a
          - label: Option B
            value: option_b
`

	err := os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	// Expected response
	response := client.CreateAITaskBuilderCollectionResponse{
		ID:              "collection-yaml-123",
		Name:            "yaml-test-collection",
		WorkspaceID:     "6716028cd934ced9bac18658",
		SchemaVersion:   1,
		CreatedBy:       "user-789",
		CollectionItems: []model.CollectionPage{},
	}

	var capturedPayload model.CreateAITaskBuilderCollection
	c.EXPECT().
		CreateAITaskBuilderCollection(gomock.Any()).
		DoAndReturn(func(payload model.CreateAITaskBuilderCollection) (*client.CreateAITaskBuilderCollectionResponse, error) {
			capturedPayload = payload
			return &response, nil
		})

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := collection.NewCreateCollectionCommand(c, writer)
	cmd.SetArgs([]string{"-t", templateFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error; got %s", err.Error())
	}

	writer.Flush()

	if !strings.Contains(b.String(), "yaml-test-collection") {
		t.Fatalf("expected output to contain 'yaml-test-collection'; got:\n%s", b.String())
	}

	// Verify task_details were correctly parsed from YAML
	if capturedPayload.TaskDetails == nil {
		t.Fatal("expected task_details to be set; got nil")
	}
	if capturedPayload.TaskDetails.TaskName != "YAML Task Name" {
		t.Fatalf("expected task_name 'YAML Task Name'; got '%s'", capturedPayload.TaskDetails.TaskName)
	}
	if capturedPayload.TaskDetails.TaskIntroduction != "Welcome to the YAML test task" {
		t.Fatalf("expected task_introduction to be set; got '%s'", capturedPayload.TaskDetails.TaskIntroduction)
	}
	if capturedPayload.TaskDetails.TaskSteps != "Follow the steps carefully" {
		t.Fatalf("expected task_steps to be set; got '%s'", capturedPayload.TaskDetails.TaskSteps)
	}

	// Verify collection items were correctly parsed from YAML
	if len(capturedPayload.CollectionItems) != 2 {
		t.Fatalf("expected 2 collection items; got %d", len(capturedPayload.CollectionItems))
	}

	// Verify first page - content blocks
	firstPage := capturedPayload.CollectionItems[0]
	if len(firstPage.PageItems) != 2 {
		t.Fatalf("expected 2 page items in first page; got %d", len(firstPage.PageItems))
	}

	// Verify rich_text content block
	richTextItem := firstPage.PageItems[0]
	if richTextItem.Type != "rich_text" {
		t.Fatalf("expected type 'rich_text'; got '%s'", richTextItem.Type)
	}
	if richTextItem.Content != "This is a rich text content block from YAML" {
		t.Fatalf("expected rich_text content to be set; got '%s'", richTextItem.Content)
	}

	// Verify image content block
	imageItem := firstPage.PageItems[1]
	if imageItem.Type != "image" {
		t.Fatalf("expected type 'image'; got '%s'", imageItem.Type)
	}
	if imageItem.URL != "https://example.com/yaml-image.png" {
		t.Fatalf("expected image URL to be set; got '%s'", imageItem.URL)
	}
	if imageItem.AltText != "YAML test image" {
		t.Fatalf("expected image alt_text to be set; got '%s'", imageItem.AltText)
	}
	if imageItem.Caption != "An image from YAML config" {
		t.Fatalf("expected image caption to be set; got '%s'", imageItem.Caption)
	}

	// Verify second page - instructions with disable_dropdown
	secondPage := capturedPayload.CollectionItems[1]
	if len(secondPage.PageItems) != 2 {
		t.Fatalf("expected 2 page items in second page; got %d", len(secondPage.PageItems))
	}

	multipleChoiceItem := secondPage.PageItems[1]
	if multipleChoiceItem.Type != "multiple_choice" {
		t.Fatalf("expected type 'multiple_choice'; got '%s'", multipleChoiceItem.Type)
	}
	if multipleChoiceItem.DisableDropdown == nil {
		t.Fatal("expected disable_dropdown to be set; got nil")
	}
	if *multipleChoiceItem.DisableDropdown != true {
		t.Fatalf("expected disable_dropdown to be true; got %v", *multipleChoiceItem.DisableDropdown)
	}
	if len(multipleChoiceItem.Options) != 2 {
		t.Fatalf("expected 2 options; got %d", len(multipleChoiceItem.Options))
	}
}

func TestNewCreateCollectionCommandHandlesAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "collection.json")
	templateContent := fmt.Sprintf(`{
  "workspace_id": "6716028cd934ced9bac18658",
  "name": "test-collection",
  "task_details": %s,
  "collection_items": %s
}`, taskDetails, collectionItems)

	err := os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	c.EXPECT().
		CreateAITaskBuilderCollection(gomock.Any()).
		Return(nil, errors.New(collection.ErrWorkspaceNotFound))

	cmd := collection.NewCreateCollectionCommand(c, os.Stdout)
	cmd.SetArgs([]string{"-t", templateFile})
	err = cmd.Execute()

	if err == nil {
		t.Fatal("expected error; got nil")
	}

	if !strings.Contains(err.Error(), collection.ErrWorkspaceNotFound) {
		t.Fatalf("expected error to contain '%s'; got '%s'", collection.ErrWorkspaceNotFound, err.Error())
	}
}

func TestNewCreateCollectionCommandRequiresTemplatePath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := collection.NewCreateCollectionCommand(c, os.Stdout)
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error when template-path is missing")
	}

	expected := "required flag(s) \"template-path\" not set"
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
	}
}

func TestNewCreateCollectionCommandRequiresName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "collection.json")
	templateContent := fmt.Sprintf(`{
  "workspace_id": "6716028cd934ced9bac18658",
  "collection_items": %s
}`, collectionItems)

	err := os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	cmd := collection.NewCreateCollectionCommand(c, os.Stdout)
	cmd.SetArgs([]string{"-t", templateFile})
	err = cmd.Execute()

	if err == nil {
		t.Fatal("expected error when name is missing")
	}

	expected := collection.ErrNameRequired
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
	}
}

func TestNewCreateCollectionCommandRequiresWorkspaceID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "collection.json")
	templateContent := fmt.Sprintf(`{
  "name": "test-collection",
  "collection_items": %s
}`, collectionItems)

	err := os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	cmd := collection.NewCreateCollectionCommand(c, os.Stdout)
	cmd.SetArgs([]string{"-t", templateFile})
	err = cmd.Execute()

	if err == nil {
		t.Fatal("expected error when workspace_id is missing")
	}

	expected := collection.ErrWorkspaceIDRequired
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
	}
}

func TestNewCreateCollectionCommandRequiresItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "collection.json")
	templateContent := `{
  "name": "test-collection",
	"workspace_id": "6716028cd934ced9bac18658",
  "collection_items": []
}`

	err := os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	cmd := collection.NewCreateCollectionCommand(c, os.Stdout)
	cmd.SetArgs([]string{"-t", templateFile})
	err = cmd.Execute()

	if err == nil {
		t.Fatal("expected error when items is missing")
	}

	expected := collection.ErrCollectionItemsRequired
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
	}
}

func TestNewCreateCollectionCommandInvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "collection.json")
	invalidContent := `{invalid json`

	err := os.WriteFile(templateFile, []byte(invalidContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	cmd := collection.NewCreateCollectionCommand(c, os.Stdout)
	cmd.SetArgs([]string{"-t", templateFile})
	err = cmd.Execute()

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "unable to read config file") {
		t.Fatalf("expected error about config file; got %s", err.Error())
	}
}

func TestNewCreateCollectionCommandNonExistentFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := collection.NewCreateCollectionCommand(c, os.Stdout)
	cmd.SetArgs([]string{"-t", "/nonexistent/path/collection.json"})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	if !strings.Contains(err.Error(), "unable to read config file") {
		t.Fatalf("expected error about config file; got %s", err.Error())
	}
}

func TestNewCreateCollectionCommandWithContentBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "collection.json")
	templateContent := fmt.Sprintf(`{
  "workspace_id": "6716028cd934ced9bac18658",
  "name": "test-collection-with-content-blocks",
  "task_details": %s,
  "collection_items": %s
}`, taskDetails, collectionItemsWithContentBlocks)

	err := os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	response := client.CreateAITaskBuilderCollectionResponse{
		ID:            "collection-content-blocks-123",
		Name:          "test-collection-with-content-blocks",
		WorkspaceID:   "6716028cd934ced9bac18658",
		SchemaVersion: 1,
		CreatedBy:     "user-456",
		CollectionItems: []model.CollectionPage{
			{
				Order: 0,
				PageItems: []model.CollectionPageItem{
					{
						Order:   0,
						Type:    "rich_text",
						Content: "Welcome to this task. Please read the following instructions carefully.",
					},
					{
						Order:   1,
						Type:    "image",
						URL:     "https://example.com/image.png",
						AltText: "Example image",
						Caption: "This is an example image",
					},
				},
			},
			{
				Order: 1,
				PageItems: []model.CollectionPageItem{
					{
						Order:       0,
						Type:        "free_text",
						Description: "How was your experience completing this task?",
					},
				},
			},
		},
	}

	var capturedPayload model.CreateAITaskBuilderCollection
	c.EXPECT().
		CreateAITaskBuilderCollection(gomock.Any()).
		DoAndReturn(func(payload model.CreateAITaskBuilderCollection) (*client.CreateAITaskBuilderCollectionResponse, error) {
			capturedPayload = payload
			return &response, nil
		})

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := collection.NewCreateCollectionCommand(c, writer)
	cmd.SetArgs([]string{
		"-t", templateFile,
	})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error; got %s", err.Error())
	}

	writer.Flush()

	if len(capturedPayload.CollectionItems) != 2 {
		t.Fatalf("expected 2 collection items; got %d", len(capturedPayload.CollectionItems))
	}

	firstPage := capturedPayload.CollectionItems[0]
	if len(firstPage.PageItems) != 2 {
		t.Fatalf("expected 2 page items in first page; got %d", len(firstPage.PageItems))
	}

	richTextItem := firstPage.PageItems[0]
	if richTextItem.Type != "rich_text" {
		t.Fatalf("expected type 'rich_text'; got '%s'", richTextItem.Type)
	}
	if richTextItem.Content != "Welcome to this task. Please read the following instructions carefully." {
		t.Fatalf("expected rich_text content to be set; got '%s'", richTextItem.Content)
	}

	imageItem := firstPage.PageItems[1]
	if imageItem.Type != "image" {
		t.Fatalf("expected type 'image'; got '%s'", imageItem.Type)
	}
	if imageItem.URL != "https://example.com/image.png" {
		t.Fatalf("expected image URL to be set; got '%s'", imageItem.URL)
	}
	if imageItem.AltText != "Example image" {
		t.Fatalf("expected image alt_text to be set; got '%s'", imageItem.AltText)
	}
	if imageItem.Caption != "This is an example image" {
		t.Fatalf("expected image caption to be set; got '%s'", imageItem.Caption)
	}

	output := b.String()
	expectedStrings := []string{
		"Collection created successfully!",
		"ID:              collection-content-blocks-123",
		"Name:            test-collection-with-content-blocks",
		"Pages:           2",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain '%s'; got:\n%s", expected, output)
		}
	}
}

func TestNewCreateCollectionCommandWithTaskDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "collection.json")
	templateContent := fmt.Sprintf(`{
  "workspace_id": "6716028cd934ced9bac18658",
  "name": "test-collection-with-task-details",
  "task_details": {
    "task_name": "Quality Assessment Task",
    "task_introduction": "Welcome to this quality assessment task. Please follow the instructions carefully.",
    "task_steps": "1. Read the content on each page\n2. Answer the questions thoughtfully\n3. Submit your responses"
  },
  "collection_items": %s
}`, collectionItems)

	err := os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	response := client.CreateAITaskBuilderCollectionResponse{
		ID:            "collection-task-details-123",
		Name:          "test-collection-with-task-details",
		WorkspaceID:   "6716028cd934ced9bac18658",
		SchemaVersion: 1,
		CreatedBy:     "user-456",
		TaskDetails: &model.TaskDetails{
			TaskName:         "Quality Assessment Task",
			TaskIntroduction: "Welcome to this quality assessment task. Please follow the instructions carefully.",
			TaskSteps:        "1. Read the content on each page\n2. Answer the questions thoughtfully\n3. Submit your responses",
		},
		CollectionItems: []model.CollectionPage{},
	}

	var capturedPayload model.CreateAITaskBuilderCollection
	c.EXPECT().
		CreateAITaskBuilderCollection(gomock.Any()).
		DoAndReturn(func(payload model.CreateAITaskBuilderCollection) (*client.CreateAITaskBuilderCollectionResponse, error) {
			capturedPayload = payload
			return &response, nil
		})

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := collection.NewCreateCollectionCommand(c, writer)
	cmd.SetArgs([]string{
		"-t", templateFile,
	})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error; got %s", err.Error())
	}

	writer.Flush()

	if capturedPayload.TaskDetails == nil {
		t.Fatal("expected task_details to be set; got nil")
	}

	if capturedPayload.TaskDetails.TaskName != "Quality Assessment Task" {
		t.Fatalf("expected task_name 'Quality Assessment Task'; got '%s'", capturedPayload.TaskDetails.TaskName)
	}

	if capturedPayload.TaskDetails.TaskIntroduction != "Welcome to this quality assessment task. Please follow the instructions carefully." {
		t.Fatalf("expected task_introduction to be set; got '%s'", capturedPayload.TaskDetails.TaskIntroduction)
	}

	if capturedPayload.TaskDetails.TaskSteps != "1. Read the content on each page\n2. Answer the questions thoughtfully\n3. Submit your responses" {
		t.Fatalf("expected task_steps to be set; got '%s'", capturedPayload.TaskDetails.TaskSteps)
	}
}

func TestNewCreateCollectionCommandRequiresTaskDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "collection.json")
	templateContent := fmt.Sprintf(`{
  "workspace_id": "6716028cd934ced9bac18658",
  "name": "test-collection",
  "collection_items": %s
}`, collectionItems)

	err := os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	cmd := collection.NewCreateCollectionCommand(c, os.Stdout)
	cmd.SetArgs([]string{"-t", templateFile})
	err = cmd.Execute()

	if err == nil {
		t.Fatal("expected error when task_details is missing")
	}

	expected := collection.ErrTaskDetailsRequired
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
	}
}

func TestNewCreateCollectionCommandRequiresTaskName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "collection.json")
	templateContent := fmt.Sprintf(`{
  "workspace_id": "6716028cd934ced9bac18658",
  "name": "test-collection",
  "task_details": {
    "task_introduction": "Introduction",
    "task_steps": "Steps"
  },
  "collection_items": %s
}`, collectionItems)

	err := os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	cmd := collection.NewCreateCollectionCommand(c, os.Stdout)
	cmd.SetArgs([]string{"-t", templateFile})
	err = cmd.Execute()

	if err == nil {
		t.Fatal("expected error when task_name is missing")
	}

	expected := collection.ErrTaskNameRequired
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
	}
}

func TestNewCreateCollectionCommandRequiresTaskIntroduction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "collection.json")
	templateContent := fmt.Sprintf(`{
  "workspace_id": "6716028cd934ced9bac18658",
  "name": "test-collection",
  "task_details": {
    "task_name": "Task Name",
    "task_steps": "Steps"
  },
  "collection_items": %s
}`, collectionItems)

	err := os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	cmd := collection.NewCreateCollectionCommand(c, os.Stdout)
	cmd.SetArgs([]string{"-t", templateFile})
	err = cmd.Execute()

	if err == nil {
		t.Fatal("expected error when task_introduction is missing")
	}

	expected := collection.ErrTaskIntroductionRequired
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
	}
}

func TestNewCreateCollectionCommandRequiresTaskSteps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "collection.json")
	templateContent := fmt.Sprintf(`{
  "workspace_id": "6716028cd934ced9bac18658",
  "name": "test-collection",
  "task_details": {
    "task_name": "Task Name",
    "task_introduction": "Introduction"
  },
  "collection_items": %s
}`, collectionItems)

	err := os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	cmd := collection.NewCreateCollectionCommand(c, os.Stdout)
	cmd.SetArgs([]string{"-t", templateFile})
	err = cmd.Execute()

	if err == nil {
		t.Fatal("expected error when task_steps is missing")
	}

	expected := collection.ErrTaskStepsRequired
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
	}
}
