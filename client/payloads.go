package client

// SendMessagePayload represents the JSON payload for sending a message.
type SendMessagePayload struct {
	RecipientID string `json:"recipient_id"`
	StudyID     string `json:"study_id"`
	Body        string `json:"body"`
}

// BulkSendMessagePayload represents the JSON payload for
// sending a message to multiple participants.
type BulkSendMessagePayload struct {
	IDs     []string `json:"ids"`
	Body    string   `json:"body"`
	StudyID string   `json:"study_id"`
}

// SendGroupMessagePayload represents the JSON payload for
// sending a message to a participant group.
type SendGroupMessagePayload struct {
	ParticipantGroupID string `json:"participant_group_id"`
	Body               string `json:"body"`
	StudyID            string `json:"study_id,omitempty"`
}

// CreateAITaskBuilderDatasetPayload represents the request for creating a dataset
type CreateAITaskBuilderDatasetPayload struct {
	Name        string `json:"name"`
	WorkspaceID string `json:"workspace_id"`
}

// CreateBatchParams represents the parameters for creating an AI Task Builder batch.
type CreateBatchParams struct {
	Name             string `json:"name"`
	WorkspaceID      string `json:"workspace_id"`
	DatasetID        string `json:"dataset_id"`
	TaskName         string `json:"task_name"`
	TaskIntroduction string `json:"task_introduction"`
	TaskSteps        string `json:"task_steps"`
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

// InstructionType represents the type of instruction.
type InstructionType string

const (
	// InstructionTypeMultipleChoice represents a multiple choice instruction.
	InstructionTypeMultipleChoice InstructionType = "multiple_choice"
	// InstructionTypeFreeText represents a free text instruction.
	InstructionTypeFreeText InstructionType = "free_text"
	// InstructionTypeMultipleChoiceWithFreeText represents a multiple choice instruction with free text.
	InstructionTypeMultipleChoiceWithFreeText InstructionType = "multiple_choice_with_free_text"
	// InstructionTypeFreeTextWithUnit represents a free text instruction with unit selection.
	InstructionTypeFreeTextWithUnit InstructionType = "free_text_with_unit"
)

// InstructionOption represents an option for multiple choice instructions
type InstructionOption struct {
	Label   string `json:"label"`
	Value   string `json:"value"`
	Heading string `json:"heading,omitempty"`
}

// UnitOption represents a unit option for multiple_choice_with_unit instructions
type UnitOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// AnswerLimit represents the answer limit for multiple choice with free text instructions
type AnswerLimit struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// Instruction represents a single instruction in the request payload
type Instruction struct {
	Type                 InstructionType     `json:"type"`
	CreatedBy            string              `json:"created_by"`
	Description          string              `json:"description"`
	Options              []InstructionOption `json:"options,omitempty"`
	AnswerLimit          *AnswerLimit        `json:"answer_limit,omitempty"`
	UnitOptions          []UnitOption        `json:"unit_options,omitempty"`
	DefaultUnit          string              `json:"default_unit,omitempty"`
	UnitPosition         string              `json:"unit_position,omitempty"`
	HelperText           string              `json:"helper_text,omitempty"`
	PlaceholderTextInput string              `json:"placeholder_text_input,omitempty"`
}

// CreateAITaskBuilderInstructionsPayload represents the JSON payload for creating AI Task Builder instructions
type CreateAITaskBuilderInstructionsPayload struct {
	Instructions []Instruction `json:"instructions"`
}

// SetupAITaskBuilderBatchPayload represents the JSON payload for setting up an AI Task Builder batch
type SetupAITaskBuilderBatchPayload struct {
	DatasetID     string `json:"dataset_id"`
	TasksPerGroup int    `json:"tasks_per_group"`
}

// CredentialPoolPayload represents the JSON payload for creating a credential pool
type CredentialPoolPayload struct {
	Credentials string `json:"credentials"`
	WorkspaceID string `json:"workspace_id"`
}

// UpdateCredentialPoolPayload represents the JSON payload for updating a credential pool
type UpdateCredentialPoolPayload struct {
	Credentials string `json:"credentials"`
}
