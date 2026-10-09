package audience

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewEligibilityCountCommand keeps `eligibility-count` working after it was
// replaced by `audience count`: the old flags, messages and stdout, in front
// of the audience implementation rather than a second copy of it.
func NewEligibilityCountCommand(c client.API, w io.Writer) *cobra.Command {
	var in filterInput

	cmd := &cobra.Command{
		Use:        "eligibility-count",
		Short:      "Count participants matching a set of filters",
		Deprecated: "use `prolific audience count` instead.",
		Long: `Count how many participants would be eligible for a study defined by a
set of filters, without creating the study.

This command has been replaced by "prolific audience count", which takes the
same template with -p/--template-path, and also accepts filters directly with
--filters and machine-readable output with --json, --csv and --table.`,
		Example: `
Count participants matching the filters in a JSON/YAML file (see
"prolific study create --help" for the filter format)
$ prolific eligibility-count -t /path/to/filters.json -w <workspace-id>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// The old messages, which differ from the audience command's.
			if in.TemplatePath == "" {
				return fmt.Errorf("error: a filter template is required, use -t/--template-path")
			}
			if in.WorkspaceID == "" {
				return fmt.Errorf("error: workspace ID is required")
			}

			count, err := getCount(c, in)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			_, err = fmt.Fprintln(w, RenderEligibilityCount(count))
			return err
		},
	}

	flags := cmd.Flags()
	// -t is --template-path here and --table on audience count, so this
	// command keeps its own flags rather than sharing the registrars.
	flags.StringVarP(&in.TemplatePath, "template-path", "t", "", "Path to a YAML/JSON file containing the filters to count against (required).")
	flags.StringVarP(&in.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "The workspace ID to count eligible participants for (required).")

	return cmd
}

// RenderEligibilityCount produces the line this command has always printed,
// including the note that the API reports counts below 25 as zero.
func RenderEligibilityCount(count int) string {
	if count == 0 {
		return "Eligible participants: 0 (or fewer than 25 — exact counts under 25 aren't shown, to protect participant privacy)"
	}

	return fmt.Sprintf("Eligible participants: %d", count)
}
