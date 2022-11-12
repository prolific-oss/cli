package workspace

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// NewListCommand creates a new command to deal with workspaces
func NewListCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide details about your workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := renderWorkspaces(client, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}

// renderWorkspaces will show your workspaces
func renderWorkspaces(client client.API, w io.Writer) error {
	workspaces, err := client.GetWorkspaces()
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\n", "ID", "Title", "Description")
	for _, workspace := range workspaces.Results {
		fmt.Fprintf(tw, "%s\t%s\t%v\n", workspace.ID, workspace.Title, workspace.Description)
	}

	tw.Flush()

	return nil
}
