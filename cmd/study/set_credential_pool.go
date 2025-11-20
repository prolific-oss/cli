package study

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	studyui "github.com/prolific-oss/cli/ui/study"
	"github.com/spf13/cobra"
)

// SetCredentialPoolOptions represents the options for the set-credential-pool command.
type SetCredentialPoolOptions struct {
	Args             []string
	CredentialPoolID string
}

// NewSetCredentialPoolCommand creates a new `study set-credential-pool` command to
// allow you to set or update the credential pool on a draft study.
func NewSetCredentialPoolCommand(client client.API, w io.Writer) *cobra.Command {
	var opts SetCredentialPoolOptions

	cmd := &cobra.Command{
		Use:   "set-credential-pool <study-id>",
		Short: "Set or update the credential pool on a draft study",
		Long: `Set or update the credential pool ID on a draft study.

This allows you to attach a credential pool to a study that was created without one,
or change the credential pool on an existing draft study.`,
		Example: `
$ prolific study set-credential-pool 64395e9c2332b8a59a65d51e -c 679271425fe00981084a5f58_a856d700-c495-11f0-adce-338d4126f6e8
$ prolific study set-credential-pool 64395e9c2332b8a59a65d51e --credential-pool-id 679271425fe00981084a5f58_a856d700-c495-11f0-adce-338d4126f6e8`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.CredentialPoolID == "" {
				return fmt.Errorf("credential pool ID is required")
			}

			updatedStudy, err := client.UpdateStudy(args[0], model.UpdateStudy{CredentialPoolID: opts.CredentialPoolID})
			if err != nil {
				return err
			}

			fmt.Fprintln(w, studyui.RenderStudy(*updatedStudy))

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.CredentialPoolID, "credential-pool-id", "c", "", "The credential pool ID to attach to the study (format: <workspace_id>_<credential_pool_id>)")
	_ = cmd.MarkFlagRequired("credential-pool-id")

	return cmd
}
