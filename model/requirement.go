package model

import (
	"fmt"
	"strings"
)

// Requirement represents an eligibility requirement in the system.
type Requirement struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	Attributes      []RequirementAttribute `json:"attributes,omitempty"`
	Query           RequirementQuestion    `json:"query,omitempty"`
	Cls             string                 `json:"_cls"`
	Category        string                 `json:"category"`
	Subcategory     any                    `json:"subcategory"`
	Order           int                    `json:"order"`
	Recommended     bool                   `json:"recommended"`
	DetailsDisplay  string                 `json:"details_display"`
	RequirementType string                 `json:"requirement_type"`
}

// RequirementAttribute are all the attributes for a given requirement
type RequirementAttribute struct {
	Label string `json:"label,omitempty"`
	Name  string `json:"name,omitempty"`
	Value any    `json:"value,omitempty"`
	Index int    `json:"index,omitempty"`
}

type RequirementQuestion struct {
	ID                  string `json:"id"`
	Question            string `json:"question"`
	Description         string `json:"description"`
	Title               string `json:"title"`
	HelpText            string `json:"help_text"`
	ParticipantHelpText string `json:"participant_help_text"`
	ResearcherHelpText  string `json:"researcher_help_text"`
	IsNew               bool   `json:"is_new"`
}

// FilterValue will help the bubbletea views run
func (r Requirement) FilterValue() string {
	title := r.Query.Question
	if title == "" {
		title = r.Query.Title
	}
	return strings.Trim(title, " ")
}

// Title will set the main string for the view.
func (r Requirement) Title() string {
	title := r.Query.Question
	if title == "" {
		title = r.Query.Title
	}
	return strings.Trim(title, " ")
}

// Description will set the secondary string the view.
func (r Requirement) Description() string {
	desc := fmt.Sprintf("Category: %s", r.Category)

	if r.Query.Description != "" {
		desc += fmt.Sprintf(". %s", r.Query.Description)
	}
	return desc
}
