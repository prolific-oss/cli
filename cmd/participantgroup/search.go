package participantgroup

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
)

// NewSearchCommand searches participant group names on the server.
func NewSearchCommand(c client.API, w io.Writer) *cobra.Command {
	var opts shared.SearchOptions

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search participant groups by name within a workspace",
		Long: `Search participant group names on the server.

By default the first 200 matches are returned. Use --limit to ask for more or
fewer, or --all (equivalently --limit 0) to fetch every match. Pages are
fetched from the API automatically, so there is no --offset.`,
		Example: `
Search the groups in a workspace
$ prolific participant-group search 'pilot cohort' -w <workspace-id>

Fetch every matching group
$ prolific participant-group search pilot -w <workspace-id> --all

Output as a table or CSV, optionally choosing the columns
$ prolific participant-group search pilot -w <workspace-id> --table
$ prolific participant-group search pilot -w <workspace-id> --csv --fields Name

Output as JSON for scripting or AI agents
$ prolific participant-group search pilot -w <workspace-id> --json`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := renderSearch(cmd, c, opts, args, w); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	shared.AddSearchFlags(cmd, &opts, shared.SearchFlags{
		DefaultFields:  defaultListFields,
		WorkspaceUsage: "Workspace to search (required).",
	})

	return cmd
}

func renderSearch(cmd *cobra.Command, c client.API, opts shared.SearchOptions, args []string, w io.Writer) error {
	query, err := shared.SearchQuery(args)
	if err != nil {
		return err
	}
	if strings.TrimSpace(opts.WorkspaceID) == "" {
		return errors.New("please provide a workspace ID")
	}
	want, err := opts.Want()
	if err != nil {
		return err
	}

	records, total, err := client.FetchPages(want, client.DefaultRecordLimit,
		func(limit, offset int) (client.Page[model.ParticipantGroup], error) {
			response, err := c.SearchParticipantGroups(query, opts.WorkspaceID, limit, offset)
			if err != nil {
				return client.Page[model.ParticipantGroup]{}, err
			}
			return client.PageOf(response.Results, response.JSONAPIMeta), nil
		})
	if err != nil {
		return err
	}

	return shared.RenderRecordsPaged(cmd, w, opts.Output, opts.Fields, records, total)
}
