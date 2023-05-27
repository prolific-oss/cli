package project

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewProjectCommand creates a new `project` command
func NewProjectCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage and view your projects in a workspace",
		Long: `Manage your projects
Projects are a way to organise studies in a workspace.
`,
	}

	cmd.AddCommand(
		NewListCommand("list", client, w),
		NewCreateCommand("create", client, w),
	)
	return cmd
}
