package collection

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewCollectionCommand creates a new `collection` command
func NewCollectionCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collection",
		Short: "Manage and view your collections",
	}

	cmd.AddCommand(
		NewListCommand(client, w),
		NewGetCommand(client, w),
		NewCreateCollectionCommand(client, w),
		NewUpdateCommand(client, w),
	)
	return cmd
}
