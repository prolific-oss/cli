package shared_test

import (
	"testing"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestSearchQuery(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{name: "single word", args: []string{"memory"}, want: "memory"},
		{name: "trims surrounding space", args: []string{"  memory  "}, want: "memory"},
		{name: "joins an unquoted multi-word query", args: []string{"memory", "task"}, want: "memory task"},
		{name: "no arguments", args: nil, wantErr: true},
		{name: "whitespace only", args: []string{" "}, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query, err := shared.SearchQuery(tc.args)

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, query)
		})
	}
}

func TestSearchOptionsWant(t *testing.T) {
	for _, tc := range []struct {
		name    string
		opts    shared.SearchOptions
		want    int
		wantErr string
	}{
		{name: "an explicit limit is the number wanted", opts: shared.SearchOptions{Limit: 10}, want: 10},
		{name: "--all means every match", opts: shared.SearchOptions{Limit: 10, All: true}, want: 0},
		{name: "--limit 0 also means every match", opts: shared.SearchOptions{Limit: 0}, want: 0},
		{name: "a negative limit is rejected", opts: shared.SearchOptions{Limit: -1}, wantErr: "limit must be greater than or equal to 0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want, err := tc.opts.Want()

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, want)
		})
	}
}

func TestAddSearchFlagsRegistersTheCommonSurface(t *testing.T) {
	var opts shared.SearchOptions
	cmd := &cobra.Command{Use: "search"}

	shared.AddSearchFlags(cmd, &opts, "ID,Name", "Workspace to search.")

	for _, name := range []string{"workspace", "limit", "all", "fields", "json", "csv", "table"} {
		require.NotNil(t, cmd.Flags().Lookup(name), "flag %q must be registered", name)
	}
	require.Equal(t, "ID,Name", cmd.Flags().Lookup("fields").DefValue)
	require.Equal(t, "Workspace to search.", cmd.Flags().Lookup("workspace").Usage)

	require.NoError(t, cmd.Flags().Parse(nil))
	require.Equal(t, client.DefaultRecordLimit, opts.Limit)
}

func TestAddSearchFlagsMakesAllAndLimitMutuallyExclusive(t *testing.T) {
	var opts shared.SearchOptions
	cmd := &cobra.Command{Use: "search", RunE: func(*cobra.Command, []string) error { return nil }}
	shared.AddSearchFlags(cmd, &opts, "ID,Name", "Workspace to search.")
	cmd.SetArgs([]string{"--all", "--limit", "10"})

	require.ErrorContains(t, cmd.Execute(), "none of the others can be")
}
