package project

import (
	"errors"
	"fmt"
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/model"
	"github.com/spf13/cobra"
)

// CreateOptions are the options to be able to create a project.
type CreateOptions struct {
	Args                    []string
	Title                   string
	Workspace               string
	NaivetyDistributionRate int32
}

// NewCreateCommand creates a new command for creating a project.
func NewCreateCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Create a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := createproject(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Title, "title", "t", "", "The title of the project.")
	flags.StringVarP(&opts.Workspace, "workspace", "w", "", "The ID of the workspace to create the project in.")
	flags.Int32VarP(&opts.NaivetyDistributionRate, "naivety", "n", 0, "The speed vs naivety value. 0 = speed, 1 = naive.")

	return cmd
}

// createproject will create a workspace for you
func createproject(client client.API, opts CreateOptions, w io.Writer) error {
	if opts.Title == "" {
		return errors.New("title is required")
	}

	if opts.Workspace == "" {
		return errors.New("workspace is required")
	}

	project := model.Project{
		Title:                   opts.Title,
		NaivetyDistributionRate: float64(opts.NaivetyDistributionRate),
	}

	record, err := client.CreateProject(opts.Workspace, project)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Created project: %s\n", record.ID)

	return nil
}
