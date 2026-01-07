package model

import (
	"fmt"
	"time"
)

// Collection represents a Prolific Collection
type Collection struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
	ItemCount int       `json:"item_count"`
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
	ID             string     `json:"id,omitempty" mapstructure:"id"`
	CreatedBy      string     `json:"created_by,omitempty" mapstructure:"created_by"`
	CreatedAt      *time.Time `json:"created_at,omitempty" mapstructure:"created_at"`
	SchemaVersion  int        `json:"schema_version,omitempty" mapstructure:"schema_version"`
	LastModifiedAt *time.Time `json:"last_modified_at,omitempty" mapstructure:"last_modified_at"`
	LastModifiedBy string     `json:"last_modified_by,omitempty" mapstructure:"last_modified_by"`
}

// InstructionType represents the type of instruction
type InstructionType string

const (
	InstructionTypeFreeText                   InstructionType = "free_text"
	InstructionTypeMultipleChoice             InstructionType = "multiple_choice"
	InstructionTypeMultipleChoiceWithFreeText InstructionType = "multiple_choice_with_free_text"
)

// MultipleChoiceOption represents an option for multiple choice instructions
type MultipleChoiceOption struct {
	Label   string `json:"label" mapstructure:"label"`
	Value   string `json:"value" mapstructure:"value"`
	Heading string `json:"heading,omitempty" mapstructure:"heading"` // Required for multiple_choice_with_free_text
}

type PageInstruction struct {
	BaseEntity `mapstructure:",squash"`

	// Required fields
	Type        InstructionType `json:"type" mapstructure:"type"`
	Description string          `json:"description" mapstructure:"description"`
	Order       int             `json:"order" mapstructure:"order"`

	// Optional - for free_text type
	PlaceholderTextInput string `json:"placeholder_text_input,omitempty" mapstructure:"placeholder_text_input"`

	// Optional - for multiple_choice and multiple_choice_with_free_text types
	AnswerLimit int                    `json:"answer_limit,omitempty" mapstructure:"answer_limit"`
	Options     []MultipleChoiceOption `json:"options,omitempty" mapstructure:"options"`
}

type Page struct {
	BaseEntity `mapstructure:",squash"`
	Order      int               `json:"order" mapstructure:"order"`
	Items      []PageInstruction `json:"items" mapstructure:"items"`
}

type UpdateCollection struct {
	BaseEntity  `mapstructure:",squash"`
	Name        string `json:"name" mapstructure:"name"`
	WorkspaceID string `json:"workspace_id" mapstructure:"workspace_id"`
	Items       []Page `json:"items" mapstructure:"items"`
}
