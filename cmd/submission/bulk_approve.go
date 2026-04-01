package submission

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// BulkApproveOptions is the options for bulk approving submissions.
type BulkApproveOptions struct {
	SubmissionIDs  []string
	StudyID        string
	ParticipantIDs []string
}

// NewBulkApproveCommand creates a new `submission bulk-approve` command.
func NewBulkApproveCommand(c client.API, w io.Writer) *cobra.Command {
	var opts BulkApproveOptions

	cmd := &cobra.Command{
		Use:   "bulk-approve",
		Short: "Bulk approve multiple submissions",
		Long: `Bulk approve multiple submissions at once.

You can approve submissions in two ways:

1. By submission IDs: provide one or more --submission-id flags
2. By study and participant IDs: provide a --study flag and one or more --participant-id flags

The approval is processed asynchronously.`,
		Example: `  prolific submission bulk-approve -i <submission_id> -i <submission_id>
  prolific submission bulk-approve -s <study_id> -p <participant_id> -p <participant_id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := bulkApproveSubmissions(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringArrayVarP(&opts.SubmissionIDs, "submission-id", "i", nil, "Submission ID to approve (can be specified multiple times)")
	flags.StringVarP(&opts.StudyID, "study", "s", "", "Study ID (required with --participant-id)")
	flags.StringArrayVarP(&opts.ParticipantIDs, "participant-id", "p", nil, "Participant ID to approve (can be specified multiple times, requires --study)")

	return cmd
}

func bulkApproveSubmissions(c client.API, opts BulkApproveOptions, w io.Writer) error {
	hasSubmissionIDs := len(opts.SubmissionIDs) > 0
	hasParticipantIDs := len(opts.ParticipantIDs) > 0

	if hasSubmissionIDs && (hasParticipantIDs || opts.StudyID != "") {
		return fmt.Errorf("cannot use --submission-id together with --study or --participant-id")
	}

	if !hasSubmissionIDs && !hasParticipantIDs {
		if opts.StudyID != "" {
			return fmt.Errorf("--participant-id is required when using --study")
		}
		return fmt.Errorf("you must provide either --submission-id or --study and --participant-id")
	}

	if hasParticipantIDs && opts.StudyID == "" {
		return fmt.Errorf("--study is required when using --participant-id")
	}

	payload := client.BulkApproveSubmissionsPayload{
		SubmissionIDs:  opts.SubmissionIDs,
		StudyID:        opts.StudyID,
		ParticipantIDs: opts.ParticipantIDs,
	}

	err := c.BulkApproveSubmissions(payload)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "The request to bulk approve has been made successfully.")

	return nil
}
