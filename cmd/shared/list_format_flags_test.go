package shared_test

import (
	"bytes"
	"testing"

	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveFormatForWriterPrefersAnExplicitFlag(t *testing.T) {
	cases := map[string]struct {
		opts shared.OutputOptions
		want string
	}{
		"json":  {opts: shared.OutputOptions{Json: true}, want: shared.FormatJSON},
		"csv":   {opts: shared.OutputOptions{Csv: true}, want: shared.FormatCSV},
		"table": {opts: shared.OutputOptions{Table: true}, want: shared.FormatTable},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, shared.ResolveFormatForWriter(tc.opts, &bytes.Buffer{}))
		})
	}
}

func TestResolveFormatForWriterFallsBackToATableWhenNotATerminal(t *testing.T) {
	assert.Equal(t, shared.FormatTable, shared.ResolveFormatForWriter(shared.OutputOptions{}, &bytes.Buffer{}))
}

func TestAddFieldsFlagRegistersTheStandardFlag(t *testing.T) {
	var fields string
	cmd := &cobra.Command{Use: "list"}

	shared.AddFieldsFlag(cmd, &fields, "ID,Name")

	flag := cmd.Flags().Lookup("fields")
	require.NotNil(t, flag)
	assert.Equal(t, "f", flag.Shorthand)
	assert.Equal(t, "ID,Name", flag.DefValue)
	assert.Equal(t, shared.FieldsFlagUsage, flag.Usage)
}
