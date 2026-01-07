package model

import (
	"time"
)

// AITaskBuilderBatch represents an AI Task Builder batch.
type AITaskBuilderBatch struct {
	ID                    string                       `json:"id"`
	CreatedAt             time.Time                    `json:"created_at"`
	CreatedBy             string                       `json:"created_by"`
	Datasets              []Dataset                    `json:"datasets"`
	Name                  string                       `json:"name"`
	Status                AITaskBuilderBatchStatusEnum `json:"status"`
	TasksPerGroup         int                          `json:"tasks_per_group"`
	TotalTaskCount        int                          `json:"total_task_count"`
	TotalInstructionCount int                          `json:"total_instruction_count"`
	WorkspaceID           string                       `json:"workspace_id"`
	SchemaVersion         int                          `json:"schema_version"`
	TaskDetails           TaskDetails                  `json:"task_details"`
}

// AITaskBuilderBatchStatus represents the status of an AI Task Builder batch.
type AITaskBuilderBatchStatus struct {
	Status AITaskBuilderBatchStatusEnum `json:"status"`
}

// Dataset represents a dataset in a batch.
type Dataset struct {
	ID                  string        `json:"id"`
	Name                string        `json:"name"`
	CreatedAt           string        `json:"created_at"`
	CreatedBy           string        `json:"created_by"`
	Status              DatasetStatus `json:"status"`
	TotalDatapointCount int           `json:"total_datapoint_count"`
	WorkspaceID         string        `json:"workspace_id"`
}

// DatasetStatus represents the status of a dataset.
type DatasetStatus string

const (
	// DatasetStatusUninitialised means the dataset has been created but no data has been uploaded.
	DatasetStatusUninitialised DatasetStatus = "UNINITIALISED"
	// DatasetStatusProcessing means the dataset is being processed into datapoints.
	DatasetStatusProcessing DatasetStatus = "PROCESSING"
	// DatasetStatusReady means the dataset is ready to be used within a batch.
	DatasetStatusReady DatasetStatus = "READY"
	// DatasetStatusError means something went wrong during processing.
	DatasetStatusError DatasetStatus = "ERROR"
)

// TaskDetails represents the task configuration details.
type TaskDetails struct {
	TaskName         string `json:"task_name"`
	TaskIntroduction string `json:"task_introduction"`
	TaskSteps        string `json:"task_steps"`
}

// InstructionOption represents an option for multiple choice instructions.
type InstructionOption struct {
	Label   string `json:"label"`
	Value   string `json:"value"`
	Heading string `json:"heading,omitempty"`
}

// Instruction represents an instruction in a batch.
type Instruction struct {
	ID          string              `json:"id"`
	Type        string              `json:"type"`
	BatchID     string              `json:"batch_id"`
	CreatedBy   string              `json:"created_by"`
	CreatedAt   string              `json:"created_at"`
	Description string              `json:"description"`
	Options     []InstructionOption `json:"options,omitempty"`
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

// CreateAITaskBuilderCollection represents the payload for creating a collection.
type CreateAITaskBuilderCollection struct {
	WorkspaceID string           `json:"workspace_id" mapstructure:"workspace_id"`
	Name        string           `json:"name" mapstructure:"name"`
	Items       []CollectionPage `json:"items" mapstructure:"items"`
}

// CollectionPage represents a page in a collection containing instructions.
type CollectionPage struct {
	Order int                     `json:"order" mapstructure:"order"`
	Items []CollectionInstruction `json:"items" mapstructure:"items"`
}

// CollectionInstruction represents an instruction item within a collection page.
type CollectionInstruction struct {
	Order       int                 `json:"order" mapstructure:"order"`
	Type        string              `json:"type" mapstructure:"type"`
	Description string              `json:"description" mapstructure:"description"`
	Options     []InstructionOption `json:"options,omitempty" mapstructure:"options"`
	AnswerLimit *int                `json:"answer_limit,omitempty" mapstructure:"answer_limit"`
}
