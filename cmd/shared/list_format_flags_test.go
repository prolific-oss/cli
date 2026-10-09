package shared_test

import (
	"bytes"
	"testing"

	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveFormatForWriterPrefersAnExplicitFlag(t *testing.T) {
	cases := map[string]struct {
		opts shared.OutputOptions
		want string
	}{
		"json":  {opts: shared.OutputOptions{Json: true}, want: ui.FormatJSON},
		"csv":   {opts: shared.OutputOptions{Csv: true}, want: ui.FormatCSV},
		"table": {opts: shared.OutputOptions{Table: true}, want: ui.FormatTable},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, shared.ResolveFormatForWriter(tc.opts, &bytes.Buffer{}))
		})
	}
}

func TestResolveFormatForWriterFallsBackToATableWhenNotATerminal(t *testing.T) {
	assert.Equal(t, ui.FormatTable, shared.ResolveFormatForWriter(shared.OutputOptions{}, &bytes.Buffer{}))
}

func TestAddFieldsFlagRegistersTheStandardFlag(t *testing.T) {
	var fields string
	cmd := &cobra.Command{Use: "list"}

	shared.AddFieldsFlag(cmd, &fields)

	flag := cmd.Flags().Lookup("fields")
	require.NotNil(t, flag)
	assert.Equal(t, "f", flag.Shorthand)
	assert.Equal(t, shared.FieldsFlagUsage, flag.Usage)
	// The default depends on the format, so the flag carries none.
	assert.Empty(t, flag.DefValue)
}

func TestFieldSetResolve(t *testing.T) {
	fields := ui.FieldSet{CSV: "ID,Name,Extra", Table: "ID,Name"}

	t.Run("a CSV takes every column worth having", func(t *testing.T) {
		assert.Equal(t, "ID,Name,Extra", fields.Resolve("", ui.FormatCSV))
	})

	t.Run("a table takes the few that identify a record", func(t *testing.T) {
		assert.Equal(t, "ID,Name", fields.Resolve("", ui.FormatTable))
	})

	t.Run("--fields wins over both", func(t *testing.T) {
		assert.Equal(t, "Name", fields.Resolve("Name", ui.FormatCSV))
		assert.Equal(t, "Name", fields.Resolve("Name", ui.FormatTable))
	})
}
