package user

import (
	"fmt"
	"io"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui"
	"github.com/spf13/cobra"
)

// NewMeCommand creates a new `user me` command to give you details about
// your account.
func NewMeCommand(client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "View details about your account",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := RenderMe(client, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
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

	content := ui.RenderHeading(fmt.Sprintf("%s %s", me.FirstName, me.LastName))
	content += fmt.Sprintln()
	content += fmt.Sprintf("ID:                %s\n", me.ID)
	content += fmt.Sprintf("Email:             %s\n", me.Email)
	// content += fmt.Sprintf("Available balance: %s\n", ui.RenderMoney((float64(me.AvailableBalance)/100), me.CurrencyCode))
	// content += fmt.Sprintf("Balance:           %s\n", ui.RenderMoney((float64(me.Balance)/100), me.CurrencyCode))

	fmt.Fprintln(w, content)

	return nil
}
