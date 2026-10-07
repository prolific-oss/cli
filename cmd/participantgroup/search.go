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
	"github.com/spf13/viper"
)

// defaultSearchFields is the default column set for table and CSV output.
const defaultSearchFields = "ID,Name"

// DefaultSearchLimit is the default number of groups returned by search.
const DefaultSearchLimit = client.DefaultRecordLimit

type SearchOptions struct {
	WorkspaceID string
	Limit       int
	All         bool
	Fields      string
	Output      shared.OutputOptions
}

// NewSearchCommand searches participant group names on the server.
func NewSearchCommand(c client.API, w io.Writer) *cobra.Command {
	var opts SearchOptions
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search participant groups by name within a workspace",
		Long: `Search participant group names on the server.

Further pages are fetched automatically up to --limit; use --all to retrieve
every match.`,
		Example: "prolific participant-group search pilot cohort --workspace WORKSPACE_ID --json",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(strings.Join(args, " "))
			if query == "" {
				return fmt.Errorf("error: search query must not be empty")
			}
			if strings.TrimSpace(opts.WorkspaceID) == "" {
				return fmt.Errorf("error: please provide a workspace ID")
			}
			if opts.Limit < 0 {
				return fmt.Errorf("error: limit must be greater than or equal to 0")
			}

			want := opts.Limit
			if opts.All {
				want = 0
			}

			fetch := func(limit, offset int) (client.Page[model.ParticipantGroup], error) {
				response, err := c.SearchParticipantGroups(query, opts.WorkspaceID, limit, offset)
				if err != nil {
					return client.Page[model.ParticipantGroup]{}, err
				}
				result := client.Page[model.ParticipantGroup]{Results: response.Results}
				if response.JSONAPIMeta != nil {
					result.Total = response.Meta.Count
				}
				return result, nil
			}

			records, total, err := client.FetchPages(want, client.DefaultRecordLimit, fetch)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}
			if records == nil {
				records = []model.ParticipantGroup{}
			}

			switch shared.ResolveFormat(opts.Output) {
			case "json":
				if err := (ui.JSONRenderer[model.ParticipantGroup]{}).Render(records, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
				return nil
			case "csv":
				if err := (ui.CsvRenderer[model.ParticipantGroup]{}).Render(records, opts.Fields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
				return nil
			}

			if err := (ui.TableRenderer[model.ParticipantGroup]{}).Render(records, opts.Fields, w); err != nil {
				return fmt.Errorf("error: %s", err)
			}
			if _, err := fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(records), total)); err != nil {
				return fmt.Errorf("error: %s", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "Workspace to search (required).")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", DefaultSearchLimit, "Maximum number of groups to return. Use 0 to fetch every match.")
	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "Return every matching group (same as --limit 0)")
	cmd.Flags().StringVarP(&opts.Fields, "fields", "f", defaultSearchFields, "Comma separated fields to display in table or CSV output.")
	shared.AddOutputFlags(cmd, &opts.Output)
	cmd.MarkFlagsMutuallyExclusive("all", "limit")
	return cmd
}
