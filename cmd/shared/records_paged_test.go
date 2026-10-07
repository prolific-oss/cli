package shared_test

import (
	"bytes"
	"testing"

	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// A bytes.Buffer is never a terminal, so the pager is bypassed and output
// lands directly — which is also why every other test in this package can
// assert on it.
func TestRenderRecordsPagedWritesStraightOutWhenNotATerminal(t *testing.T) {
	for _, tc := range []struct {
		name string
		out  shared.OutputOptions
		want string
	}{
		{name: "no format flag", out: shared.OutputOptions{}, want: "Showing 1 record of 42"},
		{name: "explicit table", out: shared.OutputOptions{Table: true}, want: "Showing 1 record of 42"},
		{name: "json", out: shared.OutputOptions{Json: true}, want: `"Name": "Memory, pilot"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			cmd := &cobra.Command{Use: "search"}

			err := shared.RenderRecordsPaged(cmd, &output, tc.out, "Name",
				[]record{{ID: "id1", Name: "Memory, pilot"}}, 42)

			require.NoError(t, err)
			require.Contains(t, output.String(), tc.want)
		})
	}
}

// NoPager reports false when the flag is not registered, so a command built
// standalone must still render rather than erroring.
func TestRenderRecordsPagedHonoursNoPager(t *testing.T) {
	var output bytes.Buffer
	cmd := &cobra.Command{Use: "search"}
	cmd.Flags().Bool(shared.NoPagerFlag, true, "")

	err := shared.RenderRecordsPaged(cmd, &output, shared.OutputOptions{}, "Name",
		[]record{{ID: "id1", Name: "Memory, pilot"}}, 42)

	require.NoError(t, err)
	require.Contains(t, output.String(), "Memory, pilot")
}
