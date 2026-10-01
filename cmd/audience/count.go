package audience

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CountOptions is the options for the count command.
type CountOptions struct {
	TemplatePath string
	FiltersJSON  string
	WorkspaceID  string
	JSON         bool
}

// countTemplate is the shape of the -t/--template-path file: a flat list of
// filters, the same format `study create` accepts. Composite (and/or) filter
// groups, which the API also supports, are not represented here.
type countTemplate struct {
	Filters []model.Filter `mapstructure:"filters"`
}

// CountResult is the machine-readable shape emitted by --json.
type CountResult struct {
	Count int `json:"count"`
}

// NewCountCommand creates a new `audience count` command to count how many
// participants match a set of filters, without creating a study or saving a
// filter set.
func NewCountCommand(client client.API, w io.Writer) *cobra.Command {
	var opts CountOptions

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
			usingTemplate := opts.TemplatePath != ""
			usingFlags := opts.FiltersJSON != ""

			switch {
			case usingTemplate && usingFlags:
				return fmt.Errorf("error: use only one of -t/--template-path or --filters")
			case !usingTemplate && !usingFlags:
				return fmt.Errorf("error: a filter template or --filters is required, use -t/--template-path or --filters")
			}

			if opts.WorkspaceID == "" {
				return fmt.Errorf("error: workspace ID is required")
			}

			count, err := getCount(client, opts)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			rendered, err := RenderCount(count, opts.JSON)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			fmt.Fprintln(w, rendered)

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.TemplatePath, "template-path", "t", "", "Path to a YAML/JSON file containing the filters to count against.")
	flags.StringVar(&opts.FiltersJSON, "filters", "", `JSON array of filters to count against, e.g. '[{"filter_id":"age","selected_range":{"lower":18,"upper":65}}]'. Alternative to -t/--template-path.`)
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "The workspace ID to count eligible participants for (required).")
	// -j is bound by hand rather than through shared.AddOutputFlags, which
	// would claim -t for --table and collide with --template-path.
	flags.BoolVarP(&opts.JSON, "json", "j", false, "Output as JSON")

	return cmd
}

func getCount(c client.API, opts CountOptions) (int, error) {
	var tmpl countTemplate

	if opts.FiltersJSON != "" {
		if err := json.Unmarshal([]byte(opts.FiltersJSON), &tmpl.Filters); err != nil {
			return 0, fmt.Errorf("unable to parse --filters as JSON: %s", err)
		}
	} else {
		v := viper.New()
		v.SetConfigFile(opts.TemplatePath)
		if err := v.ReadInConfig(); err != nil {
			return 0, err
		}

		if err := v.Unmarshal(&tmpl); err != nil {
			return 0, fmt.Errorf("unable to map %s to filters: %s", opts.TemplatePath, err)
		}
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

// RenderCount produces output for an eligibility count. A zero is printed as 0,
// matching the API and audience breakdown.
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
