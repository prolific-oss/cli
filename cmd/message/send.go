package message

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// SendOptions is the options for the send message command.
type SendOptions struct {
	Args        []string
	RecipientID string
	StudyID     string
	Body        string
}

// NewSendCommand creates a new command to deal with sending a message
func NewSendCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts SendOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Send a message",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderMessage(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.RecipientID, "recipient_id", "r", "", "Specify the recipient.")
	flags.StringVarP(&opts.StudyID, "study_id", "s", "", "Specify the study to which the message relates.")
	flags.StringVarP(&opts.Body, "body", "b", "", "Specific the body of message.")

	if err := cmd.MarkFlagRequired("recipient_id"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("study_id"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("body"); err != nil {
		panic(err)
	}

	return cmd
}

// renderMessage will show your message
func renderMessage(client client.API, opts SendOptions, w io.Writer) error {
	err := client.SendMessage(opts.Body, opts.RecipientID, opts.StudyID)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\n", "Recipient ID", "Study ID", "Body")
	fmt.Fprintf(tw, "%s\t%s\t%s\n",
		opts.RecipientID,
		opts.StudyID,
		opts.Body,
	)

	return tw.Flush()
}
