package shared_test

import (
	"bytes"
	"testing"

	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/stretchr/testify/require"
)

type record struct {
	ID   string
	Name string
}

func TestRenderRecords(t *testing.T) {
	records := []record{{ID: "id1", Name: "Memory, pilot"}}

	for _, tc := range []struct {
		name   string
		out    shared.OutputOptions
		want   string
		absent string
	}{
		{
			name: "json ignores the field selection and emits every field",
			out:  shared.OutputOptions{Json: true},
			want: "{\n    \"ID\": \"id1\",\n    \"Name\": \"Memory, pilot\"\n  }\n]\n",
		},
		{
			name: "csv quotes embedded commas and emits no counter",
			out:  shared.OutputOptions{Csv: true},
			want: "Name\n\"Memory, pilot\"\n",
		},
		{
			name: "table honours the field selection and appends a counter",
			out:  shared.OutputOptions{Table: true},
			want: "Showing 1 record of 42",
		},
		{
			name: "no format selected falls back to the table",
			out:  shared.OutputOptions{},
			want: "Showing 1 record of 42",
		},
		{
			name:   "table omits unselected fields",
			out:    shared.OutputOptions{Table: true},
			absent: "id1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer

			require.NoError(t, shared.RenderRecords(&output, tc.out, "Name", records, 42))

			if tc.want != "" {
				require.Contains(t, output.String(), tc.want)
			}
			if tc.absent != "" {
				require.NotContains(t, output.String(), tc.absent)
			}
		})
	}
}

func TestRenderRecordsAlwaysEmitsAJSONArray(t *testing.T) {
	var output bytes.Buffer

	require.NoError(t, shared.RenderRecords(&output, shared.OutputOptions{Json: true}, "Name", []record(nil), 0))

	require.JSONEq(t, "[]", output.String(), "a nil slice must not render as null")
}

func TestRenderRecordsCountsZeroRecordsAgainstTheTotal(t *testing.T) {
	var output bytes.Buffer

	require.NoError(t, shared.RenderRecords(&output, shared.OutputOptions{Table: true}, "Name", []record(nil), 42))

	require.Contains(t, output.String(), "Showing 0 records of 42")
}
