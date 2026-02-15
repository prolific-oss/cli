package model

import (
	"fmt"
	"time"
)

// Collection represents a Prolific Collection
type Collection struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	CreatedAt   time.Time    `json:"created_at"`
	CreatedBy   string       `json:"created_by"`
	ItemCount   int          `json:"item_count"`
	TaskDetails *TaskDetails `json:"task_details,omitempty"`
}

// FilterValue will help the bubbletea views run
func (c Collection) FilterValue() string { return c.Name }

// Title will set the main string for the view.
func (c Collection) Title() string { return c.Name }

// Description will set the secondary string for the view.
func (c Collection) Description() string {
	return fmt.Sprintf("%d pages - created by %s", c.ItemCount, c.CreatedBy)
}

// BaseEntity contains common fields for all collection entities
type BaseEntity struct {
	ID             string     `json:"id,omitempty" yaml:"id,omitempty" mapstructure:"id"`
	CreatedBy      string     `json:"created_by,omitempty" yaml:"created_by,omitempty" mapstructure:"created_by"`
	CreatedAt      *time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty" mapstructure:"created_at"`
	SchemaVersion  int        `json:"schema_version,omitempty" yaml:"schema_version,omitempty" mapstructure:"schema_version"`
	LastModifiedAt *time.Time `json:"last_modified_at,omitempty" yaml:"last_modified_at,omitempty" mapstructure:"last_modified_at"`
	LastModifiedBy string     `json:"last_modified_by,omitempty" yaml:"last_modified_by,omitempty" mapstructure:"last_modified_by"`
}

// InstructionType represents the type of instruction
type InstructionType string

const (
	// Instruction types (interactive - participants respond to these)
	InstructionTypeFreeText                   InstructionType = "free_text"
	InstructionTypeMultipleChoice             InstructionType = "multiple_choice"
	InstructionTypeMultipleChoiceWithFreeText InstructionType = "multiple_choice_with_free_text"
	InstructionTypeFreeTextWithUnit           InstructionType = "free_text_with_unit"

	// Content block types (non-interactive - for context or guidance)
	ContentBlockTypeRichText InstructionType = "rich_text"
	ContentBlockTypeImage    InstructionType = "image"
)

// MultipleChoiceOption represents an option for multiple choice instructions
type MultipleChoiceOption struct {
	Label   string `json:"label" yaml:"label" mapstructure:"label"`
	Value   string `json:"value" yaml:"value" mapstructure:"value"`
	Heading string `json:"heading,omitempty" yaml:"heading,omitempty" mapstructure:"heading"` // Required for multiple_choice_with_free_text
}

// UnitOption represents a unit option for free_text_with_unit instructions
type UnitOption struct {
	Label string `json:"label" yaml:"label" mapstructure:"label"`
	Value string `json:"value" yaml:"value" mapstructure:"value"`
}

// PageInstruction represents a single page item within a collection page.
// This can be either an instruction (interactive) or a content block (non-interactive).
type PageInstruction struct {
	BaseEntity `yaml:",inline" mapstructure:",squash"`

	// Required fields
	Type  InstructionType `json:"type" yaml:"type" mapstructure:"type"`
	Order int             `json:"order" yaml:"order" mapstructure:"order"`

	// Required for instruction types (free_text, multiple_choice, multiple_choice_with_free_text, free_text_with_unit, file_upload)
	Description string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description"`

	// Optional - for free_text and free_text_with_unit types
	PlaceholderTextInput string `json:"placeholder_text_input,omitempty" yaml:"placeholder_text_input,omitempty" mapstructure:"placeholder_text_input"`
	HelperText           string `json:"helper_text,omitempty" yaml:"helper_text,omitempty" mapstructure:"helper_text"`

	// Optional - for multiple_choice and multiple_choice_with_free_text types
	AnswerLimit     int                    `json:"answer_limit,omitempty" yaml:"answer_limit,omitempty" mapstructure:"answer_limit"`
	Options         []MultipleChoiceOption `json:"options,omitempty" yaml:"options,omitempty" mapstructure:"options"`
	DisableDropdown *bool                  `json:"disable_dropdown,omitempty" yaml:"disable_dropdown,omitempty" mapstructure:"disable_dropdown"`

	// Optional - for free_text_with_unit type
	UnitOptions  []UnitOption `json:"unit_options,omitempty" yaml:"unit_options,omitempty" mapstructure:"unit_options"`
	DefaultUnit  string       `json:"default_unit,omitempty" yaml:"default_unit,omitempty" mapstructure:"default_unit"`
	UnitPosition UnitPosition `json:"unit_position,omitempty" yaml:"unit_position,omitempty" mapstructure:"unit_position"`

	// Optional - for file_upload type
	AcceptedFileTypes []string `json:"accepted_file_types,omitempty" yaml:"accepted_file_types,omitempty" mapstructure:"accepted_file_types"`
	MaxFileSizeMB     *float64 `json:"max_file_size_mb,omitempty" yaml:"max_file_size_mb,omitempty" mapstructure:"max_file_size_mb"`
	MinFileCount      *int     `json:"min_file_count,omitempty" yaml:"min_file_count,omitempty" mapstructure:"min_file_count"`
	MaxFileCount      *int     `json:"max_file_count,omitempty" yaml:"max_file_count,omitempty" mapstructure:"max_file_count"`

	// Content block fields - for rich_text type
	Content string `json:"content,omitempty" yaml:"content,omitempty" mapstructure:"content"`

	// Content block fields - for image type
	URL     string `json:"url,omitempty" yaml:"url,omitempty" mapstructure:"url"`
	AltText string `json:"alt_text,omitempty" yaml:"alt_text,omitempty" mapstructure:"alt_text"`
	Caption string `json:"caption,omitempty" yaml:"caption,omitempty" mapstructure:"caption"`
}

// Page represents a single page within a collection.
type Page struct {
	BaseEntity `yaml:",inline" mapstructure:",squash"`
	Order      int               `json:"order" yaml:"order" mapstructure:"order"`
	PageItems  []PageInstruction `json:"page_items" yaml:"page_items" mapstructure:"page_items"`
}

// UpdateCollection represents the payload for updating a collection.
type UpdateCollection struct {
	BaseEntity      `yaml:",inline" mapstructure:",squash"`
	Name            string       `json:"name" yaml:"name" mapstructure:"name"`
	Description     string       `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description"`
	WorkspaceID     string       `json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" mapstructure:"workspace_id"`
	TaskDetails     *TaskDetails `json:"task_details,omitempty" yaml:"task_details,omitempty" mapstructure:"task_details"`
	CollectionItems []Page       `json:"collection_items" yaml:"collection_items" mapstructure:"collection_items"`
}
