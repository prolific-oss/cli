package shared

import (
	"github.com/spf13/cobra"
)

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

// ResolveFormat returns the resolved format string based on the flags set.
// Priority: json > csv > table. Returns "" to indicate auto (TUI if TTY, else table).
func ResolveFormat(opts OutputOptions) string {
	switch {
	case opts.Json:
		return "json"
	case opts.Csv:
		return "csv"
	case opts.Table:
		return "table"
	default:
		return ""
	}
}
