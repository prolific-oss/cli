package audience

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/prolific-oss/cli/client"
	"github.com/stretchr/testify/require"
)

// nestedFiltersExample is the study template that `study create --help`,
// `audience count --help` and `filter-sets create --help` all point readers at
// for the nested AND/OR format. Nothing else reads it, so without this test it
// could drift out of step with what the CLI actually accepts.
const nestedFiltersExample = "../../docs/examples/study-with-nested-filters.json"

func TestShippedNestedFiltersExampleRoundTrips(t *testing.T) {
	raw, err := os.ReadFile(filepath.Clean(nestedFiltersExample))
	require.NoError(t, err, "the example referenced from command help must exist")

	var template struct {
		Filters json.RawMessage `json:"filters"`
	}
	require.NoError(t, json.Unmarshal(raw, &template))
	require.NotEmpty(t, template.Filters, "the example must carry a filters array")

	spec, err := filterInput{FiltersJSON: string(template.Filters), WorkspaceID: "ws"}.resolve()
	require.NoError(t, err)

	payload, err := json.Marshal(client.EligibilityCountPayload{Filters: spec.Filters, WorkspaceID: "ws"})
	require.NoError(t, err)

	require.JSONEq(t, `{"filters":`+string(template.Filters)+`,"workspace_id":"ws"}`, string(payload),
		"the example's nested groups must survive the CLI unchanged")
}
