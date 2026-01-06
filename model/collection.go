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
