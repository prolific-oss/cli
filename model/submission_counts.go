package model

import "fmt"

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

// SubmissionCountItem represents a single status-count pair for use in interactive lists.
type SubmissionCountItem struct {
	StatusLabel string
	StatusKey   string
	Count       int
}

// FilterValue implements list.Item for bubbletea.
func (i SubmissionCountItem) FilterValue() string { return i.StatusLabel }

// Title implements list.Item for bubbletea.
func (i SubmissionCountItem) Title() string { return i.StatusLabel }

// Description implements list.Item for bubbletea.
func (i SubmissionCountItem) Description() string {
	return fmt.Sprintf("%d submissions", i.Count)
}

// ToItems converts SubmissionCounts into a slice of SubmissionCountItem,
// excluding zero-count statuses and the Total row.
func (sc *SubmissionCounts) ToItems() []SubmissionCountItem {
	entries := []struct {
		label string
		key   string
		count int
	}{
		{"Active", "ACTIVE", sc.Active},
		{"Approved", "APPROVED", sc.Approved},
		{"Awaiting Review", "AWAITING REVIEW", sc.AwaitingReview},
		{"Rejected", "REJECTED", sc.Rejected},
		{"Reserved", "RESERVED", sc.Reserved},
		{"Returned", "RETURNED", sc.Returned},
		{"Timed Out", "TIMED-OUT", sc.TimedOut},
		{"Partially Approved", "PARTIALLY APPROVED", sc.PartiallyApproved},
		{"Screened Out", "SCREENED OUT", sc.ScreenedOut},
	}

	var items []SubmissionCountItem
	for _, e := range entries {
		if e.count > 0 {
			items = append(items, SubmissionCountItem{
				StatusLabel: e.label,
				StatusKey:   e.key,
				Count:       e.count,
			})
		}
	}
	return items
}
