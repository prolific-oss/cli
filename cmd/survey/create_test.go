package survey_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/survey"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewCreateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := survey.NewCreateCommand("create", c, os.Stdout)

	if cmd.Use != "create" {
		t.Fatalf("expected use: create; got %s", cmd.Use)
	}

	if cmd.Short != "Create a survey" {
		t.Fatalf("expected short: Create a survey; got %s", cmd.Short)
	}
}

func writeSurveyTemplateFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "survey.json")
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		t.Fatalf("unable to write template file: %s", err)
	}
	return path
}

func TestCreateSurvey(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		extraArgs      []string
		setupMocks     func(c *mock_client.MockAPI)
		expectedOutput string
		expectedError  string
	}{
		{
			name: "successful creation with auto researcher ID",
			template: `{
				"title": "Screening Survey",
				"questions": [
					{
						"title": "What is your handedness?",
						"type": "single",
						"answers": [{"value": "Left"}, {"value": "Right"}]
					}
				]
			}`,
			setupMocks: func(c *mock_client.MockAPI) {
				c.EXPECT().GetMe().Return(&client.MeResponse{ID: testResearcherID}, nil).Times(1)
				c.EXPECT().CreateSurvey(gomock.Any()).
					DoAndReturn(func(s model.CreateSurvey) (*client.CreateSurveyResponse, error) {
						if s.Title != "Screening Survey" {
							return nil, fmt.Errorf("expected title 'Screening Survey'; got '%s'", s.Title)
						}
						if s.ResearcherID != testResearcherID {
							return nil, fmt.Errorf("expected researcher_id '%s'; got '%s'", testResearcherID, s.ResearcherID)
						}
						return &client.CreateSurveyResponse{
							Survey: model.Survey{ID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
						}, nil
					}).Times(1)
			},
			expectedOutput: "Created survey: 6ba7b810-9dad-11d1-80b4-00c04fd430c8\n",
		},
		{
			name:      "title override via flag",
			extraArgs: []string{"--title", "Overridden Title"},
			template: `{
				"title": "Original Title",
				"questions": [
					{"title": "Q1", "type": "single", "answers": [{"value": "Yes"}, {"value": "No"}]}
				]
			}`,
			setupMocks: func(c *mock_client.MockAPI) {
				c.EXPECT().GetMe().Return(&client.MeResponse{ID: testResearcherID}, nil).Times(1)
				c.EXPECT().CreateSurvey(gomock.Any()).
					DoAndReturn(func(s model.CreateSurvey) (*client.CreateSurveyResponse, error) {
						if s.Title != "Overridden Title" {
							return nil, fmt.Errorf("expected title 'Overridden Title'; got '%s'", s.Title)
						}
						return &client.CreateSurveyResponse{
							Survey: model.Survey{ID: "7ca8c921-0ebe-22e2-91c5-11d15ge541d9"},
						}, nil
					}).Times(1)
			},
			expectedOutput: "Created survey: 7ca8c921-0ebe-22e2-91c5-11d15ge541d9\n",
		},
		{
			name: "researcher ID from template skips GetMe",
			template: `{
				"researcher_id": "661f9511-f3ac-52f5-b7g9-55h59kh985hd",
				"title": "Pre-filled Researcher",
				"questions": [
					{"title": "Q1", "type": "single", "answers": [{"value": "Yes"}, {"value": "No"}]}
				]
			}`,
			setupMocks: func(c *mock_client.MockAPI) {
				c.EXPECT().CreateSurvey(gomock.Any()).
					DoAndReturn(func(s model.CreateSurvey) (*client.CreateSurveyResponse, error) {
						if s.ResearcherID != "661f9511-f3ac-52f5-b7g9-55h59kh985hd" {
							return nil, fmt.Errorf("expected researcher_id from template; got '%s'", s.ResearcherID)
						}
						return &client.CreateSurveyResponse{
							Survey: model.Survey{ID: "8db9d032-1fcf-33f3-a2d6-22e26hf652ea"},
						}, nil
					}).Times(1)
			},
			expectedOutput: "Created survey: 8db9d032-1fcf-33f3-a2d6-22e26hf652ea\n",
		},
		{
			name: "API error on create",
			template: `{
				"title": "Test",
				"questions": [
					{"title": "Q1", "type": "single", "answers": [{"value": "Yes"}]}
				]
			}`,
			setupMocks: func(c *mock_client.MockAPI) {
				c.EXPECT().GetMe().Return(&client.MeResponse{ID: testResearcherID}, nil).Times(1)
				c.EXPECT().CreateSurvey(gomock.Any()).
					Return(nil, errors.New("API error: bad request")).Times(1)
			},
			expectedError: "error: API error: bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			tt.setupMocks(c)

			templatePath := writeSurveyTemplateFile(t, tt.template)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			args := append([]string{"-t", templatePath}, tt.extraArgs...)
			cmd := survey.NewCreateCommand("create", c, writer)
			cmd.SetArgs(args)
			err := cmd.Execute()

			writer.Flush()

			if tt.expectedError != "" {
				if err == nil || err.Error() != tt.expectedError {
					t.Fatalf("expected error '%s'; got '%v'", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			actual := b.String()
			if actual != tt.expectedOutput {
				t.Fatalf("expected\n'%s'\ngot\n'%s'\n", tt.expectedOutput, actual)
			}
		})
	}
}

func TestCreateSurveyMissingTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := survey.NewCreateCommand("create", c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	expected := "error: a template file is required, use -t to specify the path"
	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
