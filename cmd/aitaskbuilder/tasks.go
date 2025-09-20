package aitaskbuilder

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewTaskCommand creates a new `task` command under aitaskbuilder
func NewTaskCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "AI Task Builder task operations",
		Long:  "Manage AI Task Builder tasks - retrieve task responses and manage task-related operations",
	}

	cmd.AddCommand(
		NewGetResponsesCommand(client, w),
	)

	return cmd
}
