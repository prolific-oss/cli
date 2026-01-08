package collection_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/collection"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewUpdateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := collection.NewUpdateCommand(c, os.Stdout)

	use := "update <collection-id>"
	short := "Update a collection"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestUpdateCollection(t *testing.T) {
	collectionID := "550e8400-e29b-41d4-a716-446655440000"
	pageID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	tests := []struct {
		name            string
		args            []string
		configContent   string
		configExt       string
		expectedPayload model.UpdateCollection
		mockReturn      *model.Collection
		mockError       error
		expectedOutput  string
		expectedError   string
		skipMock        bool
	}{
		{
			name:      "successful update with name and empty items",
			args:      []string{collectionID},
			configExt: ".yaml",
			configContent: `name: Updated Collection Name
items: []
`,
			expectedPayload: model.UpdateCollection{
				Name:  "Updated Collection Name",
				Items: []model.Page{},
			},
			mockReturn: &model.Collection{
				ID:        collectionID,
				Name:      "Updated Collection Name",
				CreatedAt: time.Now(),
				CreatedBy: "user123",
				ItemCount: 0,
			},
			mockError: nil,
			expectedOutput: `Collection updated successfully
ID: 550e8400-e29b-41d4-a716-446655440000
Name: Updated Collection Name
`,
			expectedError: "",
		},
		{
			name:      "successful update with pages and instructions",
			args:      []string{collectionID},
			configExt: ".yaml",
			configContent: `name: Collection With Pages
items:
  - id: "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
    order: 0
    items:
      - type: free_text
        description: "What is your name?"
        order: 0
  - order: 1
    items:
      - type: multiple_choice
        description: "How satisfied are you?"
        order: 0
        answer_limit: 1
        options:
          - label: Very Satisfied
            value: "5"
          - label: Satisfied
            value: "4"
`,
			expectedPayload: model.UpdateCollection{
				Name: "Collection With Pages",
				Items: []model.Page{
					{
						BaseEntity: model.BaseEntity{ID: pageID},
						Order:      0,
						Items: []model.PageInstruction{
							{
								Type:        model.InstructionTypeFreeText,
								Description: "What is your name?",
								Order:       0,
							},
						},
					},
					{
						Order: 1,
						Items: []model.PageInstruction{
							{
								Type:        model.InstructionTypeMultipleChoice,
								Description: "How satisfied are you?",
								Order:       0,
								AnswerLimit: 1,
								Options: []model.MultipleChoiceOption{
									{Label: "Very Satisfied", Value: "5"},
									{Label: "Satisfied", Value: "4"},
								},
							},
						},
					},
				},
			},
			mockReturn: &model.Collection{
				ID:        collectionID,
				Name:      "Collection With Pages",
				CreatedAt: time.Now(),
				CreatedBy: "user123",
				ItemCount: 2,
			},
			mockError: nil,
			expectedOutput: `Collection updated successfully
ID: 550e8400-e29b-41d4-a716-446655440000
Name: Collection With Pages
`,
			expectedError: "",
		},
		{
			name:      "successful update with JSON config",
			args:      []string{collectionID},
			configExt: ".json",
			configContent: `{
  "name": "JSON Updated Collection",
  "items": [
    {
      "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "order": 0,
      "items": [
        {
          "type": "free_text",
          "description": "Enter your feedback",
          "order": 0
        }
      ]
    }
  ]
}`,
			expectedPayload: model.UpdateCollection{
				Name: "JSON Updated Collection",
				Items: []model.Page{
					{
						BaseEntity: model.BaseEntity{ID: pageID},
						Order:      0,
						Items: []model.PageInstruction{
							{
								Type:        model.InstructionTypeFreeText,
								Description: "Enter your feedback",
								Order:       0,
							},
						},
					},
				},
			},
			mockReturn: &model.Collection{
				ID:        collectionID,
				Name:      "JSON Updated Collection",
				CreatedAt: time.Now(),
				CreatedBy: "user123",
				ItemCount: 1,
			},
			mockError: nil,
			expectedOutput: `Collection updated successfully
ID: 550e8400-e29b-41d4-a716-446655440000
Name: JSON Updated Collection
`,
			expectedError: "",
		},
		{
			name:           "missing collection ID",
			args:           []string{},
			configContent:  "",
			mockReturn:     nil,
			mockError:      nil,
			expectedOutput: "",
			expectedError:  "accepts 1 arg(s), received 0",
			skipMock:       true,
		},
		{
			name:      "missing name in config",
			args:      []string{collectionID},
			configExt: ".yaml",
			configContent: `items: []
`,
			mockReturn:     nil,
			mockError:      nil,
			expectedOutput: "",
			expectedError:  "error: name is required",
			skipMock:       true,
		},
		{
			name:      "missing items in config",
			args:      []string{collectionID},
			configExt: ".yaml",
			configContent: `name: Test Collection
`,
			mockReturn:     nil,
			mockError:      nil,
			expectedOutput: "",
			expectedError:  "error: at least one item must be provided",
			skipMock:       true,
		},
		{
			name:           "empty config file",
			args:           []string{collectionID},
			configExt:      ".yaml",
			configContent:  ``,
			mockReturn:     nil,
			mockError:      nil,
			expectedOutput: "",
			expectedError:  "error: name is required",
			skipMock:       true,
		},
		{
			name:      "api error - not found",
			args:      []string{collectionID},
			configExt: ".yaml",
			configContent: `name: Test Collection
items: []
`,
			expectedPayload: model.UpdateCollection{
				Name:  "Test Collection",
				Items: []model.Page{},
			},
			mockReturn:     nil,
			mockError:      errors.New("request failed with status 404: collection not found"),
			expectedOutput: "",
			expectedError:  "request failed with status 404: collection not found",
		},
		{
			name:      "api error - service unavailable",
			args:      []string{collectionID},
			configExt: ".yaml",
			configContent: `name: Test Collection
items: []
`,
			expectedPayload: model.UpdateCollection{
				Name:  "Test Collection",
				Items: []model.Page{},
			},
			mockReturn:     nil,
			mockError:      errors.New("request failed with status 502: service unavailable"),
			expectedOutput: "",
			expectedError:  "request failed with status 502: service unavailable",
		},
		{
			name:      "unknown field in YAML config",
			args:      []string{collectionID},
			configExt: ".yaml",
			configContent: `name: Test Collection
items: []
unknown_field: "should cause error"
`,
			mockReturn:     nil,
			mockError:      nil,
			expectedOutput: "",
			expectedError:  "error: unable to unmarshal config file: decoding failed due to the following error(s):\n\n'' has invalid keys: unknown_field",
			skipMock:       true,
		},
		{
			name:      "unknown field in JSON config",
			args:      []string{collectionID},
			configExt: ".json",
			configContent: `{
  "name": "Test Collection",
  "items": [],
  "unknown_field": "should cause error"
}`,
			mockReturn:     nil,
			mockError:      nil,
			expectedOutput: "",
			expectedError:  "error: unable to unmarshal config file: decoding failed due to the following error(s):\n\n'' has invalid keys: unknown_field",
			skipMock:       true,
		},
		{
			name:      "unsupported config file format",
			args:      []string{collectionID},
			configExt: ".txt",
			configContent: `name: Test Collection
items: []
`,
			mockReturn:     nil,
			mockError:      nil,
			expectedOutput: "",
			expectedError:  `error: unable to read config file: Unsupported Config Type "txt"`,
			skipMock:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			var configFile string
			if tt.configExt != "" {
				configFile = createTempConfigFile(t, tt.configContent, tt.configExt)
			}

			// Only expect API call if we have valid args and config
			if !tt.skipMock && len(tt.args) > 0 && tt.configExt != "" {
				c.EXPECT().
					UpdateCollection(tt.args[0], gomock.Eq(tt.expectedPayload)).
					Return(tt.mockReturn, tt.mockError).
					Times(1)
			}

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := collection.NewUpdateCommand(c, writer)

			var err error
			if len(tt.args) == 0 {
				cmd.SetArgs(tt.args)
				err = cmd.Execute()
			} else if tt.configExt != "" {
				cmd.SetArgs(append(tt.args, "-t", configFile))
				err = cmd.Execute()
			} else {
				cmd.SetArgs(tt.args)
				err = cmd.Execute()
			}
			writer.Flush()

			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error '%s', got nil", tt.expectedError)
				}
				if err.Error() != tt.expectedError {
					t.Fatalf("expected error '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			actual := b.String()
			if actual != tt.expectedOutput {
				t.Fatalf("expected output:\n'%s'\n\ngot:\n'%s'", tt.expectedOutput, actual)
			}
		})
	}
}

