package user

import (
	"fmt"
	"io"
	"os"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// NewMeCommand creates a new `user me` command to give you details about
// your account.
func NewMeCommand(client client.API) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Provide details about your account",
		Run: func(cmd *cobra.Command, args []string) {

			err := RenderMe(client, os.Stdout)
			if err != nil {
				fmt.Print(err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

// RenderMe will provide information about the user account.
func RenderMe(client client.API, w io.Writer) error {
	me, err := client.GetMe()
	if err != nil {
		return err
	}

	var docStyle = lipgloss.NewStyle().Margin(1, 2)

	content := lipgloss.NewStyle().
		// Bold(true).
		// Underline(true).
		Background(lipgloss.Color(ui.Green)).
		MarginBottom(1).
		Padding(1).
		Align(lipgloss.Center).
		Render(fmt.Sprintf("%s %s", me.FirstName, me.LastName))

	content += fmt.Sprintln()
	content += fmt.Sprintf("Email:             %s\n", me.Email)
	content += fmt.Sprintf("Currency:          %s\n", me.CurrencyCode)
	content += fmt.Sprintf("Available balance: %.2f\n", float64(me.AvailableBalance)/100)
	content += fmt.Sprintf("Balance:           %.2f\n", float64(me.Balance)/100)

	fmt.Fprintln(w, docStyle.Render(content))

	return nil
}
