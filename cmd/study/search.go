package study

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
)

// defaultSearchFields is the default column selection for search results. It
// extends defaultListFields with the internal name, which is the field most
// often searched on.
const defaultSearchFields = defaultListFields + ",InternalName"

// NewSearchCommand searches studies without opening the interactive list.
func NewSearchCommand(c client.API, w io.Writer) *cobra.Command {
	var opts shared.SearchOptions

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search accessible studies by name, internal name, or ID",
		Long: `Search studies on the server.

Results are scoped to your access, and optionally to a workspace. By default
the first 200 matches are returned. Use --limit to ask for more or fewer, or
--all (equivalently --limit 0) to fetch every match. Pages are fetched from the
API automatically, so there is no --page.`,
		Example: `
Search your studies
$ prolific study search 'memory task'

Scope the search to a workspace
$ prolific study search 'memory task' -w <workspace-id>

Fetch every matching study
$ prolific study search memory --all

Output as a table or CSV, optionally choosing the columns
$ prolific study search memory --table
$ prolific study search memory --csv --fields ID,Name

Output as JSON for scripting or AI agents
$ prolific study search memory --json`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := renderSearch(cmd, c, opts, args, w); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	shared.AddSearchFlags(cmd, &opts, shared.SearchFlags{
		DefaultFields:  defaultSearchFields,
		WorkspaceUsage: "Scope the search to a workspace.",
	})

	return cmd
}

func renderSearch(cmd *cobra.Command, c client.API, opts shared.SearchOptions, args []string, w io.Writer) error {
	query, err := shared.SearchQuery(args)
	if err != nil {
		return err
	}
	want, err := opts.Want()
	if err != nil {
		return err
	}

	// The studies endpoint paginates by page number rather than limit and
	// offset, which is why this walks pages instead of using FetchPages.
	records, total, err := client.FetchNumberedPages(want,
		func(page int) (client.Page[model.Study], error) {
			response, err := c.SearchStudies(query, opts.WorkspaceID, page)
			if err != nil {
				return client.Page[model.Study]{}, err
			}
			return client.PageOf(response.Results, response.JSONAPIMeta), nil
		})
	if err != nil {
		return err
	}

	return shared.RenderRecordsPaged(cmd, w, opts.Output, opts.Fields, records, total)
}
