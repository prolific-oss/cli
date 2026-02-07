package message

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// SendGroupOptions is the options for the send group message command.
type SendGroupOptions struct {
	GroupID string
	StudyID string
	Body    string
}

// NewSendGroupCommand creates a new command to send a message to a participant group
func NewSendGroupCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts SendGroupOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Send a message to a participant group",
		Long: `Send a message to all members of a participant group

The study ID is optional for this endpoint. If omitted, the message is
sent without association to a specific study.
`,
		Example: `
$ prolific message send-group -g group-id -b "Thanks for participating"
$ prolific message send-group -g group-id -s study-id -b "Thanks for participating"
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := sendGroupMessage(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.GroupID, "group", "g", "", "Specify the participant group ID.")
	flags.StringVarP(&opts.StudyID, "study", "s", "", "Specify the study to which the message relates (optional).")
	flags.StringVarP(&opts.Body, "body", "b", "", "Specify the body of message.")

	return cmd
}

func sendGroupMessage(c client.API, opts SendGroupOptions, w io.Writer) error {
	if opts.GroupID == "" {
		return fmt.Errorf("group is required")
	}

	if opts.Body == "" {
		return fmt.Errorf("body is required")
	}

	var studyID *string
	if opts.StudyID != "" {
		studyID = &opts.StudyID
	}

	err := c.SendGroupMessage(opts.GroupID, opts.Body, studyID)
	if err != nil {
		return err
	}

	displayStudyID := "N/A"
	if studyID != nil {
		displayStudyID = *studyID
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\n", "Group ID", "Study ID", "Body")
	fmt.Fprintf(tw, "%s\t%s\t%s\n",
		opts.GroupID,
		displayStudyID,
		opts.Body,
	)

	return tw.Flush()
}
