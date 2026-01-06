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

// CollectionItem represents an item (page) in a collection with its order
type CollectionItem struct {
	ID    string `json:"id" mapstructure:"id"`
	Order int    `json:"order" mapstructure:"order"`
}

// UpdateCollection represents the payload for updating a collection
type UpdateCollection struct {
	Name  string           `json:"name,omitempty" mapstructure:"name"`
	Items []CollectionItem `json:"items,omitempty" mapstructure:"items"`
}
