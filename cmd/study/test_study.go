package study

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewTestStudyCommand creates a new `study test` command to create a test run
// of a study to validate configuration before going live.
func NewTestStudyCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test <study-id>",
		Short: "Create a test run of a study",
		Long: `Create a test run of a study to validate configuration before going live.

This allows you to verify that a study is correctly configured by creating a
test run without publishing the study to real participants.

Prerequisites:
  - The study must be in draft status.
  - At least one test participant must exist in the workspace.
    Create one with: prolific api POST /api/v1/researchers/participants/`,
		Example: `
To create a test run of a study:
$ prolific study test 64395e9c2332b8a59a65d51e`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := client.TestStudy(args[0])
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			fmt.Fprintf(w, "Test study %s created: %s\n", response.StudyID, response.StudyURL)

			return nil
		},
	}

	return cmd
}
