package shared

import "github.com/spf13/cobra"

// NoPagerFlag is the persistent root flag that disables piping long output
// into a pager.
const NoPagerFlag = "no-pager"

// NoPager reports whether --no-pager was set on the command or any of its
// parents. It is false when the flag is not registered, so commands can be
// exercised standalone in tests.
func NoPager(cmd *cobra.Command) bool {
	value, err := cmd.Flags().GetBool(NoPagerFlag)
	return err == nil && value
}
