package ui_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/prolific-oss/cli/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type envelopeItem struct {
	ID string `json:"id"`
}

func TestJSONEnvelopeRendererRendersTheCLIEnvelope(t *testing.T) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	r := ui.JSONEnvelopeRenderer[envelopeItem]{}
	require.NoError(t, r.Render(ui.NewEnvelope([]envelopeItem{{ID: "age"}}, 90, 200, 20), writer))
	require.NoError(t, writer.Flush())

	expected := `{
  "results": [
    {
      "id": "age"
    }
  ],
  "count": 90,
  "limit": 200,
  "offset": 20
}
`
	assert.Equal(t, expected, b.String())
}

func TestJSONEnvelopeRendererRendersNoResultsAsAnEmptyArray(t *testing.T) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	r := ui.JSONEnvelopeRenderer[envelopeItem]{}
	require.NoError(t, r.Render(ui.NewEnvelope[envelopeItem](nil, 0, 200, 0), writer))
	require.NoError(t, writer.Flush())

	assert.Contains(t, b.String(), `"results": []`)
}
