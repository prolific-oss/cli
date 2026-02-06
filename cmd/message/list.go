package message

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
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
		Long: `Retrieve your messages from the Prolific Platform

This command will allow you to send and retrieve messages on the Prolific
Platform. Please note that if you retrieve messages via the CLI, the notification
count is not updated in the web application.
`,
		Example: `
If you want to see all the messages between you and another user, you can provide
their user ID
$ prolific message list -u 6262a15c0c745235a82a150c

If, however, you want to see all messages in the last 30 days (or less), you can
run
$ prolific message list -c 2023-05-01

You can also return unread messages. Please note, that if you call this command,
it will not mark the messages as read in the web application.
$ prolific message list -U
`,
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

	flags.StringVarP(&opts.UserID, "user", "u", "", "Filter messages sent to user.")
	flags.StringVarP(&opts.CreatedAfter, "created_after", "c", "", "Filter messages created after a certain date (YYYY-MM-DD). You can only fetch up to the last 30 days of messages.")
	flags.BoolVarP(&opts.Unread, "unread", "U", false, "Filter messages to show only unread. Cannot be used with any other flags.")

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

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)

	var results []model.Message

	if opts.Unread {
		messages, err := c.GetUnreadMessages()
		if err != nil {
			return err
		}
		results = messages.Results
	} else {
		messages, err := c.GetMessages(userID, createdAfter)
		if err != nil {
			return err
		}
		results = messages.Results
	}

	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", "ID", "Sender ID", "Study ID", "Category", "Created", "Body")
	for _, msg := range results {
		studyID := ""
		category := ""
		if msg.Data != nil {
			studyID = msg.Data.StudyID
			category = msg.Data.Category
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			msg.ID,
			msg.GetSenderID(),
			studyID,
			category,
			msg.DatetimeCreated.Format(ui.AppDateTimeFormat),
			msg.Body,
		)
	}

	_ = tw.Flush()

	fmt.Fprintln(w, ui.RenderApplicationLink("messages", "messages/inbox"))

	return nil
}
