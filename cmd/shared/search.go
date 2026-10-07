package shared

import (
	"errors"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SearchOptions holds the flags common to every `search` sub-command. Search
// commands page through the API on the user's behalf, so they expose a single
// --limit (and its --all shorthand) rather than the underlying endpoint's
// limit/offset or page-number parameters.
type SearchOptions struct {
	WorkspaceID string
	Limit       int
	All         bool
	Fields      string
	Output      OutputOptions
}

// SearchFlags carries the per-resource parts of the shared search flag set:
// everything else about --limit, --all, --fields and the output flags is the
// same wherever they appear. A zero DefaultLimit means client.DefaultRecordLimit.
type SearchFlags struct {
	DefaultFields  string
	DefaultLimit   int
	WorkspaceUsage string
}

// AddSearchFlags registers the flag surface shared by all search commands, so
// --limit, --all and --fields mean the same thing in each of them.
func AddSearchFlags(cmd *cobra.Command, opts *SearchOptions, cfg SearchFlags) {
	defaultLimit := cfg.DefaultLimit
	if defaultLimit == 0 {
		defaultLimit = client.DefaultRecordLimit
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), cfg.WorkspaceUsage)
	flags.IntVarP(&opts.Limit, "limit", "l", defaultLimit, "Maximum number of records to return. Use 0 to fetch every match.")
	flags.BoolVarP(&opts.All, "all", "a", false, "Return every matching record (same as --limit 0)")
	flags.StringVarP(&opts.Fields, "fields", "f", cfg.DefaultFields, "Comma separated fields to display in table or CSV output.")
	AddOutputFlags(cmd, &opts.Output)

	cmd.MarkFlagsMutuallyExclusive("all", "limit")
}

// Want returns how many records to fetch, where zero means every match. It is
// the single place --limit and --all are reconciled.
func (o SearchOptions) Want() (int, error) {
	if o.Limit < 0 {
		return 0, errors.New("limit must be greater than or equal to 0")
	}
	if o.All {
		return 0, nil
	}
	return o.Limit, nil
}

// SearchQuery joins and trims the positional arguments into a search query, so
// an unquoted multi-word query works the same as a quoted one.
func SearchQuery(args []string) (string, error) {
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return "", errors.New("please provide a search query")
	}
	return query, nil
}
