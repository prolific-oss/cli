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

// ChoicesOptions is the options shared by the filter choices commands. Query
// is empty when listing rather than searching.
type ChoicesOptions struct {
	FilterID    string
	Query       string
	WorkspaceID string
	Fields      string
	Output      shared.OutputOptions
	Pagination  shared.PaginationOptions
}

// validate rejects the options neither choices endpoint can act on.
func (o ChoicesOptions) validate() error {
	if strings.TrimSpace(o.FilterID) == "" {
		return errors.New("please provide a filter ID")
	}
	return o.Pagination.Validate()
}

// validateSearch adds the query rule, which only this command wants: a blank
// query means "list everything" to the listing command.
func (o ChoicesOptions) validateSearch() error {
	if err := validateSearchQuery(o.Query); err != nil {
		return err
	}
	return o.validate()
}

// NewChoicesCommand creates the `filters choices` command, which lists a
// filter's choices and parents the `search` sub-command.
func NewChoicesCommand(c client.API, w io.Writer) *cobra.Command {
	var opts ChoicesOptions

	cmd := &cobra.Command{
		Use:   "choices <filter-id>",
		Short: "List the choices belonging to a filter",
		Long: `List the choices belonging to a select filter.

Choice IDs are what a filter selection is built from, so this is how you get
the IDs to select. Use ` + "`prolific filters search`" + ` to find a filter's ID first.

The output is flat, in the order the API returns it. Choices can describe a
hierarchy through their parent and child counts, but the endpoint offers no
way to fetch one node's children, so no tree is reconstructed and the IDs are
left raw.

The endpoint does not support ordering, so there are no sort flags.`,
		Example: `
List a filter's choices
$ prolific filters choices job-title

Fetch every choice, not just the first 200
$ prolific filters choices job-title --all

Page through them
$ prolific filters choices job-title --limit 50 --offset 100

Scope the lookup to a workspace
$ prolific filters choices job-title -w <workspace-id>

Output as CSV or JSON, optionally choosing the columns
$ prolific filters choices job-title --csv -f ID,Label
$ prolific filters choices job-title --json

Search within a filter's choices
$ prolific filters choices search job-title nurse

The fields you can use are
- ID
- Label
- ParentID
- NumChildren
- NumDescendants`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.FilterID = args[0]

			if err := opts.validate(); err != nil {
				return fmt.Errorf("error: %s", err)
			}
			if err := renderChoices(cmd, opts, listChoices(c, opts), w); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	addChoicesFlags(cmd, &opts)
	cmd.AddCommand(NewChoicesSearchCommand(c, w))

	return cmd
}

// NewChoicesSearchCommand creates the `filters choices search` command.
func NewChoicesSearchCommand(c client.API, w io.Writer) *cobra.Command {
	var opts ChoicesOptions

	cmd := &cobra.Command{
		Use:   "search <filter-id> <query>",
		Short: "Search the choices belonging to a filter",
		Long: `Search the choices belonging to a select filter.

Results come back in relevance order, with the parts of each label that
matched your query highlighted in a terminal. For a filter like job title,
which has thousands of choices, this is the quickest way to the handful of
choice IDs you actually want.

The endpoint does not support ordering — results are always ranked by
relevance — so there are no sort flags.`,
		Example: `
Search a filter's choices
$ prolific filters choices search job-title nurse

Search with a multi-word query
$ prolific filters choices search job-title "registered nurse"

Fetch every match
$ prolific filters choices search job-title nurse --all

Scope the search to a workspace
$ prolific filters choices search job-title nurse -w <workspace-id>

Output as CSV or JSON, optionally choosing the columns
$ prolific filters choices search job-title nurse --csv -f ID,Label
$ prolific filters choices search job-title nurse --json`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.FilterID = args[0]
			opts.Query = strings.TrimSpace(strings.Join(args[1:], " "))

			if err := opts.validateSearch(); err != nil {
				return fmt.Errorf("error: %s", err)
			}
			if err := renderChoices(cmd, opts, searchChoices(c, opts), w); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	addChoicesFlags(cmd, &opts)

	return cmd
}

// addChoicesFlags registers the flags both choices commands share.
func addChoicesFlags(cmd *cobra.Command, opts *ChoicesOptions) {
	shared.AddWorkspaceFlag(cmd, &opts.WorkspaceID)
	shared.AddFieldsFlag(cmd, &opts.Fields)
	shared.AddPaginationFlags(cmd, &opts.Pagination, client.DefaultRecordLimit)
	shared.AddOutputFlags(cmd, &opts.Output)
}

// listChoices fetches pages from the plain listing endpoint.
func listChoices(c client.API, opts ChoicesOptions) client.PageFetcher[model.FilterChoiceSearchResult] {
	return func(limit, offset int) (client.Page[model.FilterChoiceSearchResult], error) {
		response, err := c.GetFilterChoices(opts.FilterID, opts.WorkspaceID, limit, offset)
		if err != nil {
			return client.Page[model.FilterChoiceSearchResult]{}, err
		}
		return choicesPage(response.Results, response.JSONAPIMeta), nil
	}
}

// searchChoices fetches pages from the relevance-ranked search endpoint.
func searchChoices(c client.API, opts ChoicesOptions) client.PageFetcher[model.FilterChoiceSearchResult] {
	return func(limit, offset int) (client.Page[model.FilterChoiceSearchResult], error) {
		response, err := c.SearchFilterChoices(opts.FilterID, opts.Query, opts.WorkspaceID, limit, offset)
		if err != nil {
			return client.Page[model.FilterChoiceSearchResult]{}, err
		}
		return choicesPage(response.Results, response.JSONAPIMeta), nil
	}
}

func choicesPage(results []model.FilterChoiceSearchResult, meta *client.JSONAPIMeta) client.Page[model.FilterChoiceSearchResult] {
	page := client.Page[model.FilterChoiceSearchResult]{Results: results}
	if meta != nil {
		page.Total = meta.Meta.Count
	}
	return page
}

// renderChoices renders whatever fetchPage returns, so both commands share a
// presentation without sharing the choice of endpoint.
func renderChoices(cmd *cobra.Command, opts ChoicesOptions, fetchPage client.PageFetcher[model.FilterChoiceSearchResult], w io.Writer) error {
	want := opts.Pagination.Want()

	records := func() ([]model.FilterChoiceSearchResult, int, error) {
		return client.FetchPages(want, client.FilterChoicesPageSize, opts.Pagination.Offset, fetchPage)
	}

	format := shared.ResolveFormatForWriter(opts.Output, w)
	fields := uifilters.ChoiceFields.Resolve(opts.Fields, format)

	switch format {
	case ui.FormatJSON:
		found, total, err := records()
		if err != nil {
			return err
		}
		// want, not the --limit default: --all resolves to 0, meaning unbounded.
		envelope := ui.NewEnvelope(found, total, want, opts.Pagination.Offset)
		return ui.JSONEnvelopeRenderer[model.FilterChoiceSearchResult]{}.Render(envelope, w)
	case ui.FormatCSV:
		found, _, err := records()
		if err != nil {
			return err
		}
		renderer := ui.CsvRenderer[uifilters.ChoiceListItem]{}
		return renderer.Render(uifilters.NewChoiceListItems(found), fields, w)
	case ui.FormatTable:
		found, total, err := records()
		if err != nil {
			return err
		}
		renderer := ui.TableRenderer[uifilters.ChoiceListItem]{}
		if err := renderer.Render(uifilters.NewChoiceListItems(found), fields, w); err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(found), total))
		return err
	}

	render := func(out io.Writer) error {
		return streamChoices(out, opts, want, fetchPage)
	}
	if shared.NoPager(cmd) {
		return render(w)
	}
	return ui.Page(cmd.Context(), w, render)
}

