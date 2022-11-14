package workspace

import (
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// NewWorkspaceCommand creates a new `workspace` command
func NewWorkspaceCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage and view your workspaces",
	}

	cmd.AddCommand(
		NewListCommand("list", client, w),
		NewCreateCommand("create", client, w),
	)
	return cmd
}
