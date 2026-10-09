package shared

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// WorkspaceFlagUsage is the single help wording for --workspace.
const WorkspaceFlagUsage = "Workspace to scope the results to (defaults to your configured workspace)"

// ErrWorkspaceRequired is the message used wherever a command cannot run
// without a workspace.
var ErrWorkspaceRequired = errors.New("please provide a workspace ID")

// AddWorkspaceFlag registers --workspace / -w, defaulting to the configured
// workspace.
func AddWorkspaceFlag(cmd *cobra.Command, workspaceID *string) {
	cmd.Flags().StringVarP(workspaceID, "workspace", "w", viper.GetString("workspace"), WorkspaceFlagUsage)
}

// AddRequiredWorkspaceFlag registers the same flag for a command that cannot
// run without one. The flag stays optional, because the value can come from
// configuration, but the help says it is required.
func AddRequiredWorkspaceFlag(cmd *cobra.Command, workspaceID *string) {
	cmd.Flags().StringVarP(workspaceID, "workspace", "w", viper.GetString("workspace"), WorkspaceFlagUsage+" (required)")
}

// RequireWorkspace returns ErrWorkspaceRequired when no workspace was given.
func RequireWorkspace(workspaceID string) error {
	if strings.TrimSpace(workspaceID) == "" {
		return ErrWorkspaceRequired
	}
	return nil
}
