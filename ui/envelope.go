package ui

import (
	"encoding/json"
	"io"
)

// Envelope is the JSON shape --json output is standardising on, owned by the
// CLI rather than mirrored from the API.
//
// Limit and Offset are the window that was asked for, not what came back: a
// limit of 200 over a count of 90 still reports 200. A Limit of zero means no
// limit applied, so Results holds everything that matched.
type Envelope[T any] struct {
	Results []T `json:"results"`
	Count   int `json:"count"`
	Limit   int `json:"limit"`
	Offset  int `json:"offset"`
}

// NewEnvelope wraps records in the CLI envelope. A nil slice is rendered as an
// empty array rather than null, so consumers always receive a list.
func NewEnvelope[T any](records []T, count, limit, offset int) Envelope[T] {
	if records == nil {
		records = []T{}
	}
	return Envelope[T]{Results: records, Count: count, Limit: limit, Offset: offset}
}

// JSONEnvelopeRenderer renders the CLI envelope as indented JSON.
type JSONEnvelopeRenderer[T any] struct{}

// Render writes the envelope to w.
func (r JSONEnvelopeRenderer[T]) Render(envelope Envelope[T], w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(envelope)
}
