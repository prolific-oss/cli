package filters

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
)

var dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.DarkGrey))

// HighlightSpan is a half-open [Start, End) range of Unicode code points to
// highlight within a piece of text.
type HighlightSpan struct {
	Start int
	End   int
}

// HighlightText applies the given spans to text, wrapping each matched range
// in the standard highlight style. Offsets are zero-based code-point positions,
// as returned by the API. Overlapping or adjacent spans are merged and
// out-of-range spans are clamped, so malformed offsets never panic.
func HighlightText(text string, spans []HighlightSpan) string {
	return highlightWith(text, spans, func(plain string) string { return plain })
}

// highlightWith is HighlightText with a style applied to the non-highlighted
// segments. Styling each plain segment separately, rather than wrapping the
// whole string, stops the highlight's ANSI reset from cancelling the outer
// style for the rest of the line.
func highlightWith(text string, spans []HighlightSpan, plain func(string) string) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}
	if len(spans) == 0 {
		return plain(text)
	}

	clamped := make([]HighlightSpan, 0, len(spans))
	for _, s := range spans {
		start, end := max(s.Start, 0), min(s.End, len(runes))
		if start < end {
			clamped = append(clamped, HighlightSpan{Start: start, End: end})
		}
	}
	if len(clamped) == 0 {
		return plain(text)
	}

	sort.Slice(clamped, func(i, j int) bool { return clamped[i].Start < clamped[j].Start })

	merged := []HighlightSpan{clamped[0]}
	for _, s := range clamped[1:] {
		last := &merged[len(merged)-1]
		if s.Start <= last.End {
			last.End = max(last.End, s.End)
			continue
		}
		merged = append(merged, s)
	}

	var b strings.Builder
	cursor := 0
	writePlain := func(segment string) {
		if segment != "" {
			b.WriteString(plain(segment))
		}
	}
	for _, s := range merged {
		writePlain(string(runes[cursor:s.Start]))
		b.WriteString(ui.RenderHighlightedText(string(runes[s.Start:s.End])))
		cursor = s.End
	}
	writePlain(string(runes[cursor:]))

	return b.String()
}

// spansForField collects the highlight spans that apply to a given field.
func spansForField(highlights []model.FilterSearchHighlight, field string) []HighlightSpan {
	var spans []HighlightSpan
	for _, h := range highlights {
		if h.Field == field {
			spans = append(spans, HighlightSpan{Start: h.Start, End: h.End})
		}
	}
	return spans
}

// highlightField renders the value of a field with its matches highlighted.
func highlightField(value string, highlights []model.FilterSearchHighlight, field string) string {
	return HighlightText(value, spansForField(highlights, field))
}

// ruleWidth is the width of the divider drawn between results.
const ruleWidth = 60

// fieldWidth is the width of the field label column in a result.
const fieldWidth = 13

// RenderSearchHeader renders the title block shown above search results, so the
// query and total are visible on the first screen without scrolling.
func RenderSearchHeader(query string, shown, total int) string {
	var b strings.Builder
	b.WriteString(ui.RenderHeading(fmt.Sprintf("Filters matching %q", query)))
	b.WriteString("\n")

	summary := fmt.Sprintf("Showing %d of %d %s", shown, total, pluralise(total, "result", "results"))
	if shown < total {
		summary += ". Use --limit or --all to see more"
	}
	b.WriteString(dimStyle.Render(summary))
	b.WriteString("\n\n")

	return b.String()
}

// RenderSearchRule renders the divider drawn between results.
func RenderSearchRule() string {
	return dimStyle.Render(strings.Repeat("─", ruleWidth)) + "\n"
}

// RenderSearchResult renders a single filter search result at the given rank,
// highlighting the parts of the filter that matched the query and previewing
// matching choices.
func RenderSearchResult(rank int, r model.FilterSearchResult) string {
	var content strings.Builder
	h := r.Match.Highlights
	indent := "   "

	// Title line: rank and highlighted title.
	content.WriteString(dimStyle.Render(fmt.Sprintf("%d.", rank)))
	content.WriteString(" ")
	content.WriteString(highlightWith(r.Title, spansForField(h, "title"), ui.RenderHeading))
	content.WriteString("\n")

	// Subtitle line: type, data type and category, dimmed.
	meta := []string{r.Type, r.DataType}
	if category := renderCategory(r); category != "" {
		meta = append(meta, category)
	}
	content.WriteString(indent)
	content.WriteString(dimStyle.Render(strings.Join(meta, " · ")))
	content.WriteString("\n")

	field := func(label, value string) {
		content.WriteString(indent)
		content.WriteString(fmt.Sprintf("%-*s%s\n", fieldWidth, label, value))
	}

	field("Filter ID", highlightField(r.FilterID, h, "filter_id"))
	if r.Question != nil && *r.Question != "" {
		field("Question", highlightField(*r.Question, h, "question"))
	}
	if r.Description != "" {
		field("Description", highlightField(r.Description, h, "description"))
	}
	if r.Type == "range" && (r.Min != nil || r.Max != nil) {
		field("Range", renderRange(r.Min, r.Max))
	}
	if r.NumChoices != nil {
		choices := fmt.Sprintf("%d", *r.NumChoices)
		if r.MatchedChoices != nil {
			choices += dimStyle.Render(fmt.Sprintf(" (%d matching)", r.MatchedChoices.Matched))
		}
		field("Choices", choices)
	}

	if r.MatchedChoices != nil && len(r.MatchedChoices.Results) > 0 {
		content.WriteString(renderMatchedChoices(*r.MatchedChoices, indent))
	}

	if len(r.Match.Fields) > 0 {
		content.WriteString(indent)
		content.WriteString(dimStyle.Render(fmt.Sprintf("%-*s%s", fieldWidth, "Matched on", strings.Join(r.Match.Fields, ", "))))
		content.WriteString("\n")
	}

	return content.String()
}

func pluralise(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

func renderCategory(r model.FilterSearchResult) string {
	var parts []string
	if r.Category != nil && *r.Category != "" {
		parts = append(parts, highlightField(*r.Category, r.Match.Highlights, "category"))
	}
	if r.Subcategory != nil && *r.Subcategory != "" {
		parts = append(parts, highlightField(*r.Subcategory, r.Match.Highlights, "subcategory"))
	}
	return strings.Join(parts, " / ")
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

func renderMatchedChoices(mc model.FilterMatchingChoices, indent string) string {
	var b strings.Builder
	choiceIndent := indent + strings.Repeat(" ", fieldWidth)

	for _, choice := range mc.Results {
		label := HighlightText(choice.Label, spansForField(choice.Match.Highlights, "label"))
		b.WriteString(choiceIndent)
		b.WriteString(dimStyle.Render(choice.ID))
		b.WriteString("  ")
		b.WriteString(label)
		if choice.NumDescendants > 0 {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  (+%d nested)", choice.NumDescendants)))
		}
		b.WriteString("\n")
	}
	if mc.Truncated {
		remaining := mc.Matched - len(mc.Results)
		b.WriteString(choiceIndent)
		b.WriteString(dimStyle.Render(fmt.Sprintf("…and %d more matching %s", remaining, pluralise(remaining, "choice", "choices"))))
		b.WriteString("\n")
	}
	return b.String()
}
