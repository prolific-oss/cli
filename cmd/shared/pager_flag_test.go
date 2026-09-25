package shared_test

import (
	"testing"

	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoPagerFalseWhenFlagNotRegistered(t *testing.T) {
	cmd := &cobra.Command{Use: "child"}
	assert.False(t, shared.NoPager(cmd))
}

func TestNoPagerReadsPersistentRootFlag(t *testing.T) {
	// run builds a fresh root/child tree so parsed flag values do not leak
	// between cases, and returns what the child observed.
	run := func(t *testing.T, args ...string) bool {
		t.Helper()
		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().Bool(shared.NoPagerFlag, false, "")

		var got bool
		child := &cobra.Command{Use: "child", RunE: func(cmd *cobra.Command, _ []string) error {
			got = shared.NoPager(cmd)
			return nil
		}}
		root.AddCommand(child)

		root.SetArgs(append([]string{"child"}, args...))
		require.NoError(t, root.Execute())
		return got
	}

	assert.True(t, run(t, "--no-pager"))
	assert.False(t, run(t))
}
