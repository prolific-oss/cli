package participantgroup

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/spf13/cobra"
)

// memberOptions are the options for changing the members of a participant group.
type memberOptions struct {
	Args           []string
	ParticipantIDs []string
	File           string
}

// memberCommand describes a command that changes the membership of a participant group.
type memberCommand struct {
	// short and long are the cobra descriptions.
	short string
	long  string
	// example is the cobra example text.
	example string
	// flagUsage is the usage text for --participant-id.
	flagUsage string
	// apply performs the API call for the change.
	apply func(c client.API, groupID string, participantIDs []string) (*client.ViewParticipantGroupResponse, error)
	// summary is a format taking the number changed, the group ID and the resulting total.
	summary string
}

func newMemberCommand(spec memberCommand, commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts memberOptions

	cmd := &cobra.Command{
		Use:     commandName + " <group-id>",
		Args:    cobra.ExactArgs(1),
		Short:   spec.short,
		Long:    spec.long,
		Example: spec.example,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if err := changeMembers(spec, c, opts, w); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringArrayVarP(&opts.ParticipantIDs, "participant-id", "p", nil, spec.flagUsage)
	flags.StringVarP(&opts.File, "file", "f", "", "Path to a file containing one participant ID per line.")

	return cmd
}

// changeMembers validates the participant IDs, applies the change and reports the outcome.
func changeMembers(spec memberCommand, c client.API, opts memberOptions, w io.Writer) error {
	groupID := opts.Args[0]

	hasInlineIDs := len(opts.ParticipantIDs) > 0
	hasFile := opts.File != ""

	if hasFile && hasInlineIDs {
		return fmt.Errorf("cannot use --file together with --participant-id")
	}

	if hasFile {
		ids, err := shared.ParseIDFile(opts.File)
		if err != nil {
			return err
		}
		opts.ParticipantIDs = ids
	}

	if len(opts.ParticipantIDs) == 0 {
		return fmt.Errorf("you must provide at least one participant ID via --participant-id or --file")
	}

	response, err := spec.apply(c, groupID, opts.ParticipantIDs)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, spec.summary+"\n", len(opts.ParticipantIDs), groupID, len(response.Results))

	return nil
}
