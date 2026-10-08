package filters

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/spf13/cobra"
)

// RuleTreeOptions holds the workspace and output settings for rule-tree.
type RuleTreeOptions struct {
	WorkspaceID string
}

// NewRuleTreeCommand creates the `filters rule-tree` command.
func NewRuleTreeCommand(c client.API, w io.Writer) *cobra.Command {
	var opts RuleTreeOptions

	cmd := &cobra.Command{
		Use:   "rule-tree",
		Short: "Get the rules for combining audience filters",
		Long: `Fetch the server's audience filter rules as JSON. Uses the configured workspace
when --workspace is omitted, or default rules if no workspace is configured.

The rules are a nested tree rather than a list, so JSON is the only format
this command emits. --json is accepted for consistency but has nothing to
select, and there are no --table or --csv flags, because a tree has no rows.`,
		Example: `
Fetch the rules for a workspace
$ prolific filters rule-tree --workspace <workspace-id>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := c.GetFilterRuleTree(opts.WorkspaceID)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			// The rules themselves are the payload; the API's wrapper key
			// adds nothing a caller has to reach through.
			encoder := json.NewEncoder(w)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(result.RuleTree); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	shared.AddWorkspaceFlag(cmd, &opts.WorkspaceID)
	// --json is accepted so a caller can pass it uniformly across commands,
	// but there is nothing for it to select: a tree has no table or CSV form,
	// so JSON is both the only format and the default. Nothing reads the
	// value, which is why it is not kept on the options struct.
	var acceptJSON bool
	cmd.Flags().BoolVarP(&acceptJSON, "json", "j", false, "Output as JSON (the only format this command emits)")

	return cmd
}
