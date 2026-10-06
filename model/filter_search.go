package model

// FilterSearchLink is a link returned alongside filter search results.
type FilterSearchLink struct {
	Href  *string `json:"href"`
	Title string  `json:"title"`
}

// FilterSearchHighlight describes a matched span within a searchable field.
// Start and End are zero-based Unicode code-point positions (End exclusive).
// Filter-level matches use the filter's own text fields; choice matches
// always use the "label" field.
type FilterSearchHighlight struct {
	Field       string `json:"field"`
	QueryTerm   string `json:"query_term"`
	MatchedText string `json:"matched_text"`
	Start       int    `json:"start"`
	End         int    `json:"end"`
}

// FilterChoiceSearchResult is a matching choice, either previewed in a filter
// search result or returned by the choice search endpoint.
type FilterChoiceSearchResult struct {
	ID             string                  `json:"id"`
	Label          string                  `json:"label"`
	ParentID       *string                 `json:"parent_id"`
	NumChildren    int                     `json:"num_children"`
	NumDescendants int                     `json:"num_descendants"`
	Matches        []FilterSearchHighlight `json:"matches"`
}

// FilterSearchChoices groups a filter's choice metadata with the preview of
// choices matching the query. It is omitted for filters without enumerable
// choices, such as range filters.
type FilterSearchChoices struct {
	// Total is every available choice in the filter, independent of the query.
	Total int `json:"total"`
	// Matched is the number of matching choices before preview truncation.
	Matched int `json:"matched"`
	// Truncated reports whether Matched exceeds the number of Results.
	Truncated bool `json:"truncated"`
	// Results is the preview of up to three matching choices.
	Results []FilterChoiceSearchResult `json:"results"`
}

// FilterSearchResultLinks holds the follow-up links for a filter search result.
type FilterSearchResultLinks struct {
	Choices         *FilterSearchLink `json:"choices,omitempty"`
	MatchingChoices *FilterSearchLink `json:"matching_choices,omitempty"`
}

// FilterSearchResult is a filter returned by the filter search endpoint.
type FilterSearchResult struct {
	FilterID    string                   `json:"filter_id"`
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Question    *string                  `json:"question"`
	Category    *string                  `json:"category"`
	Subcategory *string                  `json:"subcategory"`
	Type        string                   `json:"type"`
	DataType    string                   `json:"data_type"`
	Matches     []FilterSearchHighlight  `json:"matches"`
	Choices     *FilterSearchChoices     `json:"choices,omitempty"`
	Min         any                      `json:"min,omitempty"`
	Max         any                      `json:"max,omitempty"`
	Links       *FilterSearchResultLinks `json:"_links,omitempty"`
}
