package credentials

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// UpdateOptions are the options for updating a credential pool
type UpdateOptions struct {
	FilePath    string
	Credentials string
}

// NewUpdateCommand creates a new `credentials update` command to update a credential pool
func NewUpdateCommand(client client.API, w io.Writer) *cobra.Command {
	var opts UpdateOptions

	cmd := &cobra.Command{
		Use:   "update <credential-pool-id> [credentials]",
		Short: "Update an existing credential pool",
		Long: `Update an existing credential pool with new comma-separated credentials

Credentials should be provided as comma-separated values with newlines between entries.
You can provide credentials directly as an argument or from a file.

Required:
- Credential Pool ID: The ID of the credential pool to update (positional argument)
- Credentials: Either as an argument or via the -f flag`,
		Example: `
Update a credential pool from a string:
$ prolific credentials update pool123 "user1,pass1\nuser2,pass2\nuser3,pass3"

Update a credential pool from a file:
$ prolific credentials update pool123 -f credentials.csv
$ prolific credentials update pool123 -f docs/examples/credentials.csv

File format example (credentials.csv):
user1@example.com,p4ssw0rd1
user2@example.com,p4ssw0rd2
user3@example.com,p4ssw0rd3`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			credentialPoolID := args[0]

			credentials, err := getCredentials(opts.FilePath, args, 1)
			if err != nil {
				return err
			}

			response, err := client.UpdateCredentialPool(credentialPoolID, credentials, "")
			if err != nil {
				return err
			}

			fmt.Fprintf(w, "Credential pool updated successfully\n")
			fmt.Fprintf(w, "Credential Pool ID: %s\n", response.CredentialPoolID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.FilePath, "file", "f", "", "Path to file containing credentials")

	return cmd
}
