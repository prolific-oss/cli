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
	InstructionTypeFreeText                   InstructionType = "free_text"
	InstructionTypeMultipleChoice             InstructionType = "multiple_choice"
	InstructionTypeMultipleChoiceWithFreeText InstructionType = "multiple_choice_with_free_text"
)

// MultipleChoiceOption represents an option for multiple choice instructions
type MultipleChoiceOption struct {
	Label   string `json:"label" yaml:"label" mapstructure:"label"`
	Value   string `json:"value" yaml:"value" mapstructure:"value"`
	Heading string `json:"heading,omitempty" yaml:"heading,omitempty" mapstructure:"heading"` // Required for multiple_choice_with_free_text
}

// PageInstruction represents a single instruction item within a collection page.
type PageInstruction struct {
	BaseEntity `yaml:",inline" mapstructure:",squash"`

	// Required fields
	Type        InstructionType `json:"type" yaml:"type" mapstructure:"type"`
	Description string          `json:"description" yaml:"description" mapstructure:"description"`
	Order       int             `json:"order" yaml:"order" mapstructure:"order"`

	// Optional - for free_text type
	PlaceholderTextInput string `json:"placeholder_text_input,omitempty" yaml:"placeholder_text_input,omitempty" mapstructure:"placeholder_text_input"`

	// Optional - for multiple_choice and multiple_choice_with_free_text types
	AnswerLimit int                    `json:"answer_limit,omitempty" yaml:"answer_limit,omitempty" mapstructure:"answer_limit"`
	Options     []MultipleChoiceOption `json:"options,omitempty" yaml:"options,omitempty" mapstructure:"options"`
}

// Page represents a single page within a collection.
type Page struct {
	BaseEntity `yaml:",inline" mapstructure:",squash"`
	Order      int               `json:"order" yaml:"order" mapstructure:"order"`
	Items      []PageInstruction `json:"items" yaml:"items" mapstructure:"items"`
}

// UpdateCollection represents the payload for updating a collection.
type UpdateCollection struct {
	BaseEntity  `yaml:",inline" mapstructure:",squash"`
	Name        string `json:"name" yaml:"name" mapstructure:"name"`
	WorkspaceID string `json:"workspace_id,omitempty" yaml:"workspace_id,omitempty" mapstructure:"workspace_id"`
	Items       []Page `json:"items" yaml:"items" mapstructure:"items"`
}
