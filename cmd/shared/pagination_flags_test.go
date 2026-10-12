package shared_test

import (
	"testing"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddPaginationFlagsRegistersTheStandardFlags(t *testing.T) {
	var opts shared.PaginationOptions
	cmd := &cobra.Command{Use: "list"}

	shared.AddPaginationFlags(cmd, &opts, client.DefaultRecordLimit)

	for name, shorthand := range map[string]string{"limit": "l", "offset": "o", "all": "a"} {
		flag := cmd.Flags().Lookup(name)
		require.NotNil(t, flag, "expected a --%s flag", name)
		assert.Equal(t, shorthand, flag.Shorthand)
	}

	assert.Equal(t, "200", cmd.Flags().Lookup("limit").DefValue)
	assert.Equal(t, "0", cmd.Flags().Lookup("offset").DefValue)
}

// Some endpoints default to a different page size than the CLI-wide 200.
func TestAddPaginationFlagsHonoursTheGivenDefaultLimit(t *testing.T) {
	var opts shared.PaginationOptions
	cmd := &cobra.Command{Use: "search"}

	shared.AddPaginationFlags(cmd, &opts, 25)

	assert.Equal(t, "25", cmd.Flags().Lookup("limit").DefValue)
}

func TestAddPaginationFlagsMakesAllAndLimitMutuallyExclusive(t *testing.T) {
	var opts shared.PaginationOptions
	cmd := &cobra.Command{Use: "list", RunE: func(*cobra.Command, []string) error { return nil }}
	shared.AddPaginationFlags(cmd, &opts, client.DefaultRecordLimit)

	cmd.SetArgs([]string{"--all", "--limit", "10"})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "[all limit] were all set")
}

func TestPaginationOptionsWant(t *testing.T) {
	// --all is the same request as --limit 0: everything.
	assert.Equal(t, 0, shared.PaginationOptions{Limit: 25, All: true}.Want())
	assert.Equal(t, 25, shared.PaginationOptions{Limit: 25}.Want())
	assert.Equal(t, 0, shared.PaginationOptions{Limit: 0}.Want())
}

func TestPaginationOptionsValidate(t *testing.T) {
	tests := map[string]struct {
		opts shared.PaginationOptions
		want string
	}{
		"negative limit":  {opts: shared.PaginationOptions{Limit: -1}, want: "limit must be greater than or equal to 0"},
		"negative offset": {opts: shared.PaginationOptions{Offset: -1}, want: "offset must be greater than or equal to 0"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.opts.Validate()
			require.Error(t, err)
			assert.EqualError(t, err, tt.want)
		})
	}

	t.Run("valid", func(t *testing.T) {
		assert.NoError(t, shared.PaginationOptions{Limit: 200, Offset: 50}.Validate())
	})
}
