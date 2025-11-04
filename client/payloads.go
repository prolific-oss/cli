package client

// MessagePayload represents the JSON payload for sending a message
type SendMessagePayload struct {
	RecipientID string `json:"recipient_id"`
	StudyID     string `json:"study_id"`
	Body        string `json:"body"`
}

// CreateAITaskBuilderDatasetPayload represents the request for creating a dataset
type CreateAITaskBuilderDatasetPayload struct {
	Name string `json:"name"`
}

// TaskDetails represents the task configuration details for batch creation
type TaskDetails struct {
	TaskName         string `json:"task_name"`
	TaskIntroduction string `json:"task_introduction"`
	TaskSteps        string `json:"task_steps"`
}

// CreateAITaskBuilderBatchPayload represents the JSON payload for creating an AI Task Builder batch
type CreateAITaskBuilderBatchPayload struct {
	Name        string      `json:"name"`
	WorkspaceID string      `json:"workspace_id"`
	DatasetID   string      `json:"dataset_id"`
	TaskDetails TaskDetails `json:"task_details"`
}
