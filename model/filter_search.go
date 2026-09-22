package model

// FilterSearchLink is a link returned alongside filter search results.
type FilterSearchLink struct {
	Href  *string `json:"href"`
	Title string  `json:"title"`
}

// FilterSearchHighlight describes a matched span within a searchable field.
// Start and End are zero-based Unicode code-point positions (End exclusive).
type FilterSearchHighlight struct {
	Field       string `json:"field"`
	QueryTerm   string `json:"query_term"`
	MatchedText string `json:"matched_text"`
	Start       int    `json:"start"`
	End         int    `json:"end"`
}

// FilterSearchMatch explains why a filter matched a search query.
type FilterSearchMatch struct {
	Fields     []string                `json:"fields"`
	Highlights []FilterSearchHighlight `json:"highlights"`
}

// FilterChoiceMatch explains why a choice matched a search query.
type FilterChoiceMatch struct {
	Highlights []FilterSearchHighlight `json:"highlights"`
}

// FilterChoiceSearchResult is a matching choice previewed in a filter search result.
type FilterChoiceSearchResult struct {
	ID             string            `json:"id"`
	Label          string            `json:"label"`
	ParentID       *string           `json:"parent_id"`
	NumChildren    int               `json:"num_children"`
	NumDescendants int               `json:"num_descendants"`
	Match          FilterChoiceMatch `json:"match"`
}

// FilterMatchingChoices is the preview of choices matching a search query.
type FilterMatchingChoices struct {
	Matched   int                        `json:"matched"`
	Truncated bool                       `json:"truncated"`
	Results   []FilterChoiceSearchResult `json:"results"`
}

// FilterSearchResultLinks holds the follow-up links for a filter search result.
type FilterSearchResultLinks struct {
	Choices         *FilterSearchLink `json:"choices,omitempty"`
	MatchingChoices *FilterSearchLink `json:"matching_choices,omitempty"`
}

// FilterSearchResult is a filter returned by the filter search endpoint.
type FilterSearchResult struct {
	FilterID       string                   `json:"filter_id"`
	Title          string                   `json:"title"`
	Description    string                   `json:"description"`
	Question       *string                  `json:"question"`
	Category       *string                  `json:"category"`
	Subcategory    *string                  `json:"subcategory"`
	Type           string                   `json:"type"`
	DataType       string                   `json:"data_type"`
	Match          FilterSearchMatch        `json:"match"`
	NumChoices     *int                     `json:"num_choices,omitempty"`
	MatchedChoices *FilterMatchingChoices   `json:"matched_choices,omitempty"`
	Min            any                      `json:"min,omitempty"`
	Max            any                      `json:"max,omitempty"`
	Links          *FilterSearchResultLinks `json:"_links,omitempty"`
}
