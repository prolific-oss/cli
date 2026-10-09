package shared_test

import (
	"testing"

	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddWorkspaceFlagRegistersTheStandardFlag(t *testing.T) {
	var workspaceID string
	cmd := &cobra.Command{Use: "list"}

	shared.AddWorkspaceFlag(cmd, &workspaceID)

	flag := cmd.Flags().Lookup("workspace")
	require.NotNil(t, flag)
	assert.Equal(t, "w", flag.Shorthand)
	assert.Equal(t, shared.WorkspaceFlagUsage, flag.Usage)
}

func TestAddWorkspaceFlagDefaultsToTheConfiguredWorkspace(t *testing.T) {
	viper.Set("workspace", "ws-from-config")
	defer viper.Set("workspace", "")

	var workspaceID string
	cmd := &cobra.Command{Use: "list"}
	shared.AddWorkspaceFlag(cmd, &workspaceID)

	require.NoError(t, cmd.Flags().Parse(nil))
	assert.Equal(t, "ws-from-config", workspaceID)
}

func TestAddWorkspaceFlagOverridesTheConfiguredWorkspace(t *testing.T) {
	viper.Set("workspace", "ws-from-config")
	defer viper.Set("workspace", "")

	var workspaceID string
	cmd := &cobra.Command{Use: "list"}
	shared.AddWorkspaceFlag(cmd, &workspaceID)

	require.NoError(t, cmd.Flags().Parse([]string{"-w", "ws-from-flag"}))
	assert.Equal(t, "ws-from-flag", workspaceID)
}

func TestRequireWorkspace(t *testing.T) {
	for name, workspaceID := range map[string]string{
		"empty":      "",
		"whitespace": "   ",
	} {
		t.Run(name, func(t *testing.T) {
			assert.ErrorIs(t, shared.RequireWorkspace(workspaceID), shared.ErrWorkspaceRequired)
		})
	}

	t.Run("provided", func(t *testing.T) {
		assert.NoError(t, shared.RequireWorkspace("ws-id"))
	})
}
