package message

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// BulkSendOptions is the options for the bulk send message command.
type BulkSendOptions struct {
	IDs     string
	StudyID string
	Body    string
}

// NewBulkSendCommand creates a new command to send a message to multiple participants
func NewBulkSendCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts BulkSendOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Send a message to multiple participants",
		Long: `Send a message to multiple participants at once

Provide a comma-separated list of participant IDs to send the same message
to all of them.
`,
		Example: `
$ prolific message bulk-send -i id1,id2,id3 -s study-id -b "Thanks for participating"
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := bulkSendMessage(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.IDs, "ids", "i", "", "Comma-separated list of participant IDs.")
	flags.StringVarP(&opts.StudyID, "study", "s", "", "Specify the study to which the message relates.")
	flags.StringVarP(&opts.Body, "body", "b", "", "Specify the body of message.")

	return cmd
}

func bulkSendMessage(c client.API, opts BulkSendOptions, w io.Writer) error {
	ids := splitAndTrim(opts.IDs)
	if len(ids) == 0 {
		return fmt.Errorf("at least one participant ID is required")
	}

	err := c.BulkSendMessage(ids, opts.Body, opts.StudyID)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\n", "Recipients", "Study ID", "Body")
	fmt.Fprintf(tw, "%d\t%s\t%s\n",
		len(ids),
		opts.StudyID,
		opts.Body,
	)

	return tw.Flush()
}

func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
