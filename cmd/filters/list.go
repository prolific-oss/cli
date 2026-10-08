package filters

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	uifilters "github.com/prolific-oss/cli/ui/filters"
	"github.com/spf13/cobra"
)

// ListOptions is the options for the filter list command.
type ListOptions struct {
	WorkspaceID string
	Fields      string
	Output      shared.OutputOptions
}

// NewListCommand creates the `filters list` command.
func NewListCommand(c client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all filters available for your study",
		Long: `List every filter in the catalogue.

Use this when you want to browse the full set of filters. To find filters by
keyword, use ` + "`prolific filters search`" + ` instead.

Scoping to a workspace asks the API for that workspace's catalogue, which is
both smaller and cheaper to fetch than the unscoped one.

The catalogue endpoint does not paginate — it returns every filter in one
response — so there are no --limit or --offset flags here.

Run in a terminal without a format flag, the filters are shown in an
interactive, searchable interface. Piped into another program, they are
rendered as a table instead.`,
		Example: `
List all filters in an interactive, searchable interface
$ prolific filters list

Scope the catalogue to a workspace
$ prolific filters list -w <workspace-id>

Output as a table or CSV, optionally choosing the columns
$ prolific filters list --table
$ prolific filters list --csv -f FilterID,Title,Choices

Output as JSON for scripting or AI agents
$ prolific filters list --json

The fields you can use are
- FilterID
- Title
- Description
- Question
- Type
- DataType
- Min
- Max
- Choices`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := renderList(c, opts, w); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	shared.AddWorkspaceFlag(cmd, &opts.WorkspaceID)
	shared.AddFieldsFlag(cmd, &opts.Fields, uifilters.ListFields)
	shared.AddOutputFlags(cmd, &opts.Output)

	return cmd
}

func renderList(c client.API, opts ListOptions, w io.Writer) error {
	filters, err := c.GetFilters(opts.WorkspaceID)
	if err != nil {
		return err
	}
	records := filters.Results

	switch shared.ResolveFormatForWriter(opts.Output, w) {
	case shared.FormatJSON:
		// The catalogue arrives whole, so it is reported as a single window
		// covering every filter.
		envelope := ui.NewEnvelope(records, len(records), len(records), 0)
		return ui.JSONEnvelopeRenderer[model.Filter]{}.Render(envelope, w)
	case shared.FormatCSV:
		renderer := ui.CsvRenderer[uifilters.ListItem]{}
		return renderer.Render(uifilters.NewListItems(records), opts.Fields, w)
	case shared.FormatTable:
		renderer := ui.TableRenderer[uifilters.ListItem]{}
		if err := renderer.Render(uifilters.NewListItems(records), opts.Fields, w); err != nil {
			return err
		}
		_, err := fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(records), len(records)))
		return err
	default: // shared.FormatInteractive
		return renderInteractiveList(c, records)
	}
}

func renderInteractiveList(c client.API, filters []model.Filter) error {
	var items []list.Item

	for _, f := range filters {
		items = append(items, f)
	}

	lv := ListView{
		List:   list.New(items, list.NewDefaultDelegate(), 0, 0),
		Client: c,
	}
	lv.List.Title = "Filters"

	p := tea.NewProgram(lv)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("cannot render filters: %s", err)
	}

	return nil
}
