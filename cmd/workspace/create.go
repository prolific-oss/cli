package workspace

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// CreateOptions are the options to be able to create a workspace.
type CreateOptions struct {
	Args  []string
	Title string
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
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := createWorkspace(client, opts, w)
			if err != nil {
				return err
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Title, "title", "t", "", "The title of the workspace.")

	return cmd
}

// createWorkspace will create a workspace for you
func createWorkspace(client client.API, opts CreateOptions, w io.Writer) error {
	if opts.Title == "" {
		return errors.New("title is required")
	}

	workspace := model.Workspace{
		Title: opts.Title,
	}

	record, err := client.CreateWorkspace(workspace)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Created workspace: %s", ui.Dim(record.ID))
	ui.WriteSuccess(w, msg)

	return nil
}
