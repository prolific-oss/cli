package filters

import (
	"errors"
	"fmt"
	"io"
	"strings"

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

// SearchOptions is the options for the filter search command.
type SearchOptions struct {
	Query       string
	WorkspaceID string
	Limit       int
	Offset      int
	All         bool
	Output      shared.OutputOptions
	Fields      string
}

// NewSearchCommand creates the `filters search` command.
func NewSearchCommand(c client.API, w io.Writer) *cobra.Command {
	var opts SearchOptions

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
the API automatically, so --offset is only needed to skip past results you
have already seen.

When run in a terminal, output longer than one screen is shown in your pager
(PROLIFIC_PAGER, then PAGER, defaulting to less) so the top result stays in
view and you can scroll through the rest. Use the global --no-pager flag to
print everything directly. Piped into another program without a format flag,
results are rendered as a table instead.`,
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

Skip the first 50 matches
$ prolific filters search developer --offset 50

Print everything without a pager
$ prolific filters search developer --all --no-pager

Output as a table or CSV, optionally choosing the columns
$ prolific filters search developer --table
$ prolific filters search developer --csv -f Rank,FilterID,Title

Output as JSON for scripting or AI agents
$ prolific filters search developer --json`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Query = strings.TrimSpace(strings.Join(args, " "))

			if err := renderSearch(cmd, c, opts, w); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.IntVarP(&opts.Limit, "limit", "l", DefaultSearchLimit, "Maximum number of filters to return. Use 0 to fetch every match.")
	flags.IntVarP(&opts.Offset, "offset", "o", client.DefaultRecordOffset, "Number of matching filters to skip before returning results")
	flags.BoolVarP(&opts.All, "all", "a", false, "Return every matching filter (same as --limit 0)")
	shared.AddWorkspaceFlag(cmd, &opts.WorkspaceID)
	shared.AddFieldsFlag(cmd, &opts.Fields, uifilters.SearchListFields)
	shared.AddOutputFlags(cmd, &opts.Output)

	cmd.MarkFlagsMutuallyExclusive("all", "limit")

	return cmd
}

func renderSearch(cmd *cobra.Command, c client.API, opts SearchOptions, w io.Writer) error {
	if opts.Query == "" {
		return errors.New("please provide a search query")
	}
	if len([]rune(opts.Query)) > maxSearchQueryLength {
		return fmt.Errorf("search query must be at most %d characters", maxSearchQueryLength)
	}

	if opts.Limit < 0 {
		return errors.New("limit must be greater than or equal to 0")
	}
	if opts.Offset < 0 {
		return errors.New("offset must be greater than or equal to 0")
	}
	want := opts.Limit
	if opts.All {
		want = 0
	}

	// Paging starts from the caller's offset, so the offsets the fetcher is
	// given are relative to it.
	fetch := func(limit, offset int) (client.Page[model.FilterSearchResult], error) {
		response, err := c.SearchFilters(opts.Query, opts.WorkspaceID, limit, opts.Offset+offset)
		if err != nil {
			return client.Page[model.FilterSearchResult]{}, err
		}
		page := client.Page[model.FilterSearchResult]{Results: response.Results}
		if response.JSONAPIMeta != nil {
			page.Total = response.Meta.Count
		}
		return page, nil
	}

	switch shared.ResolveFormatForWriter(opts.Output, w) {
	case shared.FormatJSON:
		records, _, err := client.FetchPages(want, client.FilterSearchPageSize, fetch)
		if err != nil {
			return err
		}
		return ui.JSONRenderer[model.FilterSearchResult]{}.Render(records, w)
	case shared.FormatCSV:
		records, _, err := client.FetchPages(want, client.FilterSearchPageSize, fetch)
		if err != nil {
			return err
		}
		renderer := ui.CsvRenderer[uifilters.SearchListItem]{}
		return renderer.Render(uifilters.NewSearchListItems(records, opts.Offset+1), opts.Fields, w)
	case shared.FormatTable:
		records, total, err := client.FetchPages(want, client.FilterSearchPageSize, fetch)
		if err != nil {
			return err
		}
		renderer := ui.TableRenderer[uifilters.SearchListItem]{}
		if err := renderer.Render(uifilters.NewSearchListItems(records, opts.Offset+1), opts.Fields, w); err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(records), total))
		return err
	}

	// Show progress on stderr while the first page loads. It is cleared as
	// soon as output begins, or before an error is returned, and is a no-op
	// when stderr is not a terminal.
	clearStatus := ui.Status(fmt.Sprintf("Searching filters for %q…", opts.Query))
	defer clearStatus()

	render := func(out io.Writer) error {
		return streamSearchResults(out, opts.Query, want, opts.Offset+1, fetch, clearStatus)
	}
	if shared.NoPager(cmd) {
		return render(w)
	}
	return ui.Page(cmd.Context(), w, render)
}

// streamSearchResults writes formatted results to out page by page as they
// arrive from the API, so the first screen appears before every page has been
// fetched. Results are numbered from firstRank, so an offset search reports
// each filter's rank in the whole result set rather than in this slice of it.
// beforeOutput is called once the first page has arrived, before anything is
// written, so any progress indicator can be cleared.
func streamSearchResults(out io.Writer, query string, want, firstRank int, fetch client.PageFetcher[model.FilterSearchResult], beforeOutput func()) error {
	shown := 0
	total := 0

	err := client.EachPage(want, client.FilterSearchPageSize, fetch, func(page client.Page[model.FilterSearchResult]) error {
		if shown == 0 {
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
			if shown > 0 {
				if _, err := fmt.Fprint(out, uifilters.RenderSearchRule()); err != nil {
					return err
				}
			}
			shown++
			if _, err := fmt.Fprint(out, uifilters.RenderSearchResult(firstRank+shown-1, query, record)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil || shown == 0 {
		return err
	}

	// The footer counts what was actually rendered, which can be fewer than
	// the header's estimate if the API returned less than its own count.
	_, err = fmt.Fprint(out, uifilters.RenderSearchFooter(shown, max(total, shown)))
	return err
}
