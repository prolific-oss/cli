package template

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListTemplates(t *testing.T) {
	templates := listTemplates()

	assert.NotEmpty(t, templates)

	ids := make(map[string]bool)
	for _, tmpl := range templates {
		ids[tmpl.ID+"."+tmpl.Format] = true
		assert.NotEmpty(t, tmpl.ID)
		assert.NotEmpty(t, tmpl.Filename)
		assert.NotEmpty(t, tmpl.Category)
		assert.Contains(t, []string{"study", "collection"}, tmpl.Category)
		assert.Contains(t, []string{"json", "yaml"}, tmpl.Format)
	}

	assert.True(t, ids["standard-sample.json"])
	assert.True(t, ids["collection.json"])
	assert.True(t, ids["collection.yaml"])
}

func TestListTemplatesExcludesBatchInstructions(t *testing.T) {
	templates := listTemplates()

	for _, tmpl := range templates {
		assert.NotEqual(t, "batch-instructions", tmpl.ID)
	}
}

func TestListCommand(t *testing.T) {
	var b bytes.Buffer

	cmd := NewListCommand(&b)
	err := cmd.Execute()

	assert.NoError(t, err)

	output := b.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "Category")
	assert.Contains(t, output, "Format")
	assert.Contains(t, output, "standard-sample")
	assert.Contains(t, output, "collection")
	assert.Contains(t, output, "study")
}

func TestCategorise(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"standard-sample", "study"},
		{"standard-sample-aitaskbuilder", "study"},
		{"minimal-study", "study"},
		{"multi-submission-study", "study"},
		{"multiple-participant-groups-either-or", "study"},
		{"study-with-filters", "study"},
		{"star-trek", "study"},
		{"star-wars", "study"},
		{"collection", "collection"},
		{"batch-instructions", ""},
		{"aitb-model-evaluation", ""},
		{"credentials", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, categorise(tt.name))
		})
	}
}
