package project

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
)

// CreateOptions are the options to be able to create a project.
type CreateOptions struct {
	Args      []string
	Title     string
	Workspace string
}

// NewCreateCommand creates a new command for creating a project.
func NewCreateCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Create a project",
		Long: `Create a project in your workspace

As a user, you can have many projects inside your workspace. You can then assign
studies to your projects, to neatly organise your work.`,
		Example: `
To create a project inside a workspace
$ prolific project create -t "Research into AI" -w 6261321e223a605c7a4f7564
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := createProject(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Title, "title", "t", "", "The title of the project.")
	flags.StringVarP(&opts.Workspace, "workspace", "w", "", "The ID of the workspace to create the project in.")

	return cmd
}

// createProject will create a project for you
func createProject(client client.API, opts CreateOptions, w io.Writer) error {
	if opts.Title == "" {
		return errors.New("title is required")
	}

	if opts.Workspace == "" {
		return errors.New("workspace is required")
	}

	project := model.Project{
		Title: opts.Title,
	}

	record, err := client.CreateProject(opts.Workspace, project)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Created project: %s\n", record.ID)

	return nil
}
