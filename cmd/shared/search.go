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

// AddSearchFlags registers the flag surface shared by all search commands, so
// --limit, --all and --fields mean the same thing wherever they appear.
// defaultFields is the resource's default column selection, and workspaceUsage
// describes what scoping by workspace does for that resource.
func AddSearchFlags(cmd *cobra.Command, opts *SearchOptions, defaultFields, workspaceUsage string) {
	flags := cmd.Flags()
	flags.StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), workspaceUsage)
	flags.IntVarP(&opts.Limit, "limit", "l", client.DefaultRecordLimit, "Maximum number of records to return. Use 0 to fetch every match.")
	flags.BoolVarP(&opts.All, "all", "a", false, "Return every matching record (same as --limit 0)")
	flags.StringVarP(&opts.Fields, "fields", "f", defaultFields, "Comma separated fields to display in table or CSV output.")
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
