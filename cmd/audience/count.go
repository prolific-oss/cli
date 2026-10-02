//nolint:dupl // Similar patterns are expected for CLI commands
package audience

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"

	"github.com/spf13/cobra"
)

// CountResult is the machine-readable shape emitted by --json.
type CountResult struct {
	Count int `json:"count"`
}

// NewCountCommand creates a new `audience count` command to count how many
// participants match a set of filters, without creating a study or saving a
// filter set.
func NewCountCommand(client client.API, w io.Writer) *cobra.Command {
	var in filterInput
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "count",
		Short: "Count participants matching a set of filters",
		Long: `Count how many participants would be eligible for a study defined by a
set of filters, without creating the study or saving a filter set.

Count a set of filters given via -t/--template-path or --filters.

Filters are a flat list, which the API combines with AND. The API also
supports nested and/or filter groups, but those cannot yet be expressed
via -t/--template-path or --filters.`,
		Example: `
Count participants matching the filters in a JSON/YAML file (see
"prolific study create --help" for the filter format)
$ prolific audience count -t /path/to/filters.json -w <workspace-id>

Count participants matching filters given directly as a flag
$ prolific audience count --filters '[{"filter_id":"age","selected_range":{"lower":18,"upper":65}}]' -w <workspace-id>

Emit machine-readable output for scripting
$ prolific audience count -t /path/to/filters.json -w <workspace-id> --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := in.validate(false); err != nil {
				return err
			}

			count, err := getCount(client, in)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			rendered, err := RenderCount(count, asJSON)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			fmt.Fprintln(w, rendered)

			return nil
		},
	}

	addFilterFlags(cmd, &in, false)
	// -j is bound by hand rather than through shared.AddOutputFlags, which
	// would claim -t for --table and collide with --template-path.
	cmd.Flags().BoolVarP(&asJSON, "json", "j", false, "Output as JSON")

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

// RenderCount produces output for an eligibility count. A zero is printed as
// 0, matching the API and audience breakdown.
func RenderCount(count int, asJSON bool) (string, error) {
	if asJSON {
		payload, err := json.Marshal(CountResult{Count: count})
		if err != nil {
			return "", err
		}

		return string(payload), nil
	}

	return fmt.Sprintf("Eligible participants: %d", count), nil
}
