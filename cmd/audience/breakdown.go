package audience

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// BreakdownOptions is the options for the breakdown command. Filters are
// provided either via TemplatePath, or via FiltersJSON/BreakdownJSON — not
// both.
type BreakdownOptions struct {
	TemplatePath  string
	FiltersJSON   string
	BreakdownJSON string
	WorkspaceID   string
}

// breakdownTemplate is the shape of the -t/--template-path file: the same
// flat list of base filters that `eligibility-count` accepts, plus a single
// breakdown_filter to split the resulting counts by. --filters/--breakdown
// populate the same fields directly, bypassing the file.
type breakdownTemplate struct {
	Filters         []model.Filter `mapstructure:"filters"`
	BreakdownFilter model.Filter   `mapstructure:"breakdown_filter"`
}

// NewBreakdownCommand creates a new `audience breakdown` command to count how
// many participants match a set of base filters, split by the values of a
// single breakdown filter.
func NewBreakdownCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BreakdownOptions

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

Or provide the filters directly as flags:
$ prolific audience breakdown \
    --filters '[{"filter_id":"age","selected_range":{"lower":18,"upper":65}}]' \
    --breakdown '{"filter_id":"handedness"}' \
    -w <workspace-id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			usingTemplate := opts.TemplatePath != ""
			usingFlags := opts.FiltersJSON != "" || opts.BreakdownJSON != ""

			switch {
			case usingTemplate && usingFlags:
				return fmt.Errorf("error: use either -t/--template-path or --filters/--breakdown, not both")
			case !usingTemplate && !usingFlags:
				return fmt.Errorf("error: provide filters via -t/--template-path or --filters/--breakdown")
			case usingFlags && opts.BreakdownJSON == "":
				return fmt.Errorf("error: --breakdown is required when using --filters")
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
	flags.StringVarP(&opts.TemplatePath, "template-path", "t", "", "Path to a YAML/JSON file containing the base filters and breakdown_filter to count against. Alternative to --filters/--breakdown.")
	flags.StringVar(&opts.FiltersJSON, "filters", "", `JSON array of base filters to count against, e.g. '[{"filter_id":"age","selected_range":{"lower":18,"upper":65}}]'. Optional; alternative to -t/--template-path.`)
	flags.StringVar(&opts.BreakdownJSON, "breakdown", "", `JSON object for the single filter to break results down by, e.g. '{"filter_id":"handedness"}'. Required with --filters; alternative to -t/--template-path.`)
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "The workspace ID to count eligible participants for (required).")

	return cmd
}

func getBreakdown(c client.API, opts BreakdownOptions) (map[string]int, error) {
	var tmpl breakdownTemplate

	if opts.TemplatePath != "" {
		v := viper.New()
		v.SetConfigFile(opts.TemplatePath)
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
		if err := v.Unmarshal(&tmpl); err != nil {
			return nil, fmt.Errorf("unable to map %s to filters: %s", opts.TemplatePath, err)
		}
	} else {
		if opts.FiltersJSON != "" {
			if err := json.Unmarshal([]byte(opts.FiltersJSON), &tmpl.Filters); err != nil {
				return nil, fmt.Errorf("unable to parse --filters as JSON: %s", err)
			}
		}
		if err := json.Unmarshal([]byte(opts.BreakdownJSON), &tmpl.BreakdownFilter); err != nil {
			return nil, fmt.Errorf("unable to parse --breakdown as JSON: %s", err)
		}
	}

	if tmpl.BreakdownFilter.FilterID == "" {
		return nil, fmt.Errorf("breakdown filter must include a filter_id")
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
// counts per breakdown value, sorted alphabetically with
// client.FilterBreakdownNAKey always shown last.
func RenderBreakdown(breakdown map[string]int) string {
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

	return buf.String()
}
