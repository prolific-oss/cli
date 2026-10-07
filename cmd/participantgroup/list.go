package participantgroup

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// defaultListFields is the default column selection for participant groups,
// shared by the list and search commands so both render the same table.
const defaultListFields = "ID,Name"

// ListOptions is the options for the listing participant groups command.
type ListOptions struct {
	Args        []string
	WorkspaceID string
	Limit       int
	Offset      int
	Fields      string
	Output      shared.OutputOptions
}

// NewListCommand creates a new command to deal with participant groups
func NewListCommand(commandName string, c client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Provide details about your participant groups",
		Long: `List your participant groups

Participant groups are assigned to a workspace.
`,
		Example: `
List the participant groups you have defined in a given workspace

$ prolific participant-group list -w 6261321e223a605c7a4f7623

You can output as a table, CSV or JSON
$ prolific participant-group list -w 6261321e223a605c7a4f7623 --table
$ prolific participant-group list -w 6261321e223a605c7a4f7623 --csv
$ prolific participant-group list -w 6261321e223a605c7a4f7623 --json
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
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "Filter participant groups by workspace.")
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Limit the number of participant groups returned")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "The number of participant groups to offset")
	flags.StringVarP(&opts.Fields, "fields", "f", defaultListFields, "Comma separated fields to display in table or CSV output.")
	shared.AddOutputFlags(cmd, &opts.Output)

	return cmd
}

// render will list your participant groups
func render(c client.API, opts ListOptions, w io.Writer) error {
	if opts.WorkspaceID == "" {
		return errors.New("please provide a workspace ID")
	}

	groups, err := c.GetParticipantGroups(opts.WorkspaceID, opts.Limit, opts.Offset)
	if err != nil {
		return err
	}

	page := client.PageOf(groups.Results, groups.JSONAPIMeta)

	return shared.RenderRecords(w, opts.Output, opts.Fields, page.Results, page.Total)
}
