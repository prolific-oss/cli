package workspace

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui"
	"github.com/spf13/cobra"
)

// WorkspaceListOptions is the options for the listing workspaces command.
type WorkspaceListOptions struct {
	Args   []string
	Limit  int
	Offset int
}

// NewListCommand creates a new command to deal with workspaces
func NewListCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts WorkspaceListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide details about your workspaces",
		Long: `List your workspaces

As a user, you can have many workspaces, maybe personally or for your organisation.
This allows you to list all of the workspaces your token has access to. Each
workspace then has one or many projects to organise your studies.
`,
		Example: `
List your workspaces
$ prolific workspace list

Utilise the paging options to limit your workspaces, for example one workspace
$ prolific workspace list -l 1

Offset records in the result set, for example by 2
$ prolific workspace list -l 1 -o 2
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderWorkspaces(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of workspaces returned")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of workspaces to offset")

	return cmd
}

// renderWorkspaces will show your workspaces
func renderWorkspaces(c client.API, opts WorkspaceListOptions, w io.Writer) error {
	workspaces, err := c.GetWorkspaces(opts.Limit, opts.Offset)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\n", "ID", "Title", "Description")
	for _, workspace := range workspaces.Results {
		fmt.Fprintf(tw, "%s\t%s\t%v\n", workspace.ID, workspace.Title, workspace.Description)
	}

	_ = tw.Flush()

	fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(workspaces.Results), workspaces.Meta.Count))

	return nil
}
