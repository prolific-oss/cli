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