// streamChoices writes choices page by page as they arrive, so the first
// screen appears before every page is fetched.
func streamChoices(out io.Writer, opts ChoicesOptions, want int, fetch client.PageFetcher[model.FilterChoiceSearchResult]) error {
	shown := 0
	total := 0

	err := client.EachPage(want, client.FilterChoicesPageSize, opts.Pagination.Offset, fetch, func(page client.Page[model.FilterChoiceSearchResult]) error {
		if shown == 0 {
			if page.Total == 0 && len(page.Results) == 0 {
				_, err := fmt.Fprint(out, uifilters.RenderNoChoices(opts.FilterID, opts.Query))
				return err
			}
			total = max(page.Total, len(page.Results))
			truncated := want > 0 && want < total
			if _, err := fmt.Fprint(out, uifilters.RenderChoicesHeader(opts.FilterID, opts.Query, total, truncated)); err != nil {
				return err
			}
			if _, err := fmt.Fprint(out, uifilters.RenderChoicesTableHeader()); err != nil {
				return err
			}
		}

		for _, record := range page.Results {
			shown++
			if _, err := fmt.Fprint(out, uifilters.RenderChoiceRow(record)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil || shown == 0 {
		return err
	}

	_, err = fmt.Fprint(out, uifilters.RenderResultsFooter(shown, max(total, shown)))
	return err
}
