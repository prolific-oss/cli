package ui

// The formats a list command can render in. FormatInteractive is only ever
// resolved for a terminal, so output in a pipe is always machine readable.
const (
	FormatJSON        = "json"
	FormatCSV         = "csv"
	FormatTable       = "table"
	FormatInteractive = "interactive"
)

// FieldSet holds a command's default columns for each format. A CSV is read
// by a program, so it carries every column worth having; a table is read on a
// screen, so it carries the few that identify a record.
type FieldSet struct {
	CSV   string
	Table string
}

// Resolve returns the columns to render: whatever --fields asked for, or this
// format's default when it was not given.
func (f FieldSet) Resolve(requested, format string) string {
	switch {
	case requested != "":
		return requested
	case format == FormatCSV:
		return f.CSV
	default:
		return f.Table
	}
}
