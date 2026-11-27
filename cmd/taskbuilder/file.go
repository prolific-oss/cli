package taskbuilder

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewFileCommand creates a new `file` command
func NewFileCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Manage hosted files",
		Long:  "Upload and delete videos hosted for use in Task Builder content, such as custom HTML blocks",
	}

	cmd.AddCommand(
		NewFileUploadCommand(client, w),
		NewFileDeleteCommand(client, w),
	)

	return cmd
}
