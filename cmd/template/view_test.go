package template

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViewCommand(t *testing.T) {
	var b bytes.Buffer

	cmd := NewViewCommand(&b)
	cmd.SetArgs([]string{"standard-sample"})
	err := cmd.Execute()

	assert.NoError(t, err)

	output := b.String()
	assert.Contains(t, output, "My first standard sample")
	assert.Contains(t, output, "external_study_url")
	assert.Contains(t, output, "prolific_id_option")
}

func TestViewCommandCollection(t *testing.T) {
	var b bytes.Buffer

	cmd := NewViewCommand(&b)
	cmd.SetArgs([]string{"collection"})
	err := cmd.Execute()

	assert.NoError(t, err)
	assert.NotEmpty(t, b.String())
}

func TestViewCommandNotFound(t *testing.T) {
	var b bytes.Buffer

	cmd := NewViewCommand(&b)
	cmd.SetArgs([]string{"nonexistent"})
	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestViewCommandFormatYAML(t *testing.T) {
	var b bytes.Buffer

	cmd := NewViewCommand(&b)
	cmd.SetArgs([]string{"standard-sample", "--format", "yaml"})
	err := cmd.Execute()

	assert.NoError(t, err)

	output := b.String()
	assert.Contains(t, output, "name:")
	assert.Contains(t, output, "device_compatibility:")
	assert.NotContains(t, output, `"name":`)
}

func TestViewCommandDefaultsToJSON(t *testing.T) {
	var b bytes.Buffer

	cmd := NewViewCommand(&b)
	cmd.SetArgs([]string{"standard-sample"})
	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Contains(t, b.String(), "{")
}

func TestViewCommandNoArgs(t *testing.T) {
	var b bytes.Buffer

	cmd := NewViewCommand(&b)
	err := cmd.Execute()

	assert.Error(t, err)
}

func TestViewCommandLegacyIDs(t *testing.T) {
	tests := map[string]string{
		"standard-sample-aitaskbuilder":         "standard-sample-taskbuilder",
		"study-with-ai-task-builder-collection": "study-with-task-builder-collection",
	}

	for oldID, newID := range tests {
		t.Run(oldID, func(t *testing.T) {
			var oldOut, newOut bytes.Buffer

			oldCmd := NewViewCommand(&oldOut)
			oldCmd.SetArgs([]string{oldID})
			assert.NoError(t, oldCmd.Execute())

			newCmd := NewViewCommand(&newOut)
			newCmd.SetArgs([]string{newID})
			assert.NoError(t, newCmd.Execute())

			assert.NotEmpty(t, newOut.String())
			assert.Equal(t, newOut.String(), oldOut.String())
		})
	}
}
