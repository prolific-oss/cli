package shared

import (
	"fmt"
	"io"

	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// RenderRecords writes a page of records in the format selected by out.
//
// JSON renders every field of every record as an array, matching the shape
// every other list command emits — a nil slice renders as [], never null, and
// no API pagination envelope is exposed. CSV and the table render only the
// requested fields; the table is followed by a record counter, where total is
// the number of matching records across all pages.
//
// Commands call this instead of switching on the format themselves, so the
// output contract stays identical across the CLI.
// RenderRecordsPaged renders like RenderRecords, but when no format flag was
// given and the output is an interactive terminal it sends the table through
// the user's pager, so a large result set doesn't scroll past. An explicit
// --json/--csv/--table, or --no-pager, writes straight out: that is scripted
// or machine-read output. This mirrors what `filters search` already does.
func RenderRecordsPaged[T any](cmd *cobra.Command, w io.Writer, out OutputOptions, fields string, records []T, total int) error {
	if ResolveFormat(out) != "" || NoPager(cmd) {
		return RenderRecords(w, out, fields, records, total)
	}

	return ui.Page(cmd.Context(), w, func(paged io.Writer) error {
		return RenderRecords(paged, out, fields, records, total)
	})
}

func RenderRecords[T any](w io.Writer, out OutputOptions, fields string, records []T, total int) error {
	switch ResolveFormat(out) {
	case "json":
		return ui.JSONRenderer[T]{}.Render(records, w)
	case "csv":
		return ui.CsvRenderer[T]{}.Render(records, fields, w)
	default:
		if err := (ui.TableRenderer[T]{}).Render(records, fields, w); err != nil {
			return err
		}
		_, err := fmt.Fprintf(w, "\n%s\n", ui.RenderRecordCounter(len(records), total))
		return err
	}
}
