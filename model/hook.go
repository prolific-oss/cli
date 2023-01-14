package model

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
