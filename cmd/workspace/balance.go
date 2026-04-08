package workspace

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// NewBalanceCommand creates a new command to show the balance of a workspace.
func NewBalanceCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   commandName + " <workspace-id>",
		Args:  cobra.ExactArgs(1),
		Short: "Show the balance of a workspace",
		Long: `Show the balance of a workspace

Displays the total and available balance for a workspace, broken down by
rewards, fees, and VAT. Amounts are shown in the workspace's currency.
`,
		Example: `
Show the balance of a workspace
$ prolific workspace balance <workspace-id>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := renderWorkspaceBalance(c, args[0], w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}

func renderWorkspaceBalance(c client.API, workspaceID string, w io.Writer) error {
	balance, err := c.GetWorkspaceBalance(workspaceID)
	if err != nil {
		return err
	}

	toCurrency := func(amount int) float64 {
		return float64(amount) / 100
	}

	tw := tabwriter.NewWriter(w, 0, 1, 2, ' ', 0)
	fmt.Fprintf(tw, "Currency:\t%s\n", balance.CurrencyCode)
	fmt.Fprintf(tw, "\n")
	fmt.Fprintf(tw, "Total Balance:\t%.2f\n", toCurrency(balance.TotalBalance))
	fmt.Fprintf(tw, "  Rewards:\t%.2f\n", toCurrency(balance.BalanceBreakdown.Rewards))
	fmt.Fprintf(tw, "  Fees:\t%.2f\n", toCurrency(balance.BalanceBreakdown.Fees))
	fmt.Fprintf(tw, "  VAT:\t%.2f\n", toCurrency(balance.BalanceBreakdown.VAT))
	fmt.Fprintf(tw, "\n")
	fmt.Fprintf(tw, "Available Balance:\t%.2f\n", toCurrency(balance.AvailableBalance))
	fmt.Fprintf(tw, "  Rewards:\t%.2f\n", toCurrency(balance.AvailableBalanceBreakdown.Rewards))
	fmt.Fprintf(tw, "  Fees:\t%.2f\n", toCurrency(balance.AvailableBalanceBreakdown.Fees))
	fmt.Fprintf(tw, "  VAT:\t%.2f\n", toCurrency(balance.AvailableBalanceBreakdown.VAT))
	_ = tw.Flush()

	return nil
}
