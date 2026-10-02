//nolint:dupl // Similar patterns are expected for CLI commands
package audience

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"

	"github.com/spf13/cobra"
)

// NewBreakdownCommand creates a new `audience breakdown` command to count how
// many participants match a set of base filters, split by the values of a
// single breakdown filter.
func NewBreakdownCommand(client client.API, w io.Writer) *cobra.Command {
	var in filterInput
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "breakdown",
		Short: "Count eligible participants broken down by a filter",
		Long: `Count how many participants would be eligible for a study defined by a
set of base filters, split by the values (or bucketed ranges, for numeric
filters) of a single breakdown filter.

Provide the filters either as a -t/--template-path file, or directly via
--filters and --breakdown — not both.

The result includes an "N/A" count: participants who match the base
filters but don't fall into any of the breakdown filter's selected values
or range — for example, because they haven't answered that screener
question, or their answer falls outside what you specified.`,
		Example: `
Count participants matching the base filters and breakdown_filter in a
JSON/YAML file (see "prolific study create --help" for the filter format):
$ prolific audience breakdown -t /path/to/filters.json -w <workspace-id>

Or provide the filters directly as flags. A choice-type breakdown filter
needs selected_values listing which choices to split by — the API rejects
a filter_id with no values:
$ prolific audience breakdown \
    --filters '[{"filter_id":"age","selected_range":{"lower":18,"upper":65}}]' \
    --breakdown '{"filter_id":"handedness","selected_values":["0","1"]}' \
    -w <workspace-id>

Emit machine-readable output for scripting
$ prolific audience breakdown -t /path/to/filters.json -w <workspace-id> --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := in.validate(true); err != nil {
				return err
			}

			breakdown, err := getBreakdown(client, in)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			rendered, err := RenderBreakdown(breakdown, asJSON)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			fmt.Fprint(w, rendered)

			return nil
		},
	}

	addFilterFlags(cmd, &in, true)
	// -j is bound by hand rather than through shared.AddOutputFlags, which
	// would claim -t for --table and collide with --template-path.
	cmd.Flags().BoolVarP(&asJSON, "json", "j", false, "Output as JSON")

	return cmd
}

func getBreakdown(c client.API, in filterInput) (map[string]int, error) {
	spec, err := in.resolve()
	if err != nil {
		return nil, err
	}

	if spec.BreakdownFilter.FilterID == "" {
		return nil, fmt.Errorf("breakdown filter must include a filter_id")
	}

	response, err := c.GetFilterBreakdown(client.FilterBreakdownPayload{
		Filters:         spec.Filters,
		BreakdownFilter: spec.BreakdownFilter,
		WorkspaceID:     in.WorkspaceID,
	})
	if err != nil {
		return nil, err
	}

	return response.Breakdown, nil
}

// RenderBreakdown produces output for a filter breakdown. The table form is
// sorted alphabetically with client.FilterBreakdownNAKey always shown last;
// --json instead emits the raw API response shape.
func RenderBreakdown(breakdown map[string]int, asJSON bool) (string, error) {
	if asJSON {
		payload, err := json.Marshal(client.FilterBreakdownResponse{Breakdown: breakdown})
		if err != nil {
			return "", err
		}

		return string(payload), nil
	}

	keys := make([]string, 0, len(breakdown))
	for key := range breakdown {
		if key != client.FilterBreakdownNAKey {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if _, ok := breakdown[client.FilterBreakdownNAKey]; ok {
		keys = append(keys, client.FilterBreakdownNAKey)
	}

	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "VALUE\tCOUNT")
	for _, key := range keys {
		fmt.Fprintf(tw, "%s\t%d\n", key, breakdown[key])
	}
	tw.Flush()

	return buf.String(), nil
}
