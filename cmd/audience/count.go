package audience

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"

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

			rendered, err := RenderCount(count, in.JSON)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			fmt.Fprintln(w, rendered)

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

// RenderCount produces output for an eligibility count. --json emits the
// CLI-owned countOutput rather than the API's response.
func RenderCount(count int, asJSON bool) (string, error) {
	if asJSON {
		payload, err := json.Marshal(countOutput{Count: count})
		if err != nil {
			return "", err
		}

		return string(payload), nil
	}

	return fmt.Sprintf("Eligible participants: %d", count), nil
}
