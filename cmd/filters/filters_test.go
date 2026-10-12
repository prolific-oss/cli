package filters_test

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/filters"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFiltersCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := filters.NewFiltersCommand(c, nil)

	assert.Equal(t, "filters", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	var names []string
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}
	assert.ElementsMatch(t, []string{"list", "search", "choices", "rule-tree"}, names)
}

// `prolific filters` listed the catalogue before it gained subcommands, and
// still has to, so the invocation is not broken for anyone on an older version.
func TestFiltersWithNoSubcommandListsTheCatalogue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)
	c.EXPECT().GetFilters("").Return(&client.ListFiltersResponse{
		Results: []model.Filter{{FilterID: "age", Type: "range"}},
	}, nil)

	var out bytes.Buffer
	cmd := filters.NewFiltersCommand(c, &out)
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "age")
}

// `prolific filters -n` printed a detail block per filter. It is hidden now,
// but it still has to produce the same output.
func TestFiltersNonInteractiveRendersDetailBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)
	c.EXPECT().GetFilters("").Return(&client.ListFiltersResponse{
		Results: []model.Filter{{FilterID: "age", Question: "What is your date of birth?", Type: "range", DataType: "integer"}},
	}, nil)

	var out bytes.Buffer
	cmd := filters.NewFiltersCommand(c, &out)
	cmd.SetArgs([]string{"-n"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	require.NoError(t, cmd.Execute())

	rendered := out.String()
	assert.Contains(t, rendered, "Filter ID:         age")
	assert.Contains(t, rendered, "Question:          What is your date of birth?")
	assert.Contains(t, rendered, "Data Type:         integer")

	// Hidden, so the help points at `filters list` instead.
	flag := cmd.Flags().Lookup("non-interactive")
	require.NotNil(t, flag)
	assert.True(t, flag.Hidden)
}

// The legacy paths are kept so old invocations still work, so their
// deprecation notices must not land in the output those invocations parse.
func TestFiltersDeprecationNoticesGoToStderrOnly(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "no subcommand", args: []string{}, want: `Running "prolific filters" without a subcommand is deprecated`},
		{name: "non-interactive", args: []string{"-n"}, want: "Flag -n/--non-interactive is deprecated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)
			c.EXPECT().GetFilters("").Return(&client.ListFiltersResponse{
				Results: []model.Filter{{FilterID: "age", Type: "range"}},
			}, nil)

			var out, errOut bytes.Buffer
			cmd := filters.NewFiltersCommand(c, &out)
			cmd.SetArgs(tt.args)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&errOut)
			require.NoError(t, cmd.Execute())

			assert.Contains(t, errOut.String(), tt.want)
			assert.NotContains(t, out.String(), "deprecated")
		})
	}
}
