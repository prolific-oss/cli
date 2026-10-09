package filters

import (
	"fmt"
	"strings"

	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
)

// ChoiceFields is the default column set for each format. Listing and
// searching return the same records, so they render the same columns.
var ChoiceFields = ui.FieldSet{
	CSV:   "ID,Label,ParentID,NumChildren,NumDescendants",
	Table: "ID,Label",
}

// ChoiceListItem is a flattened choice for table and CSV output. IDs are left
// raw, including the parent's: a filter selection is built from IDs.
type ChoiceListItem struct {
	ID             string
	Label          string
	ParentID       string
	NumChildren    int
	NumDescendants int
}

// NewChoiceListItems flattens choices for table and CSV output, in the order
// the API returned them.
func NewChoiceListItems(choices []model.FilterChoiceSearchResult) []ChoiceListItem {
	items := make([]ChoiceListItem, 0, len(choices))
	for _, c := range choices {
		items = append(items, ChoiceListItem{
			ID:             c.ID,
			Label:          c.Label,
			ParentID:       deref(c.ParentID),
			NumChildren:    c.NumChildren,
			NumDescendants: c.NumDescendants,
		})
	}
	return items
}

// RenderChoicesHeader renders the title block shown above a filter's choices,
// so the filter, the query and the total stay visible on the first screen.
func RenderChoicesHeader(filterID, query string, total int, truncated bool) string {
	var b strings.Builder

	heading := fmt.Sprintf("Choices in %q", filterID)
	summary := fmt.Sprintf("%d %s", total, ui.Pluralise(total, "choice", "choices"))
	if query != "" {
		heading += fmt.Sprintf(" matching %q", query)
		summary = fmt.Sprintf("%d matching %s", total, ui.Pluralise(total, "choice", "choices"))
	}
	if truncated {
		summary += ". Use --limit or --all to see more"
	}

	b.WriteString(ui.RenderHeading(heading))
	b.WriteString("\n")
	b.WriteString(ui.RenderDimmed(summary))
	b.WriteString("\n\n")

	return b.String()
}

// RenderNoChoices renders the message shown when a filter has no choices, or
// none matching the query.
func RenderNoChoices(filterID, query string) string {
	if query != "" {
		return fmt.Sprintf("No choices found in %q matching %q\n", filterID, query)
	}
	return fmt.Sprintf("No choices found in %q\n", filterID)
}

// RenderChoicesTableHeader renders the column headings above the rows.
func RenderChoicesTableHeader() string {
	return ui.RenderDimmed(fmt.Sprintf("%-*s%s", fieldWidth, "Choice ID", "Label")) + "\n"
}

// RenderChoiceRow renders one choice: its raw ID, its label with any query
// matches highlighted, and where it sits in the hierarchy.
func RenderChoiceRow(c model.FilterChoiceSearchResult) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%-*s", fieldWidth, c.ID)
	b.WriteString(highlightChoiceLabel(c))
	if place := describeChoicePlace(c); place != "" {
		b.WriteString(ui.RenderDimmed("  (" + place + ")"))
	}
	b.WriteString("\n")

	return b.String()
}

// describeChoicePlace states where a choice sits in the hierarchy: its parent,
// how many children hang off it, and how big the subtree below it is. The
// descendant count is left out when it only repeats the child count, which is
// the case for every choice whose children are leaves.
func describeChoicePlace(c model.FilterChoiceSearchResult) string {
	var parts []string

	if parentID := deref(c.ParentID); parentID != "" {
		parts = append(parts, "parent "+parentID)
	}
	if c.NumChildren > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", c.NumChildren, ui.Pluralise(c.NumChildren, "child", "children")))
	}
	if c.NumDescendants > c.NumChildren {
		parts = append(parts, fmt.Sprintf("+%d nested", c.NumDescendants))
	}

	return strings.Join(parts, ", ")
}

// highlightChoiceLabel renders a choice's label with its query matches
// highlighted. The catalogue search's preview and the choices commands share
// it, so a matched choice is highlighted the same way wherever it appears.
func highlightChoiceLabel(c model.FilterChoiceSearchResult) string {
	return newHighlights(c.Matches).Render("label", c.Label)
}

// renderChoiceLabel renders a highlighted label followed by the size of the
// subtree beneath it, the shorter form the catalogue search preview uses.
func renderChoiceLabel(c model.FilterChoiceSearchResult) string {
	label := highlightChoiceLabel(c)
	if c.NumDescendants > 0 {
		label += ui.RenderDimmed(fmt.Sprintf("  (+%d nested)", c.NumDescendants))
	}
	return label
}
