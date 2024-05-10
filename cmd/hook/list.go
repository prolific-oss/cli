package hook

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the listing hooks command.
type ListOptions struct {
	Args        []string
	WorkspaceID string
	Enabled     bool
	Disabled    bool
	Limit       int
	Offset      int
}

// NewListCommand creates a new `hook list` command to give you details about
// your hooks.
func NewListCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide details about your hook subscriptions",
		Long: `List your hook subscriptions.

A hook subscription registers your interest to be notified of events happening
in the the Prolific Platform. Given a workspace ID, this will return a list of
subscriptions and explain which event types you are listening to.`,
		Example: `
This will use your default workspace
$ prolific hook list

This will use the specified workspace
$ prolific hook list -w 3461321e223a605c7a4f7612

You can couple this with options to only show disabled or enabled subscriptions
$ prolific hook list -w 3461321e223a605c7a4f7612 -d
$ prolific hook list -w 3461321e223a605c7a4f7612 -e
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := renderHooks(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", "", "Filter hooks by workspace.")
	flags.BoolVarP(&opts.Enabled, "enabled", "e", true, "Filter on enabled subscriptions.")
	flags.BoolVarP(&opts.Disabled, "disabled", "d", false, "Filter on disabled subscriptions.")
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of subscriptions returned")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of subscriptions to offset")

	return cmd
}

// renderHooks will show the users subscriptions.
func renderHooks(client client.API, opts ListOptions, w io.Writer) error {
	enabled := opts.Enabled

	if opts.Disabled {
		enabled = false
	}

	hooks, err := client.GetHooks(opts.WorkspaceID, enabled, opts.Limit, opts.Offset)
	if err != nil {
		return err
	}

	count := 0
	if hooks.JSONAPIMeta != nil {
		count = hooks.Meta.Count
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", "ID", "Event", "Target URL", "Enabled", "Workspace ID")
	for _, hook := range hooks.Results {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%v\t%s\n", hook.ID, hook.EventType, hook.TargetURL, hook.IsEnabled, hook.WorkspaceID)
	}

	_ = tw.Flush()

	fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(hooks.Results), count))

	return nil
}
