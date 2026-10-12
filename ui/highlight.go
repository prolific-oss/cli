package ui

import (
	"sort"
	"strings"
)

// HighlightSpan is a half-open [Start, End) range of Unicode code points to
// highlight within a piece of text.
type HighlightSpan struct {
	Start int
	End   int
}

// FieldHighlights maps a field name to the spans that matched within it, as
// returned by search endpoints. Build it once per record and render each
// field through it.
type FieldHighlights map[string][]HighlightSpan

// Render returns text with the spans recorded for field wrapped in the
// standard highlight style. Text with no spans is returned unchanged.
func (h FieldHighlights) Render(field, text string) string {
	return h.RenderStyled(field, text, nil)
}

// RenderStyled is Render with plain applied to the non-highlighted segments.
// Styling each plain segment separately, rather than wrapping the whole
// string, stops the highlight's ANSI reset from cancelling the outer style
// for the rest of the line. A nil plain leaves those segments unstyled.
//
// Offsets are zero-based code-point positions, so multi-byte characters are
// handled correctly. Overlapping or adjacent spans are merged and out-of-range
// spans are clamped, so malformed offsets never panic.
func (h FieldHighlights) RenderStyled(field, text string, plain func(string) string) string {
	if plain == nil {
		plain = func(s string) string { return s }
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	merged := mergeSpans(h[field], len(runes))
	if len(merged) == 0 {
		return plain(text)
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
		b.WriteString(RenderHighlightedText(string(runes[s.Start:s.End])))
		cursor = s.End
	}
	writePlain(string(runes[cursor:]))

	return b.String()
}

// mergeSpans clamps spans to [0, length), drops empty ones, and merges any
// that overlap or touch, returning them in ascending order.
func mergeSpans(spans []HighlightSpan, length int) []HighlightSpan {
	clamped := make([]HighlightSpan, 0, len(spans))
	for _, s := range spans {
		start, end := max(s.Start, 0), min(s.End, length)
		if start < end {
			clamped = append(clamped, HighlightSpan{Start: start, End: end})
		}
	}
	if len(clamped) == 0 {
		return nil
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
	return merged
}
