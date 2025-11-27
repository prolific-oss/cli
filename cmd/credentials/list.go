package credentials

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// ListOptions are the options for listing credential pools
type ListOptions struct {
	WorkspaceID string
}

// NewListCommand creates a new `credentials list` command to list credential pools for a workspace
func NewListCommand(client client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List credential pools for a workspace",
		Long: `List all credential pools belonging to a specific workspace.

Each credential pool summary includes:
- Credential Pool ID
- Total number of credentials
- Number of available (unredeemed) credentials

Required:
- Workspace ID (-w/--workspace-id): The workspace to list credential pools for`,
		Example: `
List all credential pools for a workspace:
$ prolific credentials list -w <workspace_id>
$ prolific credentials list --workspace-id 507f1f77bcf86cd799439011`,
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := client.ListCredentialPools(opts.WorkspaceID)
			if err != nil {
				return err
			}

			if len(response.CredentialPools) == 0 {
				msg := fmt.Sprintf("No credential pools found for workspace %s", ui.Dim(opts.WorkspaceID))
				ui.WriteInfo(w, msg)
				return nil
			}

			fmt.Fprintf(w, "%s\n\n", ui.Bold(fmt.Sprintf("Credential Pools for workspace %s", opts.WorkspaceID)))
			for _, pool := range response.CredentialPools {
				fmt.Fprintf(w, "%s %s\n", ui.Info("Credential Pool ID:"), ui.Highlight(pool.CredentialPoolID))
				fmt.Fprintf(w, "  Total Credentials: %s\n", ui.Bold(fmt.Sprintf("%d", pool.TotalCredentials)))
				fmt.Fprintf(w, "  Available Credentials: %s\n\n", ui.Bold(fmt.Sprintf("%d", pool.AvailableCredentials)))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.WorkspaceID, "workspace-id", "w", "", "Workspace ID (required)")
	_ = cmd.MarkFlagRequired("workspace-id")

	return cmd
}
