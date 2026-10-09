package ui

// The formats a list command can render in.
const (
	FormatJSON        = "json"
	FormatCSV         = "csv"
	FormatTable       = "table"
	FormatInteractive = "interactive"
)

// FieldSet holds a command's default columns for each format. A CSV carries
// every column worth having; a table carries the few that identify a record.
type FieldSet struct {
	CSV   string
	Table string
}

// Resolve returns whatever --fields asked for, or this format's default.
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
