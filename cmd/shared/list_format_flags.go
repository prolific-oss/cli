package shared

import (
	"io"

	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
)

// The formats a list command can resolve to. FormatInteractive is only ever
// resolved for a terminal, so output in a pipe is always machine readable.
const (
	FormatJSON        = "json"
	FormatCSV         = "csv"
	FormatTable       = "table"
	FormatInteractive = "interactive"
)

// FieldsFlagUsage is the single help wording for --fields.
const FieldsFlagUsage = "Comma-separated list of columns for table or CSV output"

// OutputOptions holds the output format flags for list commands.
type OutputOptions struct {
	Json  bool
	Csv   bool
	Table bool
}

// AddOutputFlags registers --json / -j, --table / -t, --csv / -c on the given command.
// --non-interactive / -n is registered as a hidden alias for --table for backwards compatibility.
func AddOutputFlags(cmd *cobra.Command, opts *OutputOptions) {
	cmd.Flags().BoolVarP(&opts.Json, "json", "j", false, "Output as JSON")
	cmd.Flags().BoolVarP(&opts.Table, "table", "t", false, "Output as table (non-interactive)")
	cmd.Flags().BoolVarP(&opts.Csv, "csv", "c", false, "Output as CSV")

	cmd.Flags().BoolVarP(&opts.Table, "non-interactive", "n", false, "Output as table (non-interactive)")
	_ = cmd.Flags().MarkHidden("non-interactive")
}

// AddFieldsFlag registers --fields / -f on the given command, with one help
// wording across the CLI. defaultFields is the column set used when the flag
// is not given.
func AddFieldsFlag(cmd *cobra.Command, fields *string, defaultFields string) {
	cmd.Flags().StringVarP(fields, "fields", "f", defaultFields, FieldsFlagUsage)
}

// ResolveFormat returns the resolved format string based on the flags set.
// Priority: json > csv > table. Returns "" to indicate auto (TUI if TTY, else table).
func ResolveFormat(opts OutputOptions) string {
	switch {
	case opts.Json:
		return FormatJSON
	case opts.Csv:
		return FormatCSV
	case opts.Table:
		return FormatTable
	default:
		return ""
	}
}

// ResolveFormatForWriter resolves the output format for w. An explicit flag always
// wins. Without one it returns FormatInteractive when w is a terminal, and
// FormatTable otherwise, so piping a command never lands the caller in an
// interactive UI that cannot work in a pipe.
func ResolveFormatForWriter(opts OutputOptions, w io.Writer) string {
	return resolveFormat(opts, ui.IsTerminal(w))
}

func resolveFormat(opts OutputOptions, isTerminal bool) string {
	if format := ResolveFormat(opts); format != "" {
		return format
	}
	if isTerminal {
		return FormatInteractive
	}
	return FormatTable
}
