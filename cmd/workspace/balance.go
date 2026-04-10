package workspace

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewBalanceCommand creates a new command to show the balance of a workspace.
func NewBalanceCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   commandName + " [workspace-id]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Show the balance of a workspace",
		Long: `Show the balance of a workspace

Displays the total and available balance for a workspace, broken down by
rewards, fees, and VAT. Amounts are shown in the workspace's currency.
`,
		Example: `
Show the balance of a workspace
$ prolific workspace balance <workspace-id>

Show the balance using the workspace set in your config file
$ prolific workspace balance

Set a default workspace in your config file:
workspace: <workspace-id>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceID := viper.GetString("workspace")
			if len(args) > 0 {
				workspaceID = args[0]
			}

			if workspaceID == "" {
				return errors.New("error: please provide a workspace ID")
			}

			err := renderWorkspaceBalance(c, workspaceID, w)
			if err != nil {
				return fmt.Errorf("error: %s", err)
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
	fmt.Fprintf(tw, "ID:\t%s\n", workspaceID)
	fmt.Fprintf(tw, "Currency:\t%s\n", balance.CurrencyCode)
	fmt.Fprintf(tw, "\n")
	fmt.Fprintf(tw, "Total Balance:\t\t\t%.2f\n", toCurrency(balance.TotalBalance))
	fmt.Fprintf(tw, "  Rewards:\t\t\t%.2f\n", toCurrency(balance.BalanceBreakdown.Rewards))
	fmt.Fprintf(tw, "  Fees:\t\t\t%.2f\n", toCurrency(balance.BalanceBreakdown.Fees))
	fmt.Fprintf(tw, "  VAT:\t\t\t%.2f\n", toCurrency(balance.BalanceBreakdown.VAT))
	fmt.Fprintf(tw, "\n")
	fmt.Fprintf(tw, "Available Balance:\t%.2f\n", toCurrency(balance.AvailableBalance))
	fmt.Fprintf(tw, "  Rewards:\t%.2f\n", toCurrency(balance.AvailableBalanceBreakdown.Rewards))
	fmt.Fprintf(tw, "  Fees:\t%.2f\n", toCurrency(balance.AvailableBalanceBreakdown.Fees))
	fmt.Fprintf(tw, "  VAT:\t%.2f\n", toCurrency(balance.AvailableBalanceBreakdown.VAT))
	_ = tw.Flush()

	return nil
}
