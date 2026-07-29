package submission

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// RequestReturnOptions is the options for the request return command.
type RequestReturnOptions struct {
	SubmissionID string
	Reasons      []string
}

// NewRequestReturnCommand creates a new `submission request-return` command to request
// a participant to return a submission.
func NewRequestReturnCommand(client client.API, w io.Writer) *cobra.Command {
	var opts RequestReturnOptions

	cmd := &cobra.Command{
		Use:   "request-return",
		Short: "Request a participant to return a submission",
		Long: `Request a participant to return a submission.

This is an experimental feature that allows researchers to ask a participant to
return a submission. The return reason must be provided and can be any free text.

Common reasons include:
  - Didn't finish the study
  - Encountered technical problems
  - Withdrew consent`,
		Example: `  prolific submission request-return <submission-id> -r "Didn't finish the study"
  prolific submission request-return <submission-id> -r "Encountered technical problems" -r "Withdrew consent"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SubmissionID = args[0]

			err := requestReturn(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringArrayVarP(&opts.Reasons, "reason", "r", nil, "Reason for requesting return (can be specified multiple times)")
	_ = cmd.MarkFlagRequired("reason")

	return cmd
}

func requestReturn(client client.API, opts RequestReturnOptions, w io.Writer) error {
	response, err := client.RequestSubmissionReturn(opts.SubmissionID, opts.Reasons)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", "ID", "Status", "Participant", "Return Requested")
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
		response.ID,
		response.Status,
		response.Participant,
		formatReturnRequested(response.ReturnRequested),
	)

	return tw.Flush()
}

func formatReturnRequested(t *string) string {
	if t == nil {
		return "-"
	}
	return *t
}
