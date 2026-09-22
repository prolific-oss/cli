package filters

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// maxSearchQueryLength is the maximum length of a search query, in characters,
// accepted by the API after trimming.
const maxSearchQueryLength = 200

// searchPageSize is the maximum page size accepted by the API. Larger limits
// are satisfied by fetching several pages.
const searchPageSize = 100

// DefaultSearchLimit is the default number of results returned by filter
// search. It matches the API's default page size.
const DefaultSearchLimit = 25

// SearchOptions is the options for the filter search command.
type SearchOptions struct {
	Args        []string
	Query       string
	WorkspaceID string
	Limit       int
	All         bool
	JSON        bool
	NoPager     bool
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

By default the top 25 results are shown. Use --limit to ask for more, or --all
to fetch every match; pages are fetched from the API automatically.

When run in a terminal, output longer than one screen is shown in your pager
(PROLIFIC_PAGER, then PAGER, defaulting to less) so the top result stays in
view and you can scroll through the rest. Use --no-pager to print everything
directly.`,
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

Output as JSON for scripting or AI agents
$ prolific filters search developer --json`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Query = strings.TrimSpace(strings.Join(args, " "))

			if opts.All && cmd.Flags().Changed("limit") {
				return errors.New("error: --all and --limit cannot be used together")
			}

			if err := renderSearch(c, opts, w); err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "Scope the search to filters available in this workspace.")
	flags.IntVarP(&opts.Limit, "limit", "l", DefaultSearchLimit, "Maximum number of filters to return")
	flags.BoolVarP(&opts.All, "all", "a", false, "Return every matching filter")
	flags.BoolVarP(&opts.JSON, "json", "j", false, "Output as JSON")
	flags.BoolVar(&opts.NoPager, "no-pager", false, "Do not pipe output into a pager")

	return cmd
}

func renderSearch(c client.API, opts SearchOptions, w io.Writer) error {
	if opts.Query == "" {
		return errors.New("please provide a search query")
	}
	if len([]rune(opts.Query)) > maxSearchQueryLength {
		return fmt.Errorf("search query must be at most %d characters", maxSearchQueryLength)
	}
	want := opts.Limit
	if opts.All {
		want = 0
	} else if want < 1 {
		return errors.New("limit must be greater than or equal to 1")
	}

	fetch := func(limit, offset int) (client.Page[model.FilterSearchResult], error) {
		response, err := c.SearchFilters(opts.Query, opts.WorkspaceID, limit, offset)
		if err != nil {
			return client.Page[model.FilterSearchResult]{}, err
		}
		page := client.Page[model.FilterSearchResult]{Results: response.Results}
		if response.JSONAPIMeta != nil {
			page.Total = response.Meta.Count
		}
		return page, nil
	}

	records, total, err := client.FetchPages(want, searchPageSize, fetch)
	if err != nil {
		return err
	}
	if records == nil {
		records = []model.FilterSearchResult{}
	}

	if opts.JSON {
		return ui.JSONRenderer[model.FilterSearchResult]{}.Render(records, w)
	}

	if total == 0 {
		fmt.Fprintf(w, "No filters found matching %q\n", opts.Query)
		return nil
	}

	render := func(out io.Writer) error {
		if _, err := fmt.Fprint(out, RenderSearchHeader(opts.Query, len(records), total)); err != nil {
			return err
		}
		for i, record := range records {
			if i > 0 {
				if _, err := fmt.Fprint(out, RenderSearchRule()); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprint(out, RenderSearchResult(i+1, record)); err != nil {
				return err
			}
		}
		return nil
	}

	if opts.NoPager {
		return render(w)
	}
	return ui.Page(w, render)
}
