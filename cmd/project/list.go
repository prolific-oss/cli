package project

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ListOptions is the options for the listing projects command.
type ListOptions struct {
	Args        []string
	WorkspaceID string
	Limit       int
	Offset      int
}

// NewListCommand creates a new command to deal with projects
func NewListCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide details about your projects",
		Long: `List your projects

Studies are assigned to projects. This allows you to view all of the studies in
your workspaces.
		`,
		Example: `
List your projects in a given workspace
$ prolific project list -w 61a65c06b084910b3f0c00d5

Utilise the paging options to limit your projects, for example one project
$ prolific project list -w 61a65c06b084910b3f0c00d5 -l 1

Offset records in the result set, for example by 2
$ prolific project list -w 61a65c06b084910b3f0c00d5 -l 1 -o 2
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderProjects(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "Filter projects by workspace.")
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of workspaces returned")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of workspaces to offset")

	return cmd
}

// renderProjects will show your projects
func renderProjects(client client.API, opts ListOptions, w io.Writer) error {
	if opts.WorkspaceID == "" {
		return errors.New("please provide a workspace ID")
	}

	projects, err := client.GetProjects(opts.WorkspaceID, opts.Limit, opts.Offset)
	if err != nil {
		return err
	}

	count := 0
	if projects.JSONAPIMeta != nil {
		count = projects.Meta.Count
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\n", "ID", "Title", "Description")
	for _, project := range projects.Results {
		fmt.Fprintf(tw, "%s\t%s\t%v\n", project.ID, project.Title, project.Description)
	}

	_ = tw.Flush()

	fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(projects.Results), count))

	return nil
}
