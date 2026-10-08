package ui

import (
	"encoding/json"
	"io"
)

// Envelope is the JSON shape `--json` output is being standardised on. It is
// owned by the CLI rather than mirrored from the API, so a server-side change
// to the response envelope never reaches our output. Commands that already
// emit something else keep doing so until the next major version, so this is
// not yet the shape of every command's output.
//
// Count is the total number of matching records the API reports, which can be
// larger than the records returned.
//
// Limit and Offset describe the window the caller asked for, not the records
// that came back: a limit of 200 over a count of 90 returns 90 records and
// still reports a limit of 200. A Limit of zero means no limit was asked for
// or could be applied — through --all, through --limit 0, or because the
// endpoint does not paginate — and Results therefore holds everything that
// matched.
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
