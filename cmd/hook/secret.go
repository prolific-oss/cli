package hook

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ListOptions is the options for the listing secrets command.
type ListSecretOptions struct {
	Args        []string
	WorkspaceID string
}

// NewListSecretCommand creates a new `hook secrets` command to give you details about
// your secrets.
func NewListSecretCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts ListSecretOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "List your hook secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			err := renderSecrets(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "Filter secrets by workspace.")

	return cmd
}

// CreateSecretOptions are the options for the create secret command.
type CreateSecretOptions struct {
	WorkspaceID     string
	DeleteOldSecret bool
}

// NewCreateSecretCommand creates a new `hook create-secret` command to generate a secret
// for verifying webhook request signatures. If a secret already exists for the workspace,
// it will be replaced.
func NewCreateSecretCommand(c client.API, w io.Writer) *cobra.Command {
	var opts CreateSecretOptions

	cmd := &cobra.Command{
		Use:   "create-secret",
		Short: "Create a hook secret",
		Long: `Generate a secret for verifying the request signature header of webhook payloads.

If a secret already exists for the workspace, it will be replaced with a new one.`,
		Example: `
Create a secret for a workspace:
$ prolific hook create-secret -w 63722982f9cc073ecc730f6b

For a non-interactive experience, you can use the --delete-old-secret flag to confirm deletion of the existing secret without being prompted:
$ prolific hook create-secret -w 63722982f9cc073ecc730f6b --delete-old-secret
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			confirmed, err := confirmSecretCreation(opts.DeleteOldSecret, cmd.InOrStdin(), w)
			if err != nil {
				return err
			}

			if !confirmed {
				fmt.Fprintln(w, "Secret creation cancelled.")
				return nil
			}

			secret, err := c.CreateHookSecret(client.CreateSecretPayload{
				WorkspaceID: opts.WorkspaceID,
			})
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			fmt.Fprintf(w, "Secret created successfully\n")
			fmt.Fprintf(w, "ID:           %s\n", secret.ID)
			fmt.Fprintf(w, "Secret:       %s\n", secret.Value)
			fmt.Fprintf(w, "Workspace ID: %s\n", secret.WorkspaceID)

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "The workspace to create the secret in (required).")
	flags.BoolVar(&opts.DeleteOldSecret, "delete-old-secret", false, "Confirm deletion of the existing secret without being prompted.")
	_ = cmd.MarkFlagRequired("workspace")

	return cmd
}

// confirmSecretCreation prompts the user to confirm the secret creation unless
// the deleteOldSecret flag was provided.
func confirmSecretCreation(deleteOldSecret bool, r io.Reader, w io.Writer) (bool, error) {
	if deleteOldSecret {
		return true, nil
	}

	fmt.Fprint(w, "This command will delete the old secret. Are you sure? (y/N): ")

	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes", nil
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading input: %w", err)
	}

	return false, nil
}

// renderSecrets will show all of the secrets in a given workspace.
func renderSecrets(c client.API, opts ListSecretOptions, w io.Writer) error {
	secrets, err := c.GetHookSecrets(opts.WorkspaceID)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\n", "ID", "Secret", "Workspace ID")
	for _, secret := range secrets.Results {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", secret.ID, secret.Value, secret.WorkspaceID)
	}

	return tw.Flush()
}
