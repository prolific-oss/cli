// Package filters renders filter search results.
package filters

import (
	"fmt"
	"strings"

	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
)

// ruleWidth is the width of the divider drawn between results.
const ruleWidth = 60

// fieldWidth is the width of the field label column in a result.
const fieldWidth = 13

// indent is the indentation applied to every line of a result below its title.
const indent = "   "

// SearchListFields is the default column set for table and CSV output.
const SearchListFields = "Rank,FilterID,Title,Type,DataType,Category,MatchedOn"

// SearchListItem is a flattened search result for table and CSV output.
type SearchListItem struct {
	Rank      int
	FilterID  string
	Title     string
	Type      string
	DataType  string
	Category  string
	MatchedOn string
}

// NewSearchListItems flattens search results for table and CSV output,
// numbering them from firstRank in the order received.
func NewSearchListItems(results []model.FilterSearchResult, firstRank int) []SearchListItem {
	items := make([]SearchListItem, 0, len(results))
	for i, r := range results {
		items = append(items, SearchListItem{
			Rank:      firstRank + i,
			FilterID:  r.FilterID,
			Title:     r.Title,
			Type:      r.Type,
			DataType:  r.DataType,
			Category:  joinCategory(deref(r.Category), deref(r.Subcategory)),
			MatchedOn: strings.Join(r.Match.Fields, ", "),
		})
	}
	return items
}

// RenderSearchHeader renders the title block shown above search results, so
// the query and the API's match count are visible on the first screen without
// scrolling. It is written before results stream in, so it reports only what
// the API has said: the total, and whether the caller asked for fewer than
// that. The exact number rendered is reported by RenderSearchFooter.
func RenderSearchHeader(query string, total int, truncated bool) string {
	var b strings.Builder
	b.WriteString(ui.RenderHeading(fmt.Sprintf("Filters matching %q", query)))
	b.WriteString("\n")

	summary := fmt.Sprintf("%d matching %s", total, ui.Pluralise(total, "filter", "filters"))
	if truncated {
		summary += ". Use --limit or --all to see more"
	}
	b.WriteString(ui.RenderDimmed(summary))
	b.WriteString("\n\n")

	return b.String()
}

// RenderSearchFooter renders the closing line stating how many results were
// actually rendered out of the total.
func RenderSearchFooter(shown, total int) string {
	return "\n" + ui.RenderDimmed(ui.RenderRecordCounter(shown, total)) + "\n"
}

// RenderSearchRule renders the divider drawn between results.
func RenderSearchRule() string {
	return ui.RenderDimmed(strings.Repeat("─", ruleWidth)) + "\n"
}

// RenderNoSearchResults renders the message shown when nothing matched.
func RenderNoSearchResults(query string) string {
	return fmt.Sprintf("No filters found matching %q\n", query)
}

// RenderSearchResult renders a single filter search result at the given rank,
// highlighting the parts of the filter that matched the query and previewing
// matching choices.
func RenderSearchResult(rank int, r model.FilterSearchResult) string {
	var b strings.Builder
	h := newHighlights(r.Match.Highlights)

	// Title line: rank and highlighted title.
	b.WriteString(ui.RenderDimmed(fmt.Sprintf("%d.", rank)))
	b.WriteString(" ")
	b.WriteString(h.RenderStyled("title", r.Title, ui.RenderHeading))
	b.WriteString("\n")

	// Subtitle line: type, data type and category, dimmed.
	meta := []string{r.Type, r.DataType}
	if category := joinCategory(h.Render("category", deref(r.Category)), h.Render("subcategory", deref(r.Subcategory))); category != "" {
		meta = append(meta, category)
	}
	b.WriteString(indent)
	b.WriteString(ui.RenderDimmed(strings.Join(meta, " · ")))
	b.WriteString("\n")

	field := func(label, value string) {
		fmt.Fprintf(&b, "%s%-*s%s\n", indent, fieldWidth, label, value)
	}

	field("Filter ID", h.Render("filter_id", r.FilterID))
	if question := deref(r.Question); question != "" {
		field("Question", h.Render("question", question))
	}
	if r.Description != "" {
		field("Description", h.Render("description", r.Description))
	}
	if r.Type == "range" && (r.Min != nil || r.Max != nil) {
		field("Range", renderRange(r.Min, r.Max))
	}
	if r.NumChoices != nil {
		choices := fmt.Sprintf("%d total", *r.NumChoices)
		if r.MatchedChoices != nil {
			choices += fmt.Sprintf(", %d matching", r.MatchedChoices.Matched)
		}
		field("Choices", choices)
	}

	if r.MatchedChoices != nil && len(r.MatchedChoices.Results) > 0 {
		b.WriteString(renderMatchedChoices(*r.MatchedChoices))
	}

	if len(r.Match.Fields) > 0 {
		b.WriteString(indent)
		b.WriteString(ui.RenderDimmed(fmt.Sprintf("%-*s%s", fieldWidth, "Matched on", strings.Join(r.Match.Fields, ", "))))
		b.WriteString("\n")
	}

	return b.String()
}

// newHighlights indexes API highlights by field so each field is rendered
// with a single lookup.
func newHighlights(highlights []model.FilterSearchHighlight) ui.FieldHighlights {
	h := make(ui.FieldHighlights, len(highlights))
	for _, hl := range highlights {
		h[hl.Field] = append(h[hl.Field], ui.HighlightSpan{Start: hl.Start, End: hl.End})
	}
	return h
}

// renderMatchedChoices renders the preview of matching choices as a small
// indented table. It sits under the metadata block, with its ID column the
// same width as the field labels so the label column lines up with the field
// values above it.
func renderMatchedChoices(mc model.FilterMatchingChoices) string {
	var b strings.Builder
	choiceIndent := indent + strings.Repeat(" ", fieldWidth)

	b.WriteString("\n")
	b.WriteString(choiceIndent)
	b.WriteString(ui.RenderDimmed(fmt.Sprintf("%-*s%s", fieldWidth, "Choice ID", "Label")))
	b.WriteString("\n")

	for _, choice := range mc.Results {
		label := newHighlights(choice.Match.Highlights).Render("label", choice.Label)
		b.WriteString(choiceIndent)
		fmt.Fprintf(&b, "%-*s", fieldWidth, choice.ID)
		b.WriteString(label)
		if choice.NumDescendants > 0 {
			b.WriteString(ui.RenderDimmed(fmt.Sprintf("  (+%d nested)", choice.NumDescendants)))
		}
		b.WriteString("\n")
	}
	if mc.Truncated {
		remaining := mc.Matched - len(mc.Results)
		b.WriteString(choiceIndent)
		b.WriteString(ui.RenderDimmed(fmt.Sprintf("…and %d more matching %s", remaining, ui.Pluralise(remaining, "choice", "choices"))))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}

func renderRange(minValue, maxValue any) string {
	format := func(v any) string {
		if v == nil {
			return "-"
		}
		return fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf("%s to %s", format(minValue), format(maxValue))
}

func joinCategory(category, subcategory string) string {
	var parts []string
	if category != "" {
		parts = append(parts, category)
	}
	if subcategory != "" {
		parts = append(parts, subcategory)
	}
	return strings.Join(parts, " / ")
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
