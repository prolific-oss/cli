package model

import (
	"time"
)

// AITaskBuilderBatch represents an AI task builder batch.
type AITaskBuilderBatch struct {
	ID                    string      `json:"id"`
	CreatedAt             time.Time   `json:"created_at"`
	CreatedBy             string      `json:"created_by"`
	Datasets              []Dataset   `json:"datasets"`
	Name                  string      `json:"name"`
	Status                string      `json:"status"`
	TotalTaskCount        int         `json:"total_task_count"`
	TotalInstructionCount int         `json:"total_instruction_count"`
	WorkspaceID           string      `json:"workspace_id"`
	SchemaVersion         int         `json:"schema_version"`
	TaskDetails           TaskDetails `json:"task_details"`
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
