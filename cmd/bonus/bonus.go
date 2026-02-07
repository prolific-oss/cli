package bonus

import (
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewBonusCommand creates a new `bonus` command
func NewBonusCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bonus",
		Short: "Create and pay bonuses for study participants",
		Long: `Create and pay bonus payments for study participants.

The bonus workflow is two steps: create bonus records with cost breakdown,
then pay them. Non-interactive mode (-n) outputs machine-readable format
suitable for scripted pipelines.`,
		Example: `  # Create and review bonus costs interactively
  prolific bonus create <study_id> --bonus "pid1,4.25" --bonus "pid2,3.50"

  # Create from a CSV file (headerless, format: participant_id,amount)
  # Example bonuses.csv:
  #   5e15aae07bf572b8f97a847d,4.25
  #   6a22bbc18cf683c9g08b958e,3.50
  #   7b33ccd29dg794dah19c069f,2.00
  prolific bonus create <study_id> --file bonuses.csv -n

  # Scripted pipeline: create then pay
  prolific bonus create <study_id> --file bonuses.csv -n | head -1 | xargs prolific bonus pay -n`,
	}

	cmd.AddCommand(
		NewCreateCommand("create", client, w),
		NewPayCommand("pay", client, w),
	)

	return cmd
}
