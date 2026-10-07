package filters

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	uifilters "github.com/prolific-oss/cli/ui/filters"
	"github.com/spf13/cobra"
)

// maxSearchQueryLength is the maximum length of a search query, in characters,
// accepted by the API after trimming.
const maxSearchQueryLength = 200

// DefaultSearchLimit is the default number of results returned by filter
// search. It matches the API's default page size.
const DefaultSearchLimit = 25

// NewSearchCommand creates the `filters search` command.
func NewSearchCommand(c client.API, w io.Writer) *cobra.Command {
	var opts shared.SearchOptions

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search filters by keyword",
		Long: `Search the filter catalogue by keyword.

The search covers filter titles, questions, descriptions, help text,
categories, subcategories, filter IDs and choice labels. It supports stemming,
prefixes and accent-insensitive matching.

Results are returned in ranked order. The parts of each filter that matched
your query are highlighted, and up to three matching choices are previewed for
filters with a fixed set of choices.

By default the top 25 results are shown. Use --limit to ask for more, or
--all (equivalently --limit 0) to fetch every match. Pages are fetched from
the API automatically, so there is no --offset.

When run in a terminal, output longer than one screen is shown in your pager
(PROLIFIC_PAGER, then PAGER, defaulting to less) so the top result stays in
view and you can scroll through the rest. Use the global --no-pager flag to
print everything directly.`,
		Example: `
Search for filters matching a keyword
$ prolific filters search developer

Search with a multi-word query
$ prolific filters search "software developer"

Scope the search to filters available in a workspace
$ prolific filters search developer -w <workspace-id>

Show the top 50 results
$ prolific filters search developer --limit 50

Fetch every matching filter
$ prolific filters search developer --all

Print everything without a pager
$ prolific filters search developer --all --no-pager

Output as a table or CSV, optionally choosing the columns
$ prolific filters search developer --table
$ prolific filters search developer --csv --fields Rank,FilterID,Title

Output as JSON for scripting or AI agents
$ prolific filters search developer --json`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := renderSearch(cmd, c, opts, args, w); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	shared.AddSearchFlags(cmd, &opts, shared.SearchFlags{
		DefaultFields:  uifilters.SearchListFields,
		DefaultLimit:   DefaultSearchLimit,
		WorkspaceUsage: "Scope the search to filters available in this workspace.",
	})

	return cmd
}

func renderSearch(cmd *cobra.Command, c client.API, opts shared.SearchOptions, args []string, w io.Writer) error {
	query, err := shared.SearchQuery(args)
	if err != nil {
		return err
	}
	if len([]rune(query)) > maxSearchQueryLength {
		return fmt.Errorf("search query must be at most %d characters", maxSearchQueryLength)
	}

	want, err := opts.Want()
	if err != nil {
		return err
	}

	fetch := func(limit, offset int) (client.Page[model.FilterSearchResult], error) {
		response, err := c.SearchFilters(query, opts.WorkspaceID, limit, offset)
		if err != nil {
			return client.Page[model.FilterSearchResult]{}, err
		}
		return client.PageOf(response.Results, response.JSONAPIMeta), nil
	}

	switch shared.ResolveFormat(opts.Output) {
	case "json":
		records, _, err := client.FetchPages(want, client.FilterSearchPageSize, fetch)
		if err != nil {
			return err
		}
		return ui.JSONRenderer[model.FilterSearchResult]{}.Render(records, w)
	case "csv":
		records, _, err := client.FetchPages(want, client.FilterSearchPageSize, fetch)
		if err != nil {
			return err
		}
		renderer := ui.CsvRenderer[uifilters.SearchListItem]{}
		return renderer.Render(uifilters.NewSearchListItems(records, 1), opts.Fields, w)
	case "table":
		records, total, err := client.FetchPages(want, client.FilterSearchPageSize, fetch)
		if err != nil {
			return err
		}
		renderer := ui.TableRenderer[uifilters.SearchListItem]{}
		if err := renderer.Render(uifilters.NewSearchListItems(records, 1), opts.Fields, w); err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(records), total))
		return err
	}

	// Show progress on stderr while the first page loads. It is cleared as
	// soon as output begins, or before an error is returned, and is a no-op
	// when stderr is not a terminal.
	clearStatus := ui.Status(fmt.Sprintf("Searching filters for %q…", query))
	defer clearStatus()

	render := func(out io.Writer) error {
		return streamSearchResults(out, query, want, fetch, clearStatus)
	}
	if shared.NoPager(cmd) {
		return render(w)
	}
	return ui.Page(cmd.Context(), w, render)
}

// streamSearchResults writes formatted results to out page by page as they
// arrive from the API, so the first screen appears before every page has been
// fetched. beforeOutput is called once the first page has arrived, before
// anything is written, so any progress indicator can be cleared.
func streamSearchResults(out io.Writer, query string, want int, fetch client.PageFetcher[model.FilterSearchResult], beforeOutput func()) error {
	rank := 0
	total := 0

	err := client.EachPage(want, client.FilterSearchPageSize, fetch, func(page client.Page[model.FilterSearchResult]) error {
		if rank == 0 {
			beforeOutput()
			if page.Total == 0 && len(page.Results) == 0 {
				_, err := fmt.Fprint(out, uifilters.RenderNoSearchResults(query))
				return err
			}
			total = max(page.Total, len(page.Results))
			truncated := want > 0 && want < total
			if _, err := fmt.Fprint(out, uifilters.RenderSearchHeader(query, total, truncated)); err != nil {
				return err
			}
		}

		for _, record := range page.Results {
			rank++
			if rank > 1 {
				if _, err := fmt.Fprint(out, uifilters.RenderSearchRule()); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprint(out, uifilters.RenderSearchResult(rank, record)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil || rank == 0 {
		return err
	}

	// The footer counts what was actually rendered, which can be fewer than
	// the header's estimate if the API returned less than its own count.
	_, err = fmt.Fprint(out, uifilters.RenderSearchFooter(rank, max(total, rank)))
	return err
}
