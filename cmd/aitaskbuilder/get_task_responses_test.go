package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewGetResponsesCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewGetResponsesCommand(c, os.Stdout)

	use := "responses"
	short := "List batch task responses"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewGetResponsesCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "5cf3ea63-3980-4149-9ea9-bea243489cc8"

	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	textResponse := "test response"
	response := client.GetAITaskBuilderResponsesResponse{
		Results: []model.AITaskBuilderResponse{
			{
				ID:            "response-123",
				BatchID:       batchID,
				ParticipantID: "participant-456",
				TaskID:        "task-456",
				CorrelationID: "correlation-001",
				SubmissionID:  "submission-001",
				Metadata: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				Response: model.AITaskBuilderResponseData{
					InstructionID: "instruction-001",
					Type:          model.AITaskBuilderResponseTypeFreeText,
					Text:          &textResponse,
				},
				CreatedAt:     createdAt,
				SchemaVersion: 2,
			},
			{
				ID:            "response-789",
				BatchID:       batchID,
				ParticipantID: "participant-789",
				TaskID:        "task-101",
				CorrelationID: "correlation-002",
				SubmissionID:  "submission-002",
				Metadata:      map[string]string{},
				Response: model.AITaskBuilderResponseData{
					InstructionID: "instruction-002",
					Type:          model.AITaskBuilderResponseTypeMultipleChoice,
					Answer: []model.AITaskBuilderAnswerOption{
						{Value: "option1"},
						{Value: "option2"},
					},
				},
				CreatedAt:     createdAt,
				SchemaVersion: 2,
			},
		},
		Meta: client.ResponseMeta{
			Count: 2,
		},
	}

	c.
		EXPECT().
		GetAITaskBuilderResponses(gomock.Eq(batchID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewGetResponsesCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", batchID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Batch Responses:
Batch ID: 5cf3ea63-3980-4149-9ea9-bea243489cc8
Total Responses: 2

Response 1:
  ID: response-123
  Batch ID: 5cf3ea63-3980-4149-9ea9-bea243489cc8
  Participant ID: participant-456
  Task ID: task-456
  Correlation ID: correlation-001
  Submission ID: submission-001
  Schema Version: 2
  Created At: 2024-01-01 12:00:00
  Metadata:
    key1: value1
    key2: value2
  Response:
    Instruction ID: instruction-001
    Type: free_text
    Text: test response

Response 2:
  ID: response-789
  Batch ID: 5cf3ea63-3980-4149-9ea9-bea243489cc8
  Participant ID: participant-789
  Task ID: task-101
  Correlation ID: correlation-002
  Submission ID: submission-002
  Schema Version: 2
  Created At: 2024-01-01 12:00:00
  Response:
    Instruction ID: instruction-002
    Type: multiple_choice
    Selected Options:
      - option1
      - option2
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewGetResponsesCommandHandlesEmptyResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "5d883286-9480-463a-a738-9ddcfae65b8b"

	response := client.GetAITaskBuilderResponsesResponse{
		Results: []model.AITaskBuilderResponse{},
		Meta: client.ResponseMeta{
			Count: 0,
		},
	}

	c.
		EXPECT().
		GetAITaskBuilderResponses(gomock.Eq(batchID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewGetResponsesCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", batchID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Batch Responses:
Batch ID: 5d883286-9480-463a-a738-9ddcfae65b8b
Total Responses: 0

No responses found for batch 5d883286-9480-463a-a738-9ddcfae65b8b
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewGetResponsesCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "big-invalid-batch-id"
	errorMessage := aitaskbuilder.ErrBatchNotFound

	c.
		EXPECT().
		GetAITaskBuilderResponses(gomock.Eq(batchID)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := aitaskbuilder.NewGetResponsesCommand(c, os.Stdout)
	_ = cmd.Flags().Set("batch-id", batchID)
	err := cmd.RunE(cmd, nil)

	expected := errorMessage

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewGetResponsesCommandRequiresBatchID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewGetResponsesCommand(c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Fatal("expected error when batch-id is missing")
	}

	if !cmd.Flags().Changed("batch-id") {
		expected := aitaskbuilder.ErrBatchIDRequired
		if err.Error() != ""+expected {
			t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
		}
	}
}

func TestNewGetResponsesCommandHandlesResponseWithEmptyAnswer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "e0d498d3-09cc-4f11-b3ed-99b6753b0a2c"
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	response := client.GetAITaskBuilderResponsesResponse{
		Results: []model.AITaskBuilderResponse{
			{
				ID:            "response-123",
				BatchID:       batchID,
				ParticipantID: "participant-456",
				TaskID:        "task-456",
				CorrelationID: "correlation-001",
				SubmissionID:  "submission-001",
				Metadata:      map[string]string{},
				Response: model.AITaskBuilderResponseData{
					InstructionID: "instruction-001",
					Type:          model.AITaskBuilderResponseTypeFreeText,
					Text:          nil, // empty answer
				},
				CreatedAt:     createdAt,
				SchemaVersion: 2,
			},
		},
		Meta: client.ResponseMeta{
			Count: 1,
		},
	}

	c.
		EXPECT().
		GetAITaskBuilderResponses(gomock.Eq(batchID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewGetResponsesCommand(c, writer)
	_ = cmd.Flags().Set("batch-id", batchID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `AI Task Builder Batch Responses:
Batch ID: e0d498d3-09cc-4f11-b3ed-99b6753b0a2c
Total Responses: 1

Response 1:
  ID: response-123
  Batch ID: e0d498d3-09cc-4f11-b3ed-99b6753b0a2c
  Participant ID: participant-456
  Task ID: task-456
  Correlation ID: correlation-001
  Submission ID: submission-001
  Schema Version: 2
  Created At: 2024-01-01 12:00:00
  Response:
    Instruction ID: instruction-001
    Type: free_text
    Text: 
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}
