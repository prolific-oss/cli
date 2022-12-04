package study

import (
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// NewStudyCommand creates a new `study` command
func NewStudyCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "study",
		Short: "Manage and view your studies",
	}

	cmd.AddCommand(
		NewListCommand("list", client, w),
		NewViewCommand(client, w),
		NewCreateCommand(client, w),
		NewDuplicateCommand(client, w),
	)
	return cmd
}
