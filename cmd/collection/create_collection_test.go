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

const items = `[
    {
      "order": 0,
      "items": [
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
	"items": %s
}`, items)

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
		Items: []model.CollectionPage{
			{
				Order: 0,
				Items: []model.CollectionInstruction{
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
items:
  - order: 0
    items:
      - order: 0
        type: free_text
        description: YAML test description
`

	err := os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	// Expected response
	response := client.CreateAITaskBuilderCollectionResponse{
		ID:            "collection-yaml-123",
		Name:          "yaml-test-collection",
		WorkspaceID:   "6716028cd934ced9bac18658",
		SchemaVersion: 1,
		CreatedBy:     "user-789",
		Items:         []model.CollectionPage{},
	}

	c.EXPECT().
		CreateAITaskBuilderCollection(gomock.Any()).
		Return(&response, nil)

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
  "items": %s
}`, items)

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
  "items": %s
}`, items)

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
  "items": %s
}`, items)

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
  "items": []
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
