package model

import "time"

// Hook represents a subscription to an event
type Hook struct {
	ID          string `json:"id"`
	EventType   string `json:"event_type"`
	TargetURL   string `json:"target_url"`
	IsEnabled   bool   `json:"is_enabled"`
	WorkspaceID string `json:"workspace_id"`
}

// HookEventType represents event types that are available
// to register on the webhook subscription
type HookEventType struct {
	EventType   string `json:"event_type"`
	Description string `json:"description"`
}

// HookEvent represents a point when Prolific notified the
// target URL of the event that the user has subscribed to.
type HookEvent struct {
	ID          string    `json:"id"`
	DateCreated time.Time `json:"datetime_created"`
	DateUpdated time.Time `json:"datetime_updated"`
	EventType   string    `json:"event_type"`
	ResourceID  string    `json:"resource_id"`
	Status      string    `json:"status"`
	TargetURL   string    `json:"target_url"`
}
