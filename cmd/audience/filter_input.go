package audience

import (
	"encoding/json"
	"fmt"

	"github.com/prolific-oss/cli/model"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// filterInput holds the raw -t/--template-path, --filters, --breakdown, and
// -w/--workspace flag values shared by every `audience` command that counts
// against a set of filters.
type filterInput struct {
	TemplatePath  string
	FiltersJSON   string
	BreakdownJSON string
	WorkspaceID   string
}

// filterSpec is the shape every -t/--template-path file (and the
// --filters/--breakdown flags) resolve to: base filters (including nested groups), the
// same format `study create` accepts, plus an optional single breakdown
// filter to split results by.
type filterSpec struct {
	Filters         []model.Filter `mapstructure:"filters"`
	BreakdownFilter model.Filter   `mapstructure:"breakdown_filter"`
}

// addFilterFlags registers the shared -t/--template-path, --filters, and
// -w/--workspace flags on cmd, bound to in. withBreakdown also registers
// --breakdown, for commands that split results by a second filter.
func addFilterFlags(cmd *cobra.Command, in *filterInput, withBreakdown bool) {
	flags := cmd.Flags()

	templateHelp := "Path to a YAML/JSON file containing the filters to count against. Alternative to --filters."
	filtersHelp := `JSON array of filters to count against, e.g. '[{"filter_id":"age","selected_range":{"lower":18,"upper":65}}]'. Alternative to -t/--template-path.`
	if withBreakdown {
		templateHelp = "Path to a YAML/JSON file containing the base filters and breakdown_filter to count against. Alternative to --filters/--breakdown."
		filtersHelp = `JSON array of base filters to count against, e.g. '[{"filter_id":"age","selected_range":{"lower":18,"upper":65}}]'. Optional; alternative to -t/--template-path.`
	}

	flags.StringVarP(&in.TemplatePath, "template-path", "t", "", templateHelp)
	flags.StringVar(&in.FiltersJSON, "filters", "", filtersHelp)
	if withBreakdown {
		flags.StringVar(&in.BreakdownJSON, "breakdown", "", `JSON object for the single filter to break results down by, e.g. '{"filter_id":"handedness","selected_values":["0","1"]}' for a choice filter (selected_values is required) or '{"filter_id":"age","selected_range":{"lower":18,"upper":65}}' for a range filter. Required with --filters; alternative to -t/--template-path.`)
	}
	flags.StringVarP(&in.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "The workspace ID to count eligible participants for (required).")
}

// validate checks that exactly one of a template file or --filters was
// given, and — when requireBreakdown is set — that --breakdown accompanies
// --filters.
func (in filterInput) validate(requireBreakdown bool) error {
	usingTemplate := in.TemplatePath != ""
	usingFlags := in.FiltersJSON != "" || in.BreakdownJSON != ""

	flagsLabel := "--filters"
	if requireBreakdown {
		flagsLabel = "--filters/--breakdown"
	}

	switch {
	case usingTemplate && usingFlags:
		return fmt.Errorf("error: use either -t/--template-path or %s, not both", flagsLabel)
	case !usingTemplate && !usingFlags:
		return fmt.Errorf("error: provide filters via -t/--template-path or %s", flagsLabel)
	case requireBreakdown && usingFlags && in.BreakdownJSON == "":
		return fmt.Errorf("error: --breakdown is required when using --filters")
	}

	if in.WorkspaceID == "" {
		return fmt.Errorf("error: workspace ID is required")
	}

	return nil
}

// resolve loads a filterSpec from either the -t/--template-path file or the
// --filters/--breakdown flags, defaulting Filters to an empty (non-nil)
// slice, since the API rejects a null filters field.
func (in filterInput) resolve() (filterSpec, error) {
	var spec filterSpec

	if in.TemplatePath != "" {
		v := viper.New()
		v.SetConfigFile(in.TemplatePath)
		if err := v.ReadInConfig(); err != nil {
			return spec, err
		}
		if err := v.Unmarshal(&spec); err != nil {
			return spec, fmt.Errorf("unable to map %s to filters: %s", in.TemplatePath, err)
		}
	} else {
		if in.FiltersJSON != "" {
			if err := json.Unmarshal([]byte(in.FiltersJSON), &spec.Filters); err != nil {
				return spec, fmt.Errorf("unable to parse --filters as JSON: %s", err)
			}
		}
		if in.BreakdownJSON != "" {
			if err := json.Unmarshal([]byte(in.BreakdownJSON), &spec.BreakdownFilter); err != nil {
				return spec, fmt.Errorf("unable to parse --breakdown as JSON: %s", err)
			}
		}
	}

	if spec.Filters == nil {
		spec.Filters = []model.Filter{}
	}

	return spec, nil
}
