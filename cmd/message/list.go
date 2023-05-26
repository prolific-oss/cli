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
	Unread       bool
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

	flags.StringVarP(&opts.UserID, "id", "i", "", "Filter messages by user.")
	flags.StringVarP(&opts.CreatedAfter, "created_after", "c", "", "Filter messages created after a certain date (YYYY-MM-DD).")
	flags.BoolVarP(&opts.Unread, "unread", "u", false, "Filter messages to show only unread. Cannot be used with any other flags.")

	return cmd
}

// renderMessages will show your messages
func renderMessages(c client.API, opts ListOptions, w io.Writer) error {
	if opts.Unread && (opts.UserID != "" || opts.CreatedAfter != "") {
		return fmt.Errorf("'unread' cannot be used with any other flags")
	}

	var userID *string
	if opts.UserID != "" {
		userID = &opts.UserID
	}

	var createdAfter *string
	if opts.CreatedAfter != "" {
		createdAfter = &opts.CreatedAfter
	}

	if opts.Unread {
		messages, err := c.GetUnreadMessages()
		if err != nil {
			return err
		}

		tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", "Sender ID", "Channel ID", "Datetime Created", "Body")
		for _, message := range messages.Results {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
				message.Sender,
				message.ChannelID,
				message.DatetimeCreated,
				message.Body,
			)
		}

		return nil
	} else {
		messages, err := c.GetMessages(userID, createdAfter)

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

		_ = tw.Flush()
	}

	return nil
}
