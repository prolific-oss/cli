package study

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewCredentialsReportCommand creates a new `study credentials-report` command to
// retrieve the credentials usage report for a study as CSV.
func NewCredentialsReportCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "credentials-report",
		Short: "Get the credentials usage report for a study",
		Long: `Get the credentials usage report for a study as CSV

This command retrieves a CSV report showing credential usage for a study,
including participant IDs, submission IDs, usernames, and status (USED/UNUSED).

Note: This endpoint is only available for studies that have credentials configured.
If the study does not have credentials configured, you will receive an error.`,
		Example: `
To get the credentials report for a study:
$ prolific study credentials-report 64395e9c2332b8a59a65d51e

To save the report to a file:
$ prolific study credentials-report 64395e9c2332b8a59a65d51e > credentials.csv`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			csvData, err := client.GetStudyCredentialsUsageReportCSV(args[0])
			if err != nil {
				return err
			}

			fmt.Fprint(w, csvData)

			return nil
		},
	}

	return cmd
}
