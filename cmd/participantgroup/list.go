package participantgroup

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing participant groups command.
type ListOptions struct {
	Args      []string
	ProjectID string
	Limit     int
	Offset    int
}

// NewListCommand creates a new command to deal with participant groups
func NewListCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide details about your participant groups",
		Long: `List your participant groups

Participant groups are assigned to a project within your workspace.
`,
		Example: `
List the participant groups you have defined in a given project

$ prolific participant list -p 6261321e223a605c7a4f7623
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := render(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.ProjectID, "project", "p", "", "Filter participant groups by project.")
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of participant groups returned")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of participant groups to offset")

	return cmd
}

// render will list your participant groups
func render(client client.API, opts ListOptions, w io.Writer) error {
	if opts.ProjectID == "" {
		return errors.New("please provide a project ID")
	}

	groups, err := client.GetParticipantGroups(opts.ProjectID, opts.Limit, opts.Offset)
	if err != nil {
		return err
	}

	count := 0
	if groups.JSONAPIMeta != nil {
		count = groups.Meta.Count
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\n", "ID", "Name")
	for _, group := range groups.Results {
		fmt.Fprintf(tw, "%s\t%s\n", group.ID, group.Name)
	}

	_ = tw.Flush()

	fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(groups.Results), count))

	return nil
}
