package audience

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/prolific-oss/cli/client"
	"github.com/stretchr/testify/require"
)

func TestNestedFiltersSurviveInputAndPayloadEncoding(t *testing.T) {
	const filters = `[{"filter_id":"or","selected_filters":[{"filter_id":"age","selected_range":{"lower":18,"upper":30}},{"filter_id":"and","selected_filters":[{"filter_id":"age","selected_range":{"lower":50,"upper":65}},{"filter_id":"handedness","selected_values":["0"]}]}]}]`
	const dimension = `{"filter_id":"handedness","selected_values":["0","1"]}`
	const yamlTemplate = `filters:
  - filter_id: or
    selected_filters:
      - filter_id: age
        selected_range: {lower: 18, upper: 30}
      - filter_id: and
        selected_filters:
          - filter_id: age
            selected_range: {lower: 50, upper: 65}
          - filter_id: handedness
            selected_values: ["0"]
breakdown_filter:
  filter_id: handedness
  selected_values: ["0", "1"]
`
	for _, tc := range []struct{ name, extension, content string }{
		{name: "flags"},
		{name: "JSON template", extension: ".json", content: `{"filters":` + filters + `,"breakdown_filter":` + dimension + `}`},
		{name: "YAML template", extension: ".yaml", content: yamlTemplate},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := filterInput{FiltersJSON: filters, BreakdownJSON: dimension, WorkspaceID: "ws"}
			if tc.extension != "" {
				in.FiltersJSON, in.BreakdownJSON = "", ""
				in.TemplatePath = filepath.Join(t.TempDir(), "audience"+tc.extension)
				require.NoError(t, os.WriteFile(in.TemplatePath, []byte(tc.content), 0600))
			}
			spec, err := in.resolve()
			require.NoError(t, err)
			count, err := json.Marshal(client.EligibilityCountPayload{Filters: spec.Filters, WorkspaceID: in.WorkspaceID})
			require.NoError(t, err)
			require.JSONEq(t, `{"filters":`+filters+`,"workspace_id":"ws"}`, string(count))
			breakdown, err := json.Marshal(client.FilterBreakdownPayload{Filters: spec.Filters, BreakdownFilter: spec.BreakdownFilter, WorkspaceID: in.WorkspaceID})
			require.NoError(t, err)
			require.JSONEq(t, `{"filters":`+filters+`,"breakdown_filter":`+dimension+`,"workspace_id":"ws"}`, string(breakdown))
		})
	}
}
