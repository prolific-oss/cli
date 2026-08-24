package eligibilitycount

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CountOptions is the options for the eligibility-count command.
type CountOptions struct {
	TemplatePath string
	WorkspaceID  string
}

// countTemplate is the shape of the -t/--template-path file: a flat list of
// filters, the same format `study create` accepts. Composite (and/or) filter
// groups, which the API also supports, are not represented here.
type countTemplate struct {
	Filters []model.Filter `mapstructure:"filters"`
}

// NewCountCommand creates a new `eligibility-count` command to count how many
// participants match a set of filters, without creating a study.
func NewCountCommand(client client.API, w io.Writer) *cobra.Command {
	var opts CountOptions

	cmd := &cobra.Command{
		Use:   "eligibility-count",
		Short: "Count participants matching a set of filters",
		Long: `Count how many participants would be eligible for a study defined by a
set of filters, without creating the study.

Counts below 25 are reported as 0 by the Prolific API, to protect
participant privacy. A count of 0 may mean either "zero eligible" or
"somewhere between 1 and 24 eligible" — the CLI cannot tell these apart.`,
		Example: `
Count participants matching the filters in a JSON/YAML file (see
"prolific study create --help" for the filter format)
$ prolific eligibility-count -t /path/to/filters.json -w <workspace-id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.TemplatePath == "" {
				return fmt.Errorf("error: a filter template is required, use -t/--template-path")
			}

			if opts.WorkspaceID == "" {
				return fmt.Errorf("error: workspace ID is required")
			}

			count, err := getEligibilityCount(client, opts)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			fmt.Fprintln(w, RenderCount(count))

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.TemplatePath, "template-path", "t", "", "Path to a YAML/JSON file containing the filters to count against")
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "The workspace ID to count eligible participants for (required).")

	return cmd
}

func getEligibilityCount(c client.API, opts CountOptions) (int, error) {
	v := viper.New()
	v.SetConfigFile(opts.TemplatePath)
	if err := v.ReadInConfig(); err != nil {
		return 0, err
	}

	var tmpl countTemplate
	if err := v.Unmarshal(&tmpl); err != nil {
		return 0, fmt.Errorf("unable to map %s to filters: %s", opts.TemplatePath, err)
	}

	// The API requires "filters" to be present and non-null, even when empty.
	if tmpl.Filters == nil {
		tmpl.Filters = []model.Filter{}
	}

	response, err := c.GetEligibilityCount(client.EligibilityCountPayload{
		Filters:     tmpl.Filters,
		WorkspaceID: opts.WorkspaceID,
	})
	if err != nil {
		return 0, err
	}

	return response.Count, nil
}

// RenderCount produces a human-readable line for an eligibility count,
// flagging the sub-25 privacy floor explicitly rather than presenting 0 as
// an exact count.
func RenderCount(count int) string {
	if count == 0 {
		return "Eligible participants: 0 (or fewer than 25 — exact counts under 25 aren't shown, to protect participant privacy)"
	}

	return fmt.Sprintf("Eligible participants: %d", count)
}
