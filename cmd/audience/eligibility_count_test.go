package audience_test

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/audience"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// eligibility-count is the pre-1.3 surface kept working on top of the
// audience implementation. These tests pin the contract it has to keep: the
// flags it accepts, the messages it rejects with, and the one line it writes
// to stdout.

// runEligibilityCount executes the command and returns what it wrote to
// stdout, which is the part callers parse.
func runEligibilityCount(t *testing.T, count int, args ...string) string {
	t.Helper()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)
	c.EXPECT().GetEligibilityCount(gomock.Any()).Return(&client.EligibilityCountResponse{Count: count}, nil)

	var out bytes.Buffer
	cmd := audience.NewEligibilityCountCommand(c, &out)
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	require.NoError(t, cmd.Execute())

	return out.String()
}

func TestEligibilityCountKeepsItsOldFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cmd := audience.NewEligibilityCountCommand(mock_client.NewMockAPI(ctrl), nil)

	// -t means --template-path here, not --table as it does on audience count.
	template := cmd.Flags().ShorthandLookup("t")
	require.NotNil(t, template)
	assert.Equal(t, "template-path", template.Name)

	workspace := cmd.Flags().ShorthandLookup("w")
	require.NotNil(t, workspace)
	assert.Equal(t, "workspace", workspace.Name)

	// The output flags were never on this command.
	for _, name := range []string{"json", "csv", "table", "fields", "filters"} {
		assert.Nil(t, cmd.Flags().Lookup(name), "%s must not be on eligibility-count", name)
	}
}

func TestEligibilityCountWritesTheOldLine(t *testing.T) {
	out := runEligibilityCount(t, 1234, "-t", "testdata/filters.json", "-w", "ws-1")
	assert.Equal(t, "Eligible participants: 1234\n", out)
}

// The API reports counts below 25 as zero, and this command has always said so.
func TestEligibilityCountKeepsTheSub25Note(t *testing.T) {
	out := runEligibilityCount(t, 0, "-t", "testdata/filters.json", "-w", "ws-1")
	assert.Equal(t, "Eligible participants: 0 (or fewer than 25 — exact counts under 25 aren't shown, to protect participant privacy)\n", out)
}

func TestEligibilityCountRejectionsAreUnchanged(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "no template", args: []string{"-w", "ws-1"}, want: "error: a filter template is required, use -t/--template-path"},
		{name: "no workspace", args: []string{"-t", "testdata/filters.json"}, want: "error: workspace ID is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)
			c.EXPECT().GetEligibilityCount(gomock.Any()).Times(0)

			cmd := audience.NewEligibilityCountCommand(c, &bytes.Buffer{})
			cmd.SetArgs(tt.args)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			err := cmd.Execute()
			require.Error(t, err)
			assert.Equal(t, tt.want, err.Error())
		})
	}
}

func TestRenderEligibilityCount(t *testing.T) {
	assert.Equal(t, "Eligible participants: 42", audience.RenderEligibilityCount(42))
	assert.Contains(t, audience.RenderEligibilityCount(0), "fewer than 25")
}
