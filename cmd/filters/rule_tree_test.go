package filters_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/filters"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleTreeCommand(t *testing.T) {
	for _, tc := range []struct {
		name, configured, workspace string
		args                        []string
		err                         error
	}{
		{name: "default rules"},
		{name: "configured workspace", configured: "configured", workspace: "configured"},
		{name: "explicit workspace", configured: "configured", workspace: "chosen", args: []string{"-w", "chosen", "--json"}},
		{name: "API error", err: errors.New("access denied")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			viper.Set("workspace", tc.configured)
			t.Cleanup(viper.Reset)
			c := mock_client.NewMockAPI(gomock.NewController(t))
			result := &client.FilterRuleTreeResponse{RuleTree: map[string]json.RawMessage{"and": json.RawMessage(`{"max_children":null,"children":{"leaf":true}}`)}}
			if tc.err != nil {
				// The real client returns no response alongside an error, so
				// the command must not reach into the result before checking.
				result = nil
			}
			c.EXPECT().GetFilterRuleTree(tc.workspace).Return(result, tc.err)
			var output bytes.Buffer
			cmd := filters.NewRuleTreeCommand(c, &output)
			cmd.SetArgs(append([]string{}, tc.args...))
			err := cmd.Execute()
			if tc.err != nil {
				require.EqualError(t, err, "error: access denied")
				require.Empty(t, output.String())
				return
			}
			require.NoError(t, err)
			require.JSONEq(t, `{"rule_tree":{"and":{"max_children":null,"children":{"leaf":true}}}}`, output.String())
		})
	}
}

// --json used to mean "compact", the opposite of what it means everywhere
// else. It now selects the format, and the rules are always indented.
func TestRuleTreeJSONSelectsFormatAndIsAlwaysIndented(t *testing.T) {
	result := &client.FilterRuleTreeResponse{
		RuleTree: map[string]json.RawMessage{"and": json.RawMessage(`{"children":{"leaf":true}}`)},
	}

	run := func(t *testing.T, args ...string) string {
		t.Helper()
		viper.Set("workspace", "")
		t.Cleanup(viper.Reset)

		c := mock_client.NewMockAPI(gomock.NewController(t))
		c.EXPECT().GetFilterRuleTree("").Return(result, nil)

		var output bytes.Buffer
		cmd := filters.NewRuleTreeCommand(c, &output)
		cmd.SetArgs(args)
		require.NoError(t, cmd.Execute())
		return output.String()
	}

	t.Run("default is indented", func(t *testing.T) {
		assert.Contains(t, run(t), "\n  ")
	})

	t.Run("--json is accepted and changes nothing", func(t *testing.T) {
		// It used to compact the output, which is the opposite of what the
		// flag means everywhere else.
		assert.Equal(t, run(t), run(t, "--json"))
	})

	t.Run("the rules sit under a top-level key", func(t *testing.T) {
		out := run(t)
		// A bare object would have nowhere to put anything that has to
		// accompany the rules later.
		assert.Contains(t, out, `"rule_tree"`)
		assert.Contains(t, out, `"and"`)
	})
}

func TestRuleTreeOffersNoTableOrCsv(t *testing.T) {
	c := mock_client.NewMockAPI(gomock.NewController(t))
	cmd := filters.NewRuleTreeCommand(c, nil)

	// A tree has no rows, so these formats are not offered at all rather
	// than registered and silently ignored.
	assert.Nil(t, cmd.Flags().Lookup("table"))
	assert.Nil(t, cmd.Flags().Lookup("csv"))
	assert.NotNil(t, cmd.Flags().Lookup("json"))
	assert.Nil(t, cmd.Flags().Lookup("compact"), "indentation is not a flag; the rules are always indented")
}
