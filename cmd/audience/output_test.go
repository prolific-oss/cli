package audience_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/audience"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// terminalWriter stands in for a terminal, so the human-facing output can be
// exercised without attaching the test process to a pty.
type terminalWriter struct{ io.Writer }

func (terminalWriter) IsTerminal() bool { return true }

// atTerminal presents w as a terminal and disables the pager, so the output a
// user would see lands in the buffer rather than in less.
func atTerminal(t *testing.T, w io.Writer) io.Writer {
	t.Helper()
	t.Setenv("PROLIFIC_PAGER", "")
	return terminalWriter{w}
}

func countFilters() string {
	return `[{"filter_id":"age","selected_range":{"lower":18,"upper":65}}]`
}

// runCount executes `audience count` with the given extra arguments and
// returns what it wrote.
func runCount(t *testing.T, isTerminal bool, args ...string) string {
	t.Helper()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)
	c.EXPECT().GetEligibilityCount(gomock.Any()).Return(&client.EligibilityCountResponse{Count: 1234}, nil)

	var b bytes.Buffer
	var w io.Writer = &b
	if isTerminal {
		w = atTerminal(t, &b)
	}

	cmd := audience.NewCountCommand(c, w)
	cmd.SetArgs(append([]string{"--filters", countFilters(), "-w", "ws-1"}, args...))
	require.NoError(t, cmd.Execute())

	return b.String()
}

func TestCountOutputFormats(t *testing.T) {
	// --json must stay byte-identical: the beta skills already consume it.
	t.Run("json", func(t *testing.T) {
		assert.Equal(t, "{\"count\":1234}\n", runCount(t, false, "--json"))
	})

	t.Run("csv", func(t *testing.T) {
		assert.Equal(t, "Count\n1234\n", runCount(t, false, "--csv"))
	})

	t.Run("table", func(t *testing.T) {
		assert.Contains(t, runCount(t, false, "--table"), "Count")
		assert.Contains(t, runCount(t, false, "--table"), "1234")
	})

	// At a terminal a count is a single number, so it stays a sentence.
	t.Run("terminal default is the sentence", func(t *testing.T) {
		assert.Equal(t, "Eligible participants: 1234\n", runCount(t, true))
	})

	// Piped, it has to be machine-readable like every other command.
	t.Run("piped default is a table", func(t *testing.T) {
		assert.Contains(t, runCount(t, false), "Count")
	})
}

func TestCountShorthandsFollowTheCLIWideMeanings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cmd := audience.NewCountCommand(mock_client.NewMockAPI(ctrl), nil)

	for name, shorthand := range map[string]string{
		"json":          "j",
		"csv":           "c",
		"table":         "t",
		"fields":        "f",
		"workspace":     "w",
		"template-path": "p",
	} {
		flag := cmd.Flags().Lookup(name)
		require.NotNil(t, flag, "expected a --%s flag", name)
		assert.Equal(t, shorthand, flag.Shorthand, "--%s", name)
	}
}

// runBreakdown executes `audience breakdown` and returns what it wrote.
func runBreakdown(t *testing.T, args ...string) string {
	t.Helper()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)
	c.EXPECT().GetFilterBreakdown(gomock.Any()).Return(&client.FilterBreakdownResponse{
		Breakdown: map[string]int{"0": 4, "1": 3, client.FilterBreakdownNAKey: 5},
	}, nil)

	var b bytes.Buffer
	cmd := audience.NewBreakdownCommand(c, &b)
	cmd.SetArgs(append([]string{
		"--filters", countFilters(),
		"--breakdown", `{"filter_id":"handedness","selected_values":["0","1"]}`,
		"-w", "ws-1",
	}, args...))
	require.NoError(t, cmd.Execute())

	return b.String()
}

func TestBreakdownOutputFormats(t *testing.T) {
	t.Run("json keeps the counts under a top-level key", func(t *testing.T) {
		assert.JSONEq(t, `{"breakdown":{"0":4,"1":3,"N/A":5}}`, runBreakdown(t, "--json"))
	})

	// A spreadsheet has to read this without reshaping: one row per value,
	// N/A last.
	t.Run("csv is one row per value", func(t *testing.T) {
		assert.Equal(t, "Value,Count\n0,4\n1,3\nN/A,5\n", runBreakdown(t, "--csv"))
	})

	t.Run("csv honours --fields", func(t *testing.T) {
		assert.Equal(t, "Count\n4\n3\n5\n", runBreakdown(t, "--csv", "-f", "Count"))
	})

	t.Run("table keeps N/A last", func(t *testing.T) {
		assert.Equal(t, "Value Count \n0     4     \n1     3     \nN/A   5     \n", runBreakdown(t, "--table"))
	})
}
