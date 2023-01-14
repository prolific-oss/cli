package hook

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing hooks command.
type ListOptions struct {
	Args     []string
	Enabled  bool
	Disabled bool
}

// NewListCommand creates a new `hook list` command to give you details about
// your hooks.
func NewListCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide details about your hook subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := RenderHooks(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&opts.Enabled, "enabled", "e", true, "Filter on enabled subscriptions.")
	flags.BoolVarP(&opts.Disabled, "disabled", "d", false, "Filter on disabled subscriptions.")

	return cmd
}

// RenderHooks will show the users subscriptions.
func RenderHooks(client client.API, opts ListOptions, w io.Writer) error {
	enabled := opts.Enabled

	if opts.Disabled {
		enabled = false
	}

	hooks, err := client.GetHooks(enabled)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t\n", "ID", "Event", "Target URL", "Enabled", "Workspace ID")
	for _, hook := range hooks.Results {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%v\t%s\t\n", hook.ID, hook.EventType, hook.TargetURL, hook.IsEnabled, hook.WorkspaceID)
	}

	tw.Flush()

	return nil
}
