package submission

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var submissionTransitionActions = []string{
	"APPROVE",
	"COMPLETE",
	"REJECT",
	"RETURN",
	"SCREEN_OUT",
	"START",
	"UNREJECT",
	"UNRETURN",
}

var rejectionCategories = []string{
	"TOO_QUICKLY",
	"TOO_SLOWLY",
	"FAILED_INSTRUCTIONS",
	"INCOMP_LONGITUDINAL",
	"FAILED_CHECK",
	"LOW_EFFORT",
	"MALINGERING",
	"NO_CODE",
	"BAD_CODE",
	"NO_DATA",
	"UNSUPP_DEVICE",
	"OTHER",
}

// TransitionOptions is the options for transitioning a submission.
type TransitionOptions struct {
	SubmissionID         string
	Action               string
	Message              string
	RejectionCategory    string
	CompletionCode       string
	PercentageOfReward   float64
	MessageToParticipant string
}

// NewTransitionCommand creates a new `submission transition` command.
func NewTransitionCommand(c client.API, w io.Writer) *cobra.Command {
	var opts TransitionOptions

	cmd := &cobra.Command{
		Use:   "transition",
		Short: "Transition the status of a submission",
		Long: `Transition a submission to a new state.

You can approve, complete, reject, return, or perform other state transitions on a submission.

When rejecting a submission, you must provide a message (at least 100 characters)
and a rejection category.

When completing a submission, you must provide a completion code. If the completion
code has a DYNAMIC_PAYMENT action, you must also provide the percentage of reward.`,
		Example: `  prolific submission transition <submission_id> -a APPROVE
  prolific submission transition <submission_id> -a REJECT -m "Your response did not follow the instructions provided in the study description" -R FAILED_INSTRUCTIONS
  prolific submission transition <submission_id> -a COMPLETE --completion-code MY_CODE
  prolific submission transition <submission_id> -a COMPLETE --completion-code MY_CODE --percentage-of-reward 50`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SubmissionID = args[0]

			err := transitionSubmission(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Action, "action", "a", "", fmt.Sprintf("Action to perform on the submission (%s)", strings.Join(submissionTransitionActions, ", ")))
	flags.StringVarP(&opts.Message, "message", "m", "", "Message to send to the participant (required for REJECT, min 100 characters)")
	flags.StringVarP(&opts.RejectionCategory, "rejection-category", "R", "", fmt.Sprintf("Rejection category (required for REJECT): %s", strings.Join(rejectionCategories, ", ")))
	flags.StringVar(&opts.CompletionCode, "completion-code", "", "Completion code (required for COMPLETE)")
	flags.Float64Var(&opts.PercentageOfReward, "percentage-of-reward", 0, "Percentage of reward for dynamic payment (8-99, used with COMPLETE)")
	flags.StringVar(&opts.MessageToParticipant, "message-to-participant", "", "Message to participant for dynamic payment (used with COMPLETE)")

	_ = cmd.MarkFlagRequired("action")

	return cmd
}

func transitionSubmission(c client.API, opts TransitionOptions, w io.Writer) error {
	if !slices.Contains(submissionTransitionActions, opts.Action) {
		return fmt.Errorf("invalid action %q, must be one of: %s", opts.Action, strings.Join(submissionTransitionActions, ", "))
	}

	if opts.Action == "REJECT" {
		if opts.Message == "" {
			return fmt.Errorf("message is required when rejecting a submission")
		}
		if opts.RejectionCategory == "" {
			return fmt.Errorf("rejection-category is required when rejecting a submission")
		}
		if !slices.Contains(rejectionCategories, opts.RejectionCategory) {
			return fmt.Errorf("invalid rejection category %q, must be one of: %s", opts.RejectionCategory, strings.Join(rejectionCategories, ", "))
		}
	}

	if opts.Action == "COMPLETE" {
		if opts.CompletionCode == "" {
			return fmt.Errorf("completion-code is required when completing a submission")
		}
	}

	payload := client.TransitionSubmissionPayload{
		Action:            opts.Action,
		Message:           opts.Message,
		RejectionCategory: opts.RejectionCategory,
		CompletionCode:    opts.CompletionCode,
	}

	if opts.PercentageOfReward > 0 {
		payload.CompletionCodeData = &client.CompletionCodeData{
			PercentageOfReward:   opts.PercentageOfReward,
			MessageToParticipant: opts.MessageToParticipant,
		}
	}

	response, err := c.TransitionSubmission(opts.SubmissionID, payload)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", "ID", "Study", "Participant", "Status")
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", response.ID, response.StudyID, response.Participant, response.Status)

	return tw.Flush()
}
