package model

// SubmissionCounts represents the count of submissions grouped by status for a study.
type SubmissionCounts struct {
	Active            int `json:"ACTIVE"`
	Approved          int `json:"APPROVED"`
	AwaitingReview    int `json:"AWAITING REVIEW"`
	Rejected          int `json:"REJECTED"`
	Reserved          int `json:"RESERVED"`
	Returned          int `json:"RETURNED"`
	TimedOut          int `json:"TIMED-OUT"`
	PartiallyApproved int `json:"PARTIALLY APPROVED"`
	ScreenedOut       int `json:"SCREENED OUT"`
	Total             int `json:"TOTAL"`
}
