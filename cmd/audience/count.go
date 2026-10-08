package audience

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/ui"
	uiaudience "github.com/prolific-oss/cli/ui/audience"

	"github.com/spf13/cobra"
)

// NewCountCommand creates a new `audience count` command to count how many
// participants match a set of filters, without creating a study or saving a
// filter set.
func NewCountCommand(client client.API, w io.Writer) *cobra.Command {
	var in filterInput

	cmd := &cobra.Command{
		Use:   "count",
		Short: "Count participants matching a set of filters",
		Long: `Count how many participants would be eligible for a study defined by a
set of filters, without creating the study or saving a filter set.

Count a set of filters given via -p/--template-path or --filters.

Top-level filters are combined with AND. To express nested AND/OR groups,
use filter_id "and" or "or" with a selected_filters array of child filters.
Both -p/--template-path and --filters preserve this structure; the API
validates which combinations are allowed.`,
		Example: `
Count participants matching the filters in a JSON/YAML file (see
"prolific study create --help" for the filter format)
$ prolific audience count -p /path/to/filters.json -w <workspace-id>

Count participants matching filters given directly as a flag
$ prolific audience count --filters '[{"filter_id":"age","selected_range":{"lower":18,"upper":65}}]' -w <workspace-id>

Emit machine-readable output for scripting
$ prolific audience count -p /path/to/filters.json -w <workspace-id> --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := in.validateCount(); err != nil {
				return err
			}

			count, err := getCount(client, in)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			if err := renderCount(count, in, w); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	addCountFlags(cmd, &in)

	return cmd
}

func getCount(c client.API, in filterInput) (int, error) {
	spec, err := in.resolve()
	if err != nil {
		return 0, err
	}

	response, err := c.GetEligibilityCount(client.EligibilityCountPayload{
		Filters:     spec.Filters,
		WorkspaceID: in.WorkspaceID,
	})
	if err != nil {
		return 0, err
	}

	return response.Count, nil
}

// countOutput is the CLI's own shape for a count, so a change to the API's
// response never reaches our output.
type countOutput struct {
	Count int `json:"count"`
}

// renderCount writes a count in the format the caller asked for. A count is a
// single value, so the table and CSV forms are a one-row record; at a terminal
// without a format flag it stays a sentence.
func renderCount(count int, in filterInput, w io.Writer) error {
	format := shared.ResolveFormatForWriter(in.Output, w)

	fields := uiaudience.CountFields.Resolve(in.Fields, format)

	switch format {
	case ui.FormatJSON:
		rendered, err := RenderCountJSON(count)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, rendered)
		return err
	case ui.FormatCSV:
		return ui.CsvRenderer[uiaudience.CountItem]{}.Render(countItems(count), fields, w)
	case ui.FormatTable:
		return ui.TableRenderer[uiaudience.CountItem]{}.Render(countItems(count), fields, w)
	default:
		_, err := fmt.Fprintf(w, "Eligible participants: %d\n", count)
		return err
	}
}

// RenderCountJSON emits the count as the CLI-owned countOutput rather than the
// API's response, so a change to the API never reaches our output.
func RenderCountJSON(count int) (string, error) {
	payload, err := json.Marshal(countOutput{Count: count})
	if err != nil {
		return "", err
	}

	return string(payload), nil
}

func countItems(count int) []uiaudience.CountItem {
	return []uiaudience.CountItem{{Count: count}}
}
