package message

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewMessageCommand creates a new `message` command
func NewMessageCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "message",
		Short: "Send and retrieve messages",
	}

	cmd.AddCommand(
		NewListCommand("list", client, w),
		NewSendCommand("send", client, w),
		NewBulkSendCommand("bulk-send", client, w),
		NewSendGroupCommand("send-group", client, w),
	)

	return cmd
}
