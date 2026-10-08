package shared

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// WorkspaceFlagUsage is the single help wording for --workspace.
const WorkspaceFlagUsage = "Workspace to scope the results to (defaults to your configured workspace)"

// ErrWorkspaceRequired is the single message used wherever a command cannot
// run without a workspace.
var ErrWorkspaceRequired = errors.New("please provide a workspace ID")

// AddWorkspaceFlag registers --workspace / -w on the given command,
// defaulting to the configured workspace.
func AddWorkspaceFlag(cmd *cobra.Command, workspaceID *string) {
	cmd.Flags().StringVarP(workspaceID, "workspace", "w", viper.GetString("workspace"), WorkspaceFlagUsage)
}

// RequireWorkspace returns ErrWorkspaceRequired when no workspace was given,
// for the commands that cannot run without one.
func RequireWorkspace(workspaceID string) error {
	if strings.TrimSpace(workspaceID) == "" {
		return ErrWorkspaceRequired
	}
	return nil
}
