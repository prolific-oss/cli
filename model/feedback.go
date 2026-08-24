package model

// StudyFeedbackRatings holds the rating values submitted alongside a piece of
// participant study feedback.
type StudyFeedbackRatings struct {
	Clarity    *float64 `json:"clarity_rating"`
	Difficulty *float64 `json:"difficulty_rating"`
	Fairness   *float64 `json:"fairness_rating"`
}

// StudyFeedback represents a single participant's feedback record for a study.
type StudyFeedback struct {
	ParticipantID *string              `json:"participant_id"`
	Category      *string              `json:"category"`
	Text          *string              `json:"text"`
	Ratings       StudyFeedbackRatings `json:"ratings"`
}

// StudyRating summarises the aggregated feedback ratings for a study, for a
// single rating category (e.g. clarity, difficulty, fairness).
type StudyRating struct {
	AverageRating *float64 `json:"average_rating"`
	TotalCount    int      `json:"total_count"`
}
