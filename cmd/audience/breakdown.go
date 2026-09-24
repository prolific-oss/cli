package audience

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// BreakdownOptions is the options for the breakdown command.
type BreakdownOptions struct {
	TemplatePath string
	WorkspaceID  string
}

// breakdownTemplate is the shape of the -t/--template-path file: the same
// flat list of base filters that `eligibility-count` accepts, plus a single
// breakdown_filter to split the resulting counts by.
type breakdownTemplate struct {
	Filters         []model.Filter `mapstructure:"filters"`
	BreakdownFilter model.Filter   `mapstructure:"breakdown_filter"`
}

// NewBreakdownCommand creates a new `audience breakdown` command to count how
// many participants match a set of base filters, broken down by the values
// of a single distributable filter.
func NewBreakdownCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BreakdownOptions

	cmd := &cobra.Command{
		Use:   "breakdown",
		Short: "Count eligible participants broken down by a filter",
		Long: `Count how many participants would be eligible for a study defined by a
set of base filters, broken down by the values (or bucketed ranges, for
numeric filters) of a single distributable breakdown filter.

The result includes an "N/A" count: participants who match the base
filters but don't fall into any of the breakdown filter's selected values
or range — for example, because they haven't answered that screener
question, or their answer falls outside what you specified.`,
		Example: `
Count participants matching the base filters and breakdown_filter in a
JSON/YAML file (see "prolific study create --help" for the filter format):
$ prolific audience breakdown -t /path/to/filters.json -w <workspace-id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.TemplatePath == "" {
				return fmt.Errorf("error: a filter template is required, use -t/--template-path")
			}

			if opts.WorkspaceID == "" {
				return fmt.Errorf("error: workspace ID is required")
			}

			breakdown, err := getBreakdown(client, opts)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			fmt.Fprint(w, RenderBreakdown(breakdown))

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.TemplatePath, "template-path", "t", "", "Path to a YAML/JSON file containing the base filters and breakdown_filter to count against (required).")
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "The workspace ID to count eligible participants for (required).")

	return cmd
}

func getBreakdown(c client.API, opts BreakdownOptions) (map[string]int, error) {
	v := viper.New()
	v.SetConfigFile(opts.TemplatePath)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var tmpl breakdownTemplate
	if err := v.Unmarshal(&tmpl); err != nil {
		return nil, fmt.Errorf("unable to map %s to filters: %s", opts.TemplatePath, err)
	}

	if tmpl.BreakdownFilter.FilterID == "" {
		return nil, fmt.Errorf("template must include a breakdown_filter with a filter_id")
	}

	// The API requires "filters" to be present and non-null, even when empty.
	if tmpl.Filters == nil {
		tmpl.Filters = []model.Filter{}
	}

	response, err := c.GetFilterBreakdown(client.FilterBreakdownPayload{
		Filters:         tmpl.Filters,
		BreakdownFilter: tmpl.BreakdownFilter,
		WorkspaceID:     opts.WorkspaceID,
	})
	if err != nil {
		return nil, err
	}

	return response.Breakdown, nil
}

// RenderBreakdown produces a human-readable table of eligible participant
// counts per breakdown value, sorted alphabetically with "N/A" (participants
// who don't match any bucket) always shown last.
func RenderBreakdown(breakdown map[string]int) string {
	keys := make([]string, 0, len(breakdown))
	for key := range breakdown {
		if key != "N/A" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if _, ok := breakdown["N/A"]; ok {
		keys = append(keys, "N/A")
	}

	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "VALUE\tCOUNT")
	for _, key := range keys {
		fmt.Fprintf(tw, "%s\t%d\n", key, breakdown[key])
	}
	tw.Flush()

	return buf.String()
}
