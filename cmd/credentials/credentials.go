package credentials

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewCredentialsCommand creates a new `credentials` command
func NewCredentialsCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "credentials",
		Short: "Manage credential pools",
		Long:  `Create and manage credential pools for studies that require authentication credentials`,
	}

	cmd.AddCommand(
		NewCreateCommand(client, w),
		NewUpdateCommand(client, w),
	)

	return cmd
}
