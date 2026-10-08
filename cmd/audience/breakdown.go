package audience

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/ui"

	"github.com/spf13/cobra"
)

// NewBreakdownCommand creates a new `audience breakdown` command to count how
// many participants match a set of base filters, split by the values of a
// single breakdown filter.
func NewBreakdownCommand(client client.API, w io.Writer) *cobra.Command {
	var in filterInput

	cmd := &cobra.Command{
		Use:   "breakdown",
		Short: "Count eligible participants broken down by a filter",
		Long: `Count how many participants would be eligible for a study defined by a
set of base filters, split by the values (or bucketed ranges, for numeric
filters) of a single breakdown filter.

Provide the filters either as a -p/--template-path file, or directly via
--filters and --breakdown — not both.

Base filters can contain nested AND/OR groups: use filter_id "and" or "or"
with a selected_filters array of child filters. The API validates allowed
combinations for your workspace. The --breakdown filter remains a single
choice or range filter, not an AND/OR group. See "prolific study create --help"
and docs/examples/study-with-nested-filters.json for the base filter format.

The result includes an "N/A" count: participants who match the base
filters but don't fall into any of the breakdown filter's selected values
or range — for example, because they haven't answered that screener
question, or their answer falls outside what you specified.`,
		Example: `
Count participants matching the base filters and breakdown_filter in a
JSON/YAML file (see "prolific study create --help" for the filter format):
$ prolific audience breakdown -p /path/to/filters.json -w <workspace-id>

Or provide the filters directly as flags. A choice-type breakdown filter
needs selected_values listing which choices to split by — the API rejects
a filter_id with no values:
$ prolific audience breakdown \
    --filters '[{"filter_id":"age","selected_range":{"lower":18,"upper":65}}]' \
    --breakdown '{"filter_id":"handedness","selected_values":["0","1"]}' \
    -w <workspace-id>

Emit machine-readable output for scripting
$ prolific audience breakdown -p /path/to/filters.json -w <workspace-id> --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := in.validateBreakdown(); err != nil {
				return err
			}

			breakdown, err := getBreakdown(client, in)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			if err := renderBreakdown(breakdown, in, w); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	addBreakdownFlags(cmd, &in)

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

// renderBreakdown writes a breakdown in the format the caller asked for. A
// breakdown is one row per value participants gave, so the table and CSV
// forms carry the same rows and a spreadsheet can read the CSV as it stands.
func renderBreakdown(breakdown map[string]int, in filterInput, w io.Writer) error {
	items := NewBreakdownItems(breakdown)

	switch shared.ResolveFormatForWriter(in.Output, w) {
	case shared.FormatJSON:
		rendered, err := RenderBreakdownJSON(breakdown)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, rendered)
		return err
	case shared.FormatCSV:
		return ui.CsvRenderer[BreakdownItem]{}.Render(items, in.Fields, w)
	default:
		return ui.TableRenderer[BreakdownItem]{}.Render(items, in.Fields, w)
	}
}

// RenderBreakdownJSON emits the counts keyed by breakdown value. The counts
// are the whole payload, so they are not wrapped in a key that repeats the
// name of the command. A breakdown with no buckets is an empty object rather
// than null, so consumers can index it either way.
func RenderBreakdownJSON(breakdown map[string]int) (string, error) {
	if breakdown == nil {
		breakdown = map[string]int{}
	}

	payload, err := json.Marshal(breakdown)
	if err != nil {
		return "", err
	}

	return string(payload), nil
}
