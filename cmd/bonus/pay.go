package bonus

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewPayCommand creates a new command to pay bonus payments
func NewPayCommand(commandName string, apiClient client.API, w io.Writer) *cobra.Command {
	var nonInteractive bool

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Pay previously created bonus payments",
		Long: `Trigger asynchronous payment of previously created bonus records.

The bonus payment ID is obtained from the output of 'bonus create'.
Payment is processed asynchronously — your account balance will be
updated within minutes.`,
		Example: `  # Pay with confirmation prompt
  prolific bonus pay <bonus_payment_id>

  # Pay without confirmation (for scripting)
  prolific bonus pay <bonus_payment_id> -n`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bonusID := args[0]
			reader := cmd.InOrStdin()

			err := payBonusPayments(apiClient, bonusID, nonInteractive, reader, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&nonInteractive, "non-interactive", "n", false, "Skip confirmation prompt")

	return cmd
}

// payBonusPayments orchestrates the pay bonus workflow.
func payBonusPayments(apiClient client.API, bonusID string, nonInteractive bool, reader io.Reader, w io.Writer) error {
	confirmed, err := confirmPayment(bonusID, nonInteractive, reader, w)
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Fprintln(w, "Payment cancelled. No bonuses were paid.")
		return nil
	}

	err = apiClient.PayBonusPayments(bonusID)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "Bonus payment request accepted. Bonuses will be paid asynchronously.")

	return nil
}
