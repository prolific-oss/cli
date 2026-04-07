package submission

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/spf13/cobra"
)

// BulkApproveOptions is the options for bulk approving submissions.
type BulkApproveOptions struct {
	SubmissionIDs  []string
	StudyID        string
	ParticipantIDs []string
	File           string
}

// NewBulkApproveCommand creates a new `submission bulk-approve` command.
func NewBulkApproveCommand(c client.API, w io.Writer) *cobra.Command {
	var opts BulkApproveOptions

	cmd := &cobra.Command{
		Use:   "bulk-approve",
		Short: "Bulk approve multiple submissions",
		Long: `Bulk approve multiple submissions at once.

You can approve submissions in two ways:

1. By submission IDs: provide one or more --submission-id flags or a file with --file
2. By study and participant IDs: provide a --study flag with --participant-id flags or a file with --file

When using --file, the file should contain one ID per line. By default, IDs are
treated as submission IDs. Use --study together with --file to treat IDs as
participant IDs instead.

The approval is processed asynchronously.`,
		Example: `  # Approve by submission IDs
  prolific submission bulk-approve -i <submission_id> -i <submission_id>
  prolific submission bulk-approve -f submissions.csv

  # Approve by study and participant IDs
  prolific submission bulk-approve -s <study_id> -p <participant_id> -p <participant_id>
  prolific submission bulk-approve -s <study_id> -f participants.csv`,
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
	flags.StringVarP(&opts.StudyID, "study", "s", "", "Study ID (required with --participant-id, optional with --file)")
	flags.StringArrayVarP(&opts.ParticipantIDs, "participant-id", "p", nil, "Participant ID to approve (can be specified multiple times, requires --study)")
	flags.StringVarP(&opts.File, "file", "f", "", "Path to a file containing one ID per line")

	return cmd
}

func bulkApproveSubmissions(c client.API, opts BulkApproveOptions, w io.Writer) error {
	hasSubmissionIDs := len(opts.SubmissionIDs) > 0
	hasParticipantIDs := len(opts.ParticipantIDs) > 0
	hasFile := opts.File != ""

	if hasFile && (hasSubmissionIDs || hasParticipantIDs) {
		return fmt.Errorf("cannot use --file together with --submission-id or --participant-id")
	}

	if hasFile {
		ids, err := shared.ParseIDFile(opts.File)
		if err != nil {
			return err
		}
		if opts.StudyID != "" {
			opts.ParticipantIDs = ids
		} else {
			opts.SubmissionIDs = ids
		}
		hasSubmissionIDs = len(opts.SubmissionIDs) > 0
		hasParticipantIDs = len(opts.ParticipantIDs) > 0
	}

	if hasSubmissionIDs && (hasParticipantIDs || opts.StudyID != "") {
		return fmt.Errorf("cannot use --submission-id together with --study or --participant-id")
	}

	if !hasSubmissionIDs && !hasParticipantIDs {
		if opts.StudyID != "" {
			return fmt.Errorf("--participant-id or --file is required when using --study")
		}
		return fmt.Errorf("you must provide either --submission-id, --study and --participant-id, or --file")
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
