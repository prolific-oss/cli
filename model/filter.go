package model

// Filter holds information about the filter that makes up a filter set
type Filter struct {
	ID                string      `json:"id" mapstructure:"id"`
	FilterID          string      `json:"filter_id" mapstructure:"filter_id"`
	FilterTitle       string      `json:"title" mapstructure:"title"`
	FilterDescription string      `json:"description" mapstructure:"description"`
	Question          string      `json:"question" mapstructure:"question"`
	Type              string      `json:"type" mapstructure:"type"`
	SelectedValues    []string    `json:"selected_values,omitempty" mapstructure:"selected_values"`
	SelectedRange     FilterRange `json:"selected_range,omitempty" mapstructure:"selected_range"`
}

// FilterRange holds the lower and upper bounds of a filter
type FilterRange struct {
	Lower interface{} `json:"lower,omitempty" mapstructure:"lower"`
	Upper interface{} `json:"upper,omitempty" mapstructure:"upper"`
}

// FilterValue will help the bubbletea views run
func (f Filter) FilterValue() string {
	return f.FilterTitle
}

// Title will return the title of the filter
func (f Filter) Title() string {
	return f.FilterTitle
}

// Description will return the description of the filter
func (f Filter) Description() string {
	return f.FilterDescription
}
