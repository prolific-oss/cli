package project

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// ViewOptions is the options for the detail view of a project.
type ViewOptions struct {
	Args []string
}

// NewViewCommand creates a new command to show a project.
func NewViewCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Args:  cobra.MinimumNArgs(1),
		Short: "Provide details about your project",
		Long: `View your project

A detailed view of how your project is configured.
`,
		Example: `
View the details of a specific project

$ prolific project view 6261321e223a605c7a4f7678
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderProject(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}

// renderProject will show your project
func renderProject(client client.API, opts ListOptions, w io.Writer) error {
	if len(opts.Args) < 1 || opts.Args[0] == "" {
		return errors.New("please provide a project ID")
	}

	project, err := client.GetProject(opts.Args[0])
	if err != nil {
		return err
	}

	content := fmt.Sprintln(ui.RenderHeading(project.Title))

	if project.Description != "" {
		content += fmt.Sprintf("%s\n", project.Description)
	}

	content += fmt.Sprintf("\nWorkspace:                 %v", project.Workspace)
	content += fmt.Sprintf("\nOwner:                     %v", project.Owner)
	content += fmt.Sprintf("\nNaivety distribution rate: %v", project.NaivetyDistributionRate)
	content += "\n"

	content += "\nUsers:"

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\n", "ID", "Name", "Email")
	for _, user := range project.Users {
		fmt.Fprintf(tw, "%s\t%s\t%v\n", user.ID, user.Name, user.Email)
	}

	fmt.Fprintln(w, content)
	_ = tw.Flush()

	fmt.Fprintln(w, ui.RenderApplicationLink("project", fmt.Sprintf("researcher/workspaces/projects/%s/", project.ID)))

	return nil
}
