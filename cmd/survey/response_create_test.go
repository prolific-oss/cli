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

func TestNewResponseCreateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := survey.NewResponseCreateCommand("create", c, os.Stdout)

	if cmd.Use != "create <survey_id>" {
		t.Fatalf("expected use: create <survey_id>; got %s", cmd.Use)
	}

	if cmd.Short != "Create a survey response" {
		t.Fatalf("expected short: Create a survey response; got %s", cmd.Short)
	}
}

func writeResponseTemplateFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "response.json")
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		t.Fatalf("unable to write template file: %s", err)
	}
	return path
}

func TestCreateSurveyResponse(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		setupMocks     func(c *mock_client.MockAPI)
		expectedOutput string
		expectedError  string
	}{
		{
			name: "successful creation with questions",
			template: `{
				"participant_id": "part-001",
				"submission_id": "sub-001",
				"questions": [
					{
						"question_id": "q-001",
						"question_title": "What is your handedness?",
						"answers": [{"answer_id": "a-001", "value": "Left"}]
					}
				]
			}`,
			setupMocks: func(c *mock_client.MockAPI) {
				c.EXPECT().CreateSurveyResponse(gomock.Eq(testSurveyID), gomock.Any()).
					DoAndReturn(func(surveyID string, r model.CreateSurveyResponseRequest) (*client.CreateSurveyResponseResponse, error) {
						if r.ParticipantID != "part-001" {
							return nil, fmt.Errorf("expected participant_id 'part-001'; got '%s'", r.ParticipantID)
						}
						if r.SubmissionID != "sub-001" {
							return nil, fmt.Errorf("expected submission_id 'sub-001'; got '%s'", r.SubmissionID)
						}
						return &client.CreateSurveyResponseResponse{
							SurveyResponse: model.SurveyResponse{ID: "resp-001"},
						}, nil
					}).Times(1)
			},
			expectedOutput: "Created survey response: resp-001\n",
		},
		{
			name: "successful creation with sections",
			template: `{
				"participant_id": "part-002",
				"submission_id": "sub-002",
				"sections": [
					{
						"section_id": "sec-001",
						"questions": [
							{
								"question_id": "q-001",
								"question_title": "Age?",
								"answers": [{"answer_id": "a-001", "value": "26-35"}]
							}
						]
					}
				]
			}`,
			setupMocks: func(c *mock_client.MockAPI) {
				c.EXPECT().CreateSurveyResponse(gomock.Eq(testSurveyID), gomock.Any()).
					Return(&client.CreateSurveyResponseResponse{
						SurveyResponse: model.SurveyResponse{ID: "resp-002"},
					}, nil).Times(1)
			},
			expectedOutput: "Created survey response: resp-002\n",
		},
		{
			name: "API error on create",
			template: `{
				"participant_id": "part-001",
				"submission_id": "sub-001",
				"questions": [
					{"question_id": "q-001", "question_title": "Q1", "answers": [{"answer_id": "a-001", "value": "Yes"}]}
				]
			}`,
			setupMocks: func(c *mock_client.MockAPI) {
				c.EXPECT().CreateSurveyResponse(gomock.Eq(testSurveyID), gomock.Any()).
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

			templatePath := writeResponseTemplateFile(t, tt.template)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := survey.NewResponseCreateCommand("create", c, writer)
			cmd.SetArgs([]string{testSurveyID, "-t", templatePath})
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

func TestCreateSurveyResponseMissingTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := survey.NewResponseCreateCommand("create", c, os.Stdout)
	cmd.SetArgs([]string{"some-survey-id"})
	err := cmd.Execute()

	expected := "error: a template file is required, use -t to specify the path"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%v'\n", expected, err)
	}
}
