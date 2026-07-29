package study

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewDemographicExportCommand creates a new `study demographic-export` command to
// trigger a demographic data export for all submissions in a study.
func NewDemographicExportCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demographic-export <study-id>",
		Short: "Trigger a demographic data export for a study",
		Long: `Trigger a demographic data export for all submissions in a study.

This initiates an export of demographic data across all submissions within the
specified study. This is distinct from per-submission demographic retrieval.`,
		Example: `
To trigger a demographic export for a study:
$ prolific study demographic-export 64395e9c2332b8a59a65d51e`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			csvData, err := client.ExportDemographics(args[0])
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			fmt.Fprint(w, csvData)

			return nil
		},
	}

	return cmd
}
