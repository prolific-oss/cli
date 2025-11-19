package credentials

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// CreateOptions are the options for creating a credential pool
type CreateOptions struct {
	FilePath    string
	Credentials string
	WorkspaceID string
}

// NewCreateCommand creates a new `credentials create` command to create a credential pool
func NewCreateCommand(client client.API, w io.Writer) *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new credential pool",
		Long: `Create a new credential pool with comma-separated credentials

Credentials should be provided as comma-separated values with newlines between entries.
You can provide credentials directly as an argument or from a file.

Required:
- Workspace ID (-w/--workspace-id): The workspace where the credential pool will be created
- Credentials: Either as an argument or via the -f flag`,
		Example: `
Create a credential pool from a string:
$ prolific credentials create -w <workspace_id> "user1,pass1\nuser2,pass2\nuser3,pass3"

Create a credential pool from a file:
$ prolific credentials create -w <workspace_id> -f credentials.csv
$ prolific credentials create -w <workspace_id> -f docs/examples/credentials.csv

File format example (credentials.csv):
user1@example.com,p4ssw0rd1
user2@example.com,p4ssw0rd2
user3@example.com,p4ssw0rd3`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			credentials, err := getCredentials(opts.FilePath, args, 0)
			if err != nil {
				return err
			}

			response, err := client.CreateCredentialPool(credentials, opts.WorkspaceID)
			if err != nil {
				return err
			}

			fmt.Fprintf(w, "Credential pool created successfully\n")
			fmt.Fprintf(w, "Credential Pool ID: %s\n", response.CredentialPoolID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.FilePath, "file", "f", "", "Path to file containing credentials")
	cmd.Flags().StringVarP(&opts.WorkspaceID, "workspace-id", "w", "", "Workspace ID (required)")
	_ = cmd.MarkFlagRequired("workspace-id")

	return cmd
}
