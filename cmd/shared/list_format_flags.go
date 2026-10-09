package shared

import (
	"io"

	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
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

// AddFieldsFlag registers --fields / -f. It carries no default: the default
// depends on the format, so ui.FieldSet.Resolve supplies it.
func AddFieldsFlag(cmd *cobra.Command, fields *string) {
	cmd.Flags().StringVarP(fields, "fields", "f", "", FieldsFlagUsage)
}

// ResolveFormat returns the resolved format based on the flags set.
// Priority: json > csv > table. Returns "" to indicate auto (TUI if TTY, else table).
func ResolveFormat(opts OutputOptions) string {
	switch {
	case opts.Json:
		return ui.FormatJSON
	case opts.Csv:
		return ui.FormatCSV
	case opts.Table:
		return ui.FormatTable
	default:
		return ""
	}
}

// ResolveFormatForWriter resolves the format for w. An explicit flag wins;
// without one, a terminal gets ui.FormatInteractive and a pipe gets
// ui.FormatTable, so piping never lands the caller in an interactive UI.
func ResolveFormatForWriter(opts OutputOptions, w io.Writer) string {
	return resolveFormat(opts, ui.IsTerminal(w))
}

func resolveFormat(opts OutputOptions, isTerminal bool) string {
	if format := ResolveFormat(opts); format != "" {
		return format
	}
	if isTerminal {
		return ui.FormatInteractive
	}
	return ui.FormatTable
}
