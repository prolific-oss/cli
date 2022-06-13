package requirement

import (
	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// NewRequirementCommand creates a new `requirement` command
func NewRequirementCommand(client client.API) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "requirement",
		Short: "Requirement related commands",
	}

	cmd.AddCommand(
		NewListCommand(client),
	)
	return cmd
}
