package filtersets

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ListOptions is the options for the listing filter sets command.
type ListOptions struct {
	Args        []string
	WorkspaceID string
	Limit       int
	Offset      int
}

// NewListCommand creates a new command to deal with filter sets
func NewListCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide a list of your filter sets",
		Long: `List your Filter Sets

Filter Sets are assigned to a workspace.
`,
		Example: `
List the Filter Sets you have defined in a given workspace

$ prolific filters list -w 6261321e223a605c7a4f7623
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := render(c, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "Filter filter sets by workspace.")
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of filter sets returned")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of filter sets to offset")

	return cmd
}

// render will list your filter sets
func render(client client.API, opts ListOptions, w io.Writer) error {
	if opts.WorkspaceID == "" {
		return errors.New("please provide a workspace ID")
	}

	records, err := client.GetFilterSets(opts.WorkspaceID, opts.Limit, opts.Offset)
	if err != nil {
		return err
	}

	count := 0
	if records.JSONAPIMeta != nil {
		count = records.Meta.Count
	}

	tw := tabwriter.NewWriter(w, 0, 1, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\n", "ID", "Name")
	for _, record := range records.Results {
		fmt.Fprintf(tw, "%s\t%s\n", record.ID, record.Name)
	}

	_ = tw.Flush()

	fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(records.Results), count))

	return nil
}
