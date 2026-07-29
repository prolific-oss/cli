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
