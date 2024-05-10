package workspace

import (
	"errors"
	"fmt"
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/model"
	"github.com/spf13/cobra"
)

// CreateOptions are the options to be able to create a workspace.
type CreateOptions struct {
	Args                    []string
	Title                   string
	NaivetyDistributionRate int32
}

// NewCreateCommand creates a new command for creating workspaces.
func NewCreateCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Create a workspace",
		Long: `Create a workspace on your account

As a user, you can have many workspaces, maybe personally or for your organisation.
This allows you to create a workspace. Each workspace can then have one or more
projects to organise your studies.
`,
		Example: `
To create a workspace
$ prolific workspace create -t "Research into AI"

You can also configure if your studies should focus on speed, or participant
naivety. By specifying 0, your study will become available to all eligible
participants straight away. By specifying anything greater than 0, up to 1, you
are delaying your study for participants who are rather active on the Prolific
Platform. This can be overridden at the project level.
$ prolific workspace create -t "Research into AI" -n 0
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := createWorkspace(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Title, "title", "t", "", "The title of the workspace.")
	flags.Int32VarP(&opts.NaivetyDistributionRate, "naivety", "n", 0, "The speed vs naivety value. 0 = speed, 1 = naive.")

	return cmd
}

// createWorkspace will create a workspace for you
func createWorkspace(client client.API, opts CreateOptions, w io.Writer) error {
	if opts.Title == "" {
		return errors.New("title is required")
	}

	workspace := model.Workspace{
		Title:                   opts.Title,
		NaivetyDistributionRate: float64(opts.NaivetyDistributionRate),
	}

	record, err := client.CreateWorkspace(workspace)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Created workspace: %s\n", record.ID)

	return nil
}
