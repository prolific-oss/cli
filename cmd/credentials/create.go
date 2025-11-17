package credentials

import (
	"fmt"
	"io"
	"os"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// CreateOptions are the options for creating a credential pool
type CreateOptions struct {
	FilePath    string
	Credentials string
}

// NewCreateCommand creates a new `credentials create` command to create a credential pool
func NewCreateCommand(client client.API, w io.Writer) *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new credential pool",
		Long: `Create a new credential pool with comma-separated credentials

Credentials should be provided as comma-separated values with newlines between entries.
You can provide credentials directly as an argument or from a file.`,
		Example: `
Create a credential pool from a string:
$ prolific credentials create "user1,pass1\nuser2,pass2\nuser3,pass3"

Create a credential pool from a file:
$ prolific credentials create -f credentials.csv

File format example (credentials.csv):
user1,pass1
user2,pass2
user3,pass3`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var credentials string

			if opts.FilePath != "" {
				// Read from file
				data, err := os.ReadFile(opts.FilePath)
				if err != nil {
					return fmt.Errorf("unable to read file: %w", err)
				}
				credentials = string(data)
			} else if len(args) > 0 {
				// Use provided argument
				credentials = args[0]
			} else {
				return fmt.Errorf("credentials must be provided either as an argument or via -f flag")
			}

			response, err := client.CreateCredentialPool(credentials)
			if err != nil {
				return err
			}

			fmt.Fprintf(w, "Credential pool created successfully\n")
			fmt.Fprintf(w, "Credential Pool ID: %s\n", response.CredentialPoolID)

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.FilePath, "file", "f", "", "Path to file containing credentials")

	return cmd
}
