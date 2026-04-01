package model

import (
	"fmt"
	"time"
)

// Survey represents a survey resource from the Prolific API.
type Survey struct {
	ID           string           `json:"_id"`
	ResearcherID string           `json:"researcher_id"`
	Title        string           `json:"title"`
	DateCreated  time.Time        `json:"date_created"`
	DateModified time.Time        `json:"date_modified"`
	Sections     []SurveySection  `json:"sections,omitempty"`
	Questions    []SurveyQuestion `json:"questions,omitempty"`
}

// FilterValue implements the bubbletea list.Item interface.
func (s Survey) FilterValue() string { return s.Title }

// Title returns the title for the bubbletea list.Item interface.
func (s Survey) ListTitle() string { return s.Title }

// Description returns a description for the bubbletea list.Item interface.
func (s Survey) ListDescription() string {
	return fmt.Sprintf("ID: %s - created %s", s.ID, s.DateCreated.Format("2006-01-02"))
}

// SurveySection represents a section within a survey.
type SurveySection struct {
	ID        string           `json:"id,omitempty" mapstructure:"id"`
	Title     string           `json:"title" mapstructure:"title"`
	Questions []SurveyQuestion `json:"questions" mapstructure:"questions"`
}

// SurveyQuestion represents a question within a survey.
type SurveyQuestion struct {
	ID      string               `json:"id,omitempty" mapstructure:"id"`
	Title   string               `json:"title" mapstructure:"title"`
	Type    string               `json:"type" mapstructure:"type"`
	Answers []SurveyAnswerOption `json:"answers" mapstructure:"answers"`
}

// SurveyAnswerOption represents an answer option for a survey question.
type SurveyAnswerOption struct {
	ID    string `json:"id,omitempty" mapstructure:"id"`
	Value string `json:"value" mapstructure:"value"`
}

// CreateSurvey is the request model for creating a survey.
type CreateSurvey struct {
	ResearcherID string           `json:"researcher_id,omitempty" mapstructure:"researcher_id"`
	Title        string           `json:"title" mapstructure:"title"`
	Sections     []SurveySection  `json:"sections,omitempty" mapstructure:"sections"`
	Questions    []SurveyQuestion `json:"questions,omitempty" mapstructure:"questions"`
}
