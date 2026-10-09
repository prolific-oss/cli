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
when --workspace is omitted, or default rules if no workspace is configured.`,
		Example: `prolific filters rule-tree --workspace <workspace-id>`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := c.GetFilterRuleTree(opts.WorkspaceID)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}

			// The rules sit under a top-level key so anything accompanying
			// them later has a home.
			encoder := json.NewEncoder(w)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(result); err != nil {
				return fmt.Errorf("error: %s", err)
			}

			return nil
		},
	}

	shared.AddWorkspaceFlag(cmd, &opts.WorkspaceID)
	// Accepted so --json can be passed uniformly, but nothing reads it: a tree
	// has no table or CSV form, so JSON is the only format and the default.
	cmd.Flags().BoolP("json", "j", false, "Output as JSON (the only format this command emits)")

	return cmd
}
