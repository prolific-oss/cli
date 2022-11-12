package project

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing projects command.
type ListOptions struct {
	Args        []string
	WorkspaceID string
}

// NewListCommand creates a new command to deal with projects
func NewListCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide details about your projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderProjects(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", "", "Filter on enabled subscriptions.")

	return cmd
}

// renderProjects will show your projects
func renderProjects(client client.API, opts ListOptions, w io.Writer) error {
	if opts.WorkspaceID == "" {
		return errors.New("please provide a workspace ID")
	}

	projects, err := client.GetProjects(opts.WorkspaceID)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\n", "ID", "Title", "Description")
	for _, project := range projects.Results {
		fmt.Fprintf(tw, "%s\t%s\t%v\n", project.ID, project.Title, project.Description)
	}

	tw.Flush()

	return nil
}
