package model

import (
	"fmt"
	"time"
)

// SurveyResponse represents a participant's response to a survey.
type SurveyResponse struct {
	ID            string                   `json:"_id"`
	ParticipantID string                   `json:"participant_id"`
	SubmissionID  string                   `json:"submission_id"`
	DateCreated   time.Time                `json:"date_created"`
	DateModified  time.Time                `json:"date_modified"`
	Sections      []SurveyResponseSection  `json:"sections,omitempty"`
	Questions     []SurveyQuestionResponse `json:"questions,omitempty"`
}

// SurveyResponseListItem wraps a SurveyResponse to satisfy the bubbletea list.DefaultItem interface.
type SurveyResponseListItem struct {
	SurveyResponse
}

// FilterValue implements the bubbletea list.Item interface.
func (s SurveyResponseListItem) FilterValue() string { return s.ID }

// Title implements the bubbletea list.DefaultItem interface.
func (s SurveyResponseListItem) Title() string { return s.ID }

// Description implements the bubbletea list.DefaultItem interface.
func (s SurveyResponseListItem) Description() string {
	return fmt.Sprintf("Participant: %s - submitted %s", s.ParticipantID, s.DateCreated.Format("2006-01-02"))
}

// SurveyResponseSection represents a section within a survey response.
type SurveyResponseSection struct {
	SectionID string                   `json:"section_id" mapstructure:"section_id"`
	Questions []SurveyQuestionResponse `json:"questions" mapstructure:"questions"`
}

// SurveyQuestionResponse represents an answered question in a survey response.
type SurveyQuestionResponse struct {
	QuestionID    string                 `json:"question_id" mapstructure:"question_id"`
	QuestionTitle string                 `json:"question_title" mapstructure:"question_title"`
	Answers       []SurveyResponseAnswer `json:"answers" mapstructure:"answers"`
}

// SurveyResponseAnswer represents an answer within a survey response.
type SurveyResponseAnswer struct {
	AnswerID string `json:"answer_id" mapstructure:"answer_id"`
	Value    string `json:"value" mapstructure:"value"`
}

// CreateSurveyResponseRequest is the request model for creating a survey response.
type CreateSurveyResponseRequest struct {
	ParticipantID string                   `json:"participant_id" mapstructure:"participant_id"`
	SubmissionID  string                   `json:"submission_id" mapstructure:"submission_id"`
	Sections      []SurveyResponseSection  `json:"sections,omitempty" mapstructure:"sections"`
	Questions     []SurveyQuestionResponse `json:"questions,omitempty" mapstructure:"questions"`
}

// SurveySummary represents the aggregated summary of survey responses.
type SurveySummary struct {
	SurveyID  string                  `json:"survey_id"`
	Questions []SurveySummaryQuestion `json:"questions"`
}

// SurveySummaryQuestion represents a question's aggregated response data.
type SurveySummaryQuestion struct {
	QuestionID   string                `json:"question_id"`
	Question     string                `json:"question"`
	TotalAnswers int                   `json:"total_answers"`
	Answers      []SurveySummaryAnswer `json:"answers"`
}

// SurveySummaryAnswer represents an answer option's count in the summary.
type SurveySummaryAnswer struct {
	AnswerID string `json:"answer_id"`
	Answer   string `json:"answer"`
	Count    int    `json:"count"`
}
