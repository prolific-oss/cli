package message

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing messages command.
type ListOptions struct {
	Args         []string
	UserID       string
	CreatedAfter string
}

// NewListCommand creates a new command to deal with messages
func NewListCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "View all your messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderMessages(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.UserID, "user_id", "u", "", "Filter messages by user.")
	flags.StringVarP(&opts.CreatedAfter, "created_after", "c", "", "Filter messages created after a certain date (YYYY-MM-DD).")

	return cmd
}

// renderMessages will show your messages
func renderMessages(client client.API, opts ListOptions, w io.Writer) error {
	var userID *string
	if opts.UserID != "" {
		userID = &opts.UserID
	}

	var createdAfter *string
	if opts.CreatedAfter != "" {
		createdAfter = &opts.CreatedAfter
	}

	messages, err := client.GetMessages(userID, createdAfter)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", "Sender ID", "Study ID", "Channel ID", "Datetime Created", "Body")
	for _, message := range messages.Results {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			message.SenderID,
			message.StudyID,
			message.ChannelID,
			message.DatetimeCreated,
			message.Body,
		)
	}

	return tw.Flush()
}
