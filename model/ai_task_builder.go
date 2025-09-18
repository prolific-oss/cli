package model

import (
	"time"
)

// AITaskBuilderBatch represents an AI Task Builder batch.
type AITaskBuilderBatch struct {
	ID                    string      `json:"id"`
	CreatedAt             time.Time   `json:"created_at"`
	CreatedBy             string      `json:"created_by"`
	Datasets              []Dataset   `json:"datasets"`
	Name                  string      `json:"name"`
	Status                string      `json:"status"`
	TasksPerGroup         int         `json:"tasks_per_group"`
	TotalTaskCount        int         `json:"total_task_count"`
	TotalInstructionCount int         `json:"total_instruction_count"`
	WorkspaceID           string      `json:"workspace_id"`
	SchemaVersion         int         `json:"schema_version"`
	TaskDetails           TaskDetails `json:"task_details"`
}

// AITaskBuilderBatchStatus represents the status of an AI Task Builder batch.
type AITaskBuilderBatchStatus struct {
	Status AITaskBuilderBatchStatusEnum `json:"status"`
}

// Dataset represents a dataset in a batch.
type Dataset struct {
	ID                  string `json:"id"`
	TotalDatapointCount int    `json:"total_datapoint_count"`
}

// TaskDetails represents the task configuration details.
type TaskDetails struct {
	TaskName         string `json:"task_name"`
	TaskIntroduction string `json:"task_introduction"`
	TaskSteps        string `json:"task_steps"`
}

// AITaskBuilderResponse represents a response from an AI Task Builder batch task.
type AITaskBuilderResponse struct {
	ID            string                    `json:"id"`
	BatchID       string                    `json:"batch_id"`
	ParticipantID string                    `json:"participant_id"`
	TaskID        string                    `json:"task_id"`
	CorrelationID string                    `json:"correlation_id"`
	SubmissionID  string                    `json:"submission_id"`
	Metadata      map[string]string         `json:"metadata"`
	Response      AITaskBuilderResponseData `json:"response"`
	CreatedAt     time.Time                 `json:"created_at"`
	SchemaVersion int                       `json:"schema_version"`
}

// AITaskBuilderResponseData represents the response data structure.
// This is a discriminated union based on the Type field.
type AITaskBuilderResponseData struct {
	InstructionID string                      `json:"instruction_id"`
	Type          AITaskBuilderResponseType   `json:"type"`
	Text          *string                     `json:"text,omitempty"`   // For free_text and multiple_choice_with_free_text
	Answer        []AITaskBuilderAnswerOption `json:"answer,omitempty"` // For multiple_choice and multiple_choice_with_free_text
}

// AITaskBuilderResponseType represents the type of response.
type AITaskBuilderResponseType string

const (
	AITaskBuilderResponseTypeFreeText                   AITaskBuilderResponseType = "free_text"
	AITaskBuilderResponseTypeMultipleChoice             AITaskBuilderResponseType = "multiple_choice"
	AITaskBuilderResponseTypeMultipleChoiceWithFreeText AITaskBuilderResponseType = "multiple_choice_with_free_text"
)

// AITaskBuilderAnswerOption represents an answer option for multiple choice responses.
type AITaskBuilderAnswerOption struct {
	Value string `json:"value"`
}

type AITaskBuilderBatchStatusEnum string

const (
	// UNINITIALISED: the batch has been created, but contains no tasks.
	AITaskBuilderBatchStatusUninitialised AITaskBuilderBatchStatusEnum = "UNINITIALISED"
	// PROCESSING: The batch is being processed into tasks.
	AITaskBuilderBatchStatusProcessing AITaskBuilderBatchStatusEnum = "PROCESSING"
	// READY: The batch is processed and ready to be attached to a Prolific study.
	AITaskBuilderBatchStatusReady AITaskBuilderBatchStatusEnum = "READY"
	// ERROR: The batch has encountered an error and the data may not be usable.
	AITaskBuilderBatchStatusError AITaskBuilderBatchStatusEnum = "ERROR"
)
