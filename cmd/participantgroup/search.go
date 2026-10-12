package participantgroup

import (
	"fmt"
	"io"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

type SearchOptions struct {
	WorkspaceID string
	Limit       int
	Offset      int
	Fields      string
	Output      shared.OutputOptions
}

// searchFields is the default column set. A group is identified by its ID and
// name whichever format asked for it.
var searchFields = ui.FieldSet{CSV: "ID,Name", Table: "ID,Name"}

// NewSearchCommand searches participant group names on the server.
func NewSearchCommand(c client.API, w io.Writer) *cobra.Command {
	var opts SearchOptions
	cmd := &cobra.Command{
		Use:     "search <query>",
		Short:   "Search participant groups by name within a workspace",
		Long:    "Search participant group names on the server. JSON emits the CLI envelope: results, plus the count, limit and offset describing the window they came from. Use --limit and --offset to retrieve further matches.",
		Example: "prolific participant-group search 'pilot cohort' --workspace WORKSPACE_ID --json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			if query == "" {
				return fmt.Errorf("error: search query must not be empty")
			}
			if err := shared.RequireWorkspace(opts.WorkspaceID); err != nil {
				return fmt.Errorf("error: %s", err)
			}
			if opts.Limit < 1 || opts.Offset < 0 {
				return fmt.Errorf("error: limit must be positive and offset nonnegative")
			}
			response, err := c.SearchParticipantGroups(query, opts.WorkspaceID, opts.Limit, opts.Offset)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}
			if response.Results == nil {
				response.Results = []model.ParticipantGroup{}
			}
			// One total for every format, so the JSON count and the table
			// footer cannot disagree about the same response.
			total := client.ReportedTotal(response.JSONAPIMeta, len(response.Results))

			format := shared.ResolveFormat(opts.Output)
			fields := searchFields.Resolve(opts.Fields, format)

			switch format {
			case ui.FormatJSON:
				envelope := ui.NewEnvelope(response.Results, total, opts.Limit, opts.Offset)
				if err := (ui.JSONEnvelopeRenderer[model.ParticipantGroup]{}).Render(envelope, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
				return nil
			case ui.FormatCSV:
				if err := (ui.CsvRenderer[model.ParticipantGroup]{}).Render(response.Results, fields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
				return nil
			}
			if err := (ui.TableRenderer[model.ParticipantGroup]{}).Render(response.Results, fields, w); err != nil {
				return fmt.Errorf("error: %s", err)
			}
			if _, err := fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(response.Results), total)); err != nil {
				return fmt.Errorf("error: %s", err)
			}
			return nil
		},
	}
	shared.AddRequiredWorkspaceFlag(cmd, &opts.WorkspaceID)
	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Maximum groups per page.")
	cmd.Flags().IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "Number of matching groups to skip.")
	shared.AddFieldsFlag(cmd, &opts.Fields)
	shared.AddOutputFlags(cmd, &opts.Output)
	return cmd
}
