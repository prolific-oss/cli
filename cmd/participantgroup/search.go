package participantgroup

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SearchOptions struct {
	WorkspaceID string
	Limit       int
	Offset      int
	JSON        bool
}

// NewSearchCommand searches participant group names on the server.
func NewSearchCommand(c client.API, w io.Writer) *cobra.Command {
	var opts SearchOptions
	cmd := &cobra.Command{
		Use:     "search <query>",
		Short:   "Search participant groups by name within a workspace",
		Long:    "Search participant group names on the server. JSON retains results and pagination metadata; use --limit and --offset to retrieve further matches.",
		Example: "prolific participant search 'pilot cohort' --workspace WORKSPACE_ID --json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			if query == "" {
				return fmt.Errorf("error: search query must not be empty")
			}
			if strings.TrimSpace(opts.WorkspaceID) == "" {
				return fmt.Errorf("error: please provide a workspace ID")
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
			if opts.JSON {
				return json.NewEncoder(w).Encode(response)
			}
			if err := (ui.TableRenderer[model.ParticipantGroup]{}).Render(response.Results, "ID,Name", w); err != nil {
				return fmt.Errorf("error: %s", err)
			}
			if response.JSONAPIMeta != nil {
				if _, err := fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(response.Results), response.Meta.Count)); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "Workspace to search (required).")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Maximum groups per page.")
	cmd.Flags().IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "Number of matching groups to skip.")
	cmd.Flags().BoolVarP(&opts.JSON, "json", "j", false, "Output results and pagination metadata as JSON.")
	return cmd
}