func TestUpdateCollectionMissingConfigFlag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := collection.NewUpdateCommand(c, writer)
	cmd.SetArgs([]string{"collection-id-123"})
	err := cmd.Execute()
	writer.Flush()

	expectedError := `required flag(s) "template-path" not set`
	if err == nil {
		t.Fatalf("expected error '%s', got nil", expectedError)
	}
	if err.Error() != expectedError {
		t.Fatalf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestUpdateCollectionInvalidConfigFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := collection.NewUpdateCommand(c, writer)
	cmd.SetArgs([]string{"collection-id-123", "-t", "/nonexistent/path/config.yaml"})
	err := cmd.Execute()
	writer.Flush()

	if err == nil {
		t.Fatal("expected error for nonexistent config file, got nil")
	}
}

func TestUpdateCollectionExactPayload(t *testing.T) {
	collectionID := "550e8400-e29b-41d4-a716-446655440000"
	pageID := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	configContent := `{
  "name": "Exact Payload Test",
  "items": [
    {
      "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "order": 0,
      "items": [
        {
          "type": "free_text",
          "description": "Enter your feedback",
          "order": 0,
          "placeholder_text_input": "Type here..."
        },
        {
          "type": "multiple_choice",
          "description": "Rate your experience",
          "order": 1,
          "answer_limit": 1,
          "options": [
            {"label": "Good", "value": "good"},
            {"label": "Bad", "value": "bad"}
          ]
        }
      ]
    }
  ]
}`

	expectedPayload := model.UpdateCollection{
		Name: "Exact Payload Test",
		Items: []model.Page{
			{
				BaseEntity: model.BaseEntity{
					ID: pageID,
				},
				Order: 0,
				Items: []model.PageInstruction{
					{
						Type:                 model.InstructionTypeFreeText,
						Description:          "Enter your feedback",
						Order:                0,
						PlaceholderTextInput: "Type here...",
					},
					{
						Type:        model.InstructionTypeMultipleChoice,
						Description: "Rate your experience",
						Order:       1,
						AnswerLimit: 1,
						Options: []model.MultipleChoiceOption{
							{Label: "Good", Value: "good"},
							{Label: "Bad", Value: "bad"},
						},
					},
				},
			},
		},
	}

	c.EXPECT().
		UpdateCollection(collectionID, gomock.Eq(expectedPayload)).
		Return(&model.Collection{
			ID:        collectionID,
			Name:      "Exact Payload Test",
			CreatedAt: time.Now(),
			CreatedBy: "user123",
			ItemCount: 1,
		}, nil).
		Times(1)

	configFile := createTempConfigFile(t, configContent, ".json")

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := collection.NewUpdateCommand(c, writer)
	cmd.SetArgs([]string{collectionID, "-t", configFile})
	err := cmd.Execute()
	writer.Flush()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedOutput := `Collection updated successfully
ID: 550e8400-e29b-41d4-a716-446655440000
Name: Exact Payload Test
`
	actual := b.String()
	if actual != expectedOutput {
		t.Fatalf("expected output:\n'%s'\n\ngot:\n'%s'", expectedOutput, actual)
	}
}

func createTempConfigFile(t *testing.T, content string, ext string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "collection-config-*"+ext)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	return tmpFile.Name()
}
