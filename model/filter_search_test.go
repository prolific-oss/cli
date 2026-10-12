package model_test

import (
	"encoding/json"
	"testing"

	"github.com/prolific-oss/cli/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFilterSearchResultDecodesDocumentedExample decodes the final
// filter-result example from the schema change notes for
// GET /api/v1/filters/search/, guarding the flattened matches array and the
// grouped choices object.
func TestFilterSearchResultDecodesDocumentedExample(t *testing.T) {
	payload := `{
	  "filter_id": "job-title",
	  "title": "Software roles",
	  "description": "The participant’s current occupation.",
	  "question": "What is your current job title?",
	  "category": "Employment",
	  "subcategory": "Occupation",
	  "type": "select",
	  "data_type": "ChoiceID",
	  "matches": [
	    {"field": "title", "query_term": "software", "matched_text": "Software", "start": 0, "end": 8}
	  ],
	  "choices": {
	    "total": 5,
	    "matched": 1,
	    "truncated": false,
	    "results": [
	      {
	        "id": "1",
	        "label": "Software developers",
	        "parent_id": "0",
	        "num_children": 0,
	        "num_descendants": 0,
	        "matches": [
	          {"field": "label", "query_term": "software", "matched_text": "Software", "start": 0, "end": 8}
	        ]
	      }
	    ]
	  }
	}`

	var r model.FilterSearchResult
	require.NoError(t, json.Unmarshal([]byte(payload), &r))

	assert.Equal(t, "job-title", r.FilterID)
	assert.Equal(t, []model.FilterSearchHighlight{
		{Field: "title", QueryTerm: "software", MatchedText: "Software", Start: 0, End: 8},
	}, r.Matches)

	require.NotNil(t, r.Choices)
	assert.Equal(t, 5, r.Choices.Total)
	assert.Equal(t, 1, r.Choices.Matched)
	assert.False(t, r.Choices.Truncated)
	require.Len(t, r.Choices.Results, 1)

	choice := r.Choices.Results[0]
	assert.Equal(t, "1", choice.ID)
	assert.Equal(t, "Software developers", choice.Label)
	require.NotNil(t, choice.ParentID)
	assert.Equal(t, "0", *choice.ParentID)
	assert.Equal(t, []model.FilterSearchHighlight{
		{Field: "label", QueryTerm: "software", MatchedText: "Software", Start: 0, End: 8},
	}, choice.Matches)
}

func TestFilterSearchResultOmitsChoicesForRangeFilters(t *testing.T) {
	payload := `{"filter_id":"age","title":"Age","description":"","question":null,"category":null,"subcategory":null,"type":"range","data_type":"integer","matches":[],"min":18,"max":100}`

	var r model.FilterSearchResult
	require.NoError(t, json.Unmarshal([]byte(payload), &r))

	assert.Nil(t, r.Choices)
	assert.Empty(t, r.Matches)

	// Round-tripping must not invent a choices key for range filters.
	out, err := json.Marshal(r)
	require.NoError(t, err)
	assert.NotContains(t, string(out), `"choices"`)
}
