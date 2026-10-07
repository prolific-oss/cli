package filters

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RuleTreeOptions holds the workspace and output settings for rule-tree.
type RuleTreeOptions struct {
	WorkspaceID string
	JSON        bool
}

// NewRuleTreeCommand creates the `filters rule-tree` command.
func NewRuleTreeCommand(c client.API, w io.Writer) *cobra.Command {
	var opts RuleTreeOptions
	cmd := &cobra.Command{
		Use:   "rule-tree",
		Short: "Get the rules for combining audience filters",
		Long: `Fetch the server's audience filter rules as JSON. Uses the configured workspace
when --workspace is omitted, or default rules if no workspace is configured.`,
		Example: `prolific filters rule-tree --workspace <workspace-id> --json`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := c.GetFilterRuleTree(opts.WorkspaceID)
			if err != nil {
				return fmt.Errorf("error: %s", err)
			}
			encoder := json.NewEncoder(w)
			if !opts.JSON {
				encoder.SetIndent("", "  ")
			}
			if err := encoder.Encode(result); err != nil {
				return fmt.Errorf("error: %s", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.WorkspaceID, "workspace", "w", viper.GetString("workspace"), "Workspace whose rules to retrieve (defaults to configured workspace)")
	cmd.Flags().BoolVarP(&opts.JSON, "json", "j", false, "Output compact JSON")
	return cmd
}
