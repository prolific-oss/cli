package shared

import (
	"errors"

	"github.com/spf13/cobra"
)

// PaginationOptions holds the pagination flags for list commands. Pages are
// fetched as needed, so Limit is how many records the caller wants, not the
// size of one request.
type PaginationOptions struct {
	Limit  int
	Offset int
	All    bool
}

// Want returns the number of records to fetch, where zero means every one.
// --all is the same request as --limit 0.
func (o PaginationOptions) Want() int {
	if o.All {
		return 0
	}
	return o.Limit
}

// Validate reports the one message the CLI uses for each bad bound.
func (o PaginationOptions) Validate() error {
	if o.Limit < 0 {
		return errors.New("limit must be greater than or equal to 0")
	}
	if o.Offset < 0 {
		return errors.New("offset must be greater than or equal to 0")
	}
	return nil
}

// AddPaginationFlags registers --limit / -l, --offset / -o and --all / -a.
// defaultLimit applies when --limit is not given, usually
// client.DefaultRecordLimit.
func AddPaginationFlags(cmd *cobra.Command, opts *PaginationOptions, defaultLimit int) {
	flags := cmd.Flags()
	flags.IntVarP(&opts.Limit, "limit", "l", defaultLimit, "Maximum number of records to return. Use 0 to fetch every record.")
	flags.IntVarP(&opts.Offset, "offset", "o", 0, "Number of records to skip before returning results")
	flags.BoolVarP(&opts.All, "all", "a", false, "Return every record (same as --limit 0)")

	cmd.MarkFlagsMutuallyExclusive("all", "limit")
}
