package participantgroup

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing participant groups command.
type ListOptions struct {
	Args      []string
	ProjectID string
}

// NewListCommand creates a new command to deal with participant groups
func NewListCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide details about your participant groups",
		Long: `List your participant groups

Participant Groups allow you to create, modify, and use lists of participants
directly within the Prolific ecosystem.

Participant groups allow you do the following:

- Create a new participant group within the scope of a project.
- Add and remove users manually to / from the participant group.
- Use one or more participant groups as eligibility requirements for a new study.
- Combined with study completion codes, automatically add or remove participants
  from a group when they submit a response to your study with the correct code.
`,
		Example: `
List the participant groups you have defined in a given project

$ prolific participant list -p 6261321e223a605c7a4f7623
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := render(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.ProjectID, "project", "p", "", "Filter participant groups by project.")

	return cmd
}

// render will list your participant groups
func render(client client.API, opts ListOptions, w io.Writer) error {
	if opts.ProjectID == "" {
		return errors.New("please provide a project ID")
	}

	groups, err := client.GetParticipantGroups(opts.ProjectID)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\n", "ID", "Name")
	for _, group := range groups.Results {
		fmt.Fprintf(tw, "%s\t%s\n", group.ID, group.Name)
	}

	return tw.Flush()
}
