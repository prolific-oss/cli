package user

import (
	"fmt"
	"io"
	"os"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// NewMeCommand creates a new `user me` command to give you details about
// your account.
func NewMeCommand(client client.API) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Provide details about your account",
		Run: func(cmd *cobra.Command, args []string) {

			err := renderMe(client, os.Stdout)
			if err != nil {
				fmt.Print(err)
				os.Exit(1)
			}
		},
	}

	return cmd
}

func renderMe(client client.API, w io.Writer) error {
	me, err := client.GetMe()
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "First name:           %s\n", me.FirstName)
	fmt.Fprintf(w, "Last name:            %s\n", me.LastName)
	fmt.Fprintf(w, "Email:                %s\n", me.Email)
	fmt.Fprintf(w, "Currency:             %s\n", me.CurrencyCode)
	fmt.Fprintf(w, "Available balance:    %d\n", me.AvailableBalance)
	fmt.Fprintf(w, "Balance:              %d\n", me.Balance)

	return nil
}
