// Package filters renders filter search results.
package filters

import (
	"fmt"
	"strconv"
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

// SearchListFields is the default column set for each format.
var SearchListFields = ui.FieldSet{
	CSV:   "FilterID,Title,Question,Category,Subcategory,Type,ChoicesTotal,ChoicesMatched,ChoicesTruncated",
	Table: "FilterID,Title,Type",
}

// SearchListItem is a flattened search result for table and CSV output. Every
// column is one field of the result: the category and subcategory stay apart,
// and the choice counts come straight from the choices block, which a filter
// without enumerable choices does not carry at all.
type SearchListItem struct {
	FilterID         string
	Title            string
	Question         string
	Description      string
	Category         string
	Subcategory      string
	Type             string
	DataType         string
	ChoicesTotal     int
	ChoicesMatched   int
	ChoicesTruncated bool
}

// NewSearchListItems flattens search results for table and CSV output, in the
// order received.
func NewSearchListItems(results []model.FilterSearchResult) []SearchListItem {
	items := make([]SearchListItem, 0, len(results))
	for _, r := range results {
		item := SearchListItem{
			FilterID:    r.FilterID,
			Title:       r.Title,
			Question:    deref(r.Question),
			Description: r.Description,
			Category:    deref(r.Category),
			Subcategory: deref(r.Subcategory),
			Type:        r.Type,
			DataType:    r.DataType,
		}
		if r.Choices != nil {
			item.ChoicesTotal = r.Choices.Total
			item.ChoicesMatched = r.Choices.Matched
			item.ChoicesTruncated = r.Choices.Truncated
		}
		items = append(items, item)
	}
	return items
}

// RenderSearchHeader renders the title block shown above search results, so
// the query and the API's match count are visible on the first screen without
// scrolling. It is written before results stream in, so it reports only what
// the API has said: the total, and whether the caller asked for fewer than
// that. The exact number rendered is reported by RenderResultsFooter.
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

// RenderResultsFooter renders the closing line stating how many records were
// actually rendered out of the total.
func RenderResultsFooter(shown, total int) string {
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
// matching choices. The query is repeated back in the command suggested for
// filters whose choices the preview does not cover.
func RenderSearchResult(rank int, query string, r model.FilterSearchResult) string {
	var b strings.Builder
	h := newHighlights(r.Matches)

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
	if r.Choices != nil {
		field("Choices", fmt.Sprintf("%d total, %d matching", r.Choices.Total, r.Choices.Matched))
		b.WriteString(renderMatchedChoices(r.FilterID, query, *r.Choices))
	}

	if fields := matchedFields(r); len(fields) > 0 {
		b.WriteString(indent)
		b.WriteString(ui.RenderDimmed(fmt.Sprintf("%-*s%s", fieldWidth, "Matched on", strings.Join(fields, ", "))))
		b.WriteString("\n")
	}

	return b.String()
}

// matchedFields summarises where a filter matched: the distinct fields of its
// own text matches in first-seen order, plus "choices" when any of its
// choices matched. The API no longer provides this summary directly.
func matchedFields(r model.FilterSearchResult) []string {
	var fields []string
	seen := make(map[string]bool)
	for _, m := range r.Matches {
		if !seen[m.Field] {
			seen[m.Field] = true
			fields = append(fields, m.Field)
		}
	}
	if r.Choices != nil && r.Choices.Matched > 0 {
		fields = append(fields, "choices")
	}
	return fields
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
// values above it. When the preview does not cover every choice it closes
// with the command that lists the rest, which is otherwise the end of the
// road for a filter with thousands of choices.
func renderMatchedChoices(filterID, query string, mc model.FilterSearchChoices) string {
	hint := choicesCommand(filterID, query, mc)
	if len(mc.Results) == 0 && hint == "" {
		return ""
	}

	var b strings.Builder
	choiceIndent := indent + strings.Repeat(" ", fieldWidth)

	b.WriteString("\n")
	if len(mc.Results) > 0 {
		b.WriteString(choiceIndent)
		b.WriteString(ui.RenderDimmed(fmt.Sprintf("%-*s%s", fieldWidth, "Choice ID", "Label")))
		b.WriteString("\n")
	}

	for _, choice := range mc.Results {
		b.WriteString(choiceIndent)
		fmt.Fprintf(&b, "%-*s", fieldWidth, choice.ID)
		b.WriteString(renderChoiceLabel(choice))
		b.WriteString("\n")
	}
	if mc.Truncated {
		remaining := mc.Matched - len(mc.Results)
		b.WriteString(choiceIndent)
		b.WriteString(ui.RenderDimmed(fmt.Sprintf("…and %d more matching %s", remaining, ui.Pluralise(remaining, "choice", "choices"))))
		b.WriteString("\n")
	}
	if hint != "" {
		b.WriteString(choiceIndent)
		b.WriteString(ui.RenderDimmed(hint))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}

// choicesCommand names the command that lists the choices this preview does
// not show: the matching ones when the preview was truncated, and the whole
// set when the filter has choices beyond those previewed. It returns an empty
// string when the preview already covers every choice.
func choicesCommand(filterID, query string, mc model.FilterSearchChoices) string {
	switch {
	case mc.Matched > len(mc.Results):
		return fmt.Sprintf("See them all: prolific filters choices search %s %s", filterID, quoteQuery(query))
	case mc.Total > len(mc.Results):
		return fmt.Sprintf("See all %d choices: prolific filters choices %s", mc.Total, filterID)
	default:
		return ""
	}
}

// quoteQuery quotes a multi-word query so the suggested command can be pasted
// into a shell as it stands.
func quoteQuery(query string) string {
	if strings.ContainsAny(query, " \t") {
		return strconv.Quote(query)
	}
	return query
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
