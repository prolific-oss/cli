package ui_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/prolific-oss/cli/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePager(t *testing.T) {
	tests := []struct {
		name          string
		prolificPager *string
		pager         *string
		want          string
	}{
		{name: "defaults to less", want: "less"},
		{name: "PAGER respected", pager: ptr("more"), want: "more"},
		{name: "PROLIFIC_PAGER wins", prolificPager: ptr("bat"), pager: ptr("more"), want: "bat"},
		{name: "empty PAGER disables", pager: ptr(""), want: ""},
		{name: "cat disables", pager: ptr("cat"), want: ""},
		{name: "empty PROLIFIC_PAGER disables even with PAGER", prolificPager: ptr(""), pager: ptr("more"), want: ""},
		{name: "whitespace trimmed", pager: ptr("  less -R  "), want: "less -R"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unsetEnv(t, "PROLIFIC_PAGER")
			unsetEnv(t, "PAGER")
			if tt.prolificPager != nil {
				t.Setenv("PROLIFIC_PAGER", *tt.prolificPager)
			}
			if tt.pager != nil {
				t.Setenv("PAGER", *tt.pager)
			}

			assert.Equal(t, tt.want, ui.ResolvePager())
		})
	}
}

func TestIsTerminalFalseForBuffer(t *testing.T) {
	assert.False(t, ui.IsTerminal(&bytes.Buffer{}))
}

func TestPageWritesDirectlyWhenNotATerminal(t *testing.T) {
	var b bytes.Buffer

	err := ui.Page(context.Background(), &b, func(w io.Writer) error {
		_, err := io.WriteString(w, "hello\n")
		return err
	})

	require.NoError(t, err)
	assert.Equal(t, "hello\n", b.String())
}

func TestPagePropagatesRenderError(t *testing.T) {
	boom := errors.New("boom")

	err := ui.Page(context.Background(), &bytes.Buffer{}, func(io.Writer) error { return boom })

	assert.ErrorIs(t, err, boom)
}

func TestRunPagerStreamsOutputThroughPager(t *testing.T) {
	var b bytes.Buffer

	err := ui.RunPager(context.Background(), "cat", &b, func(w io.Writer) error {
		_, err := io.WriteString(w, "line 1\nline 2\n")
		return err
	})

	require.NoError(t, err)
	assert.Equal(t, "line 1\nline 2\n", b.String())
}

func TestRunPagerTreatsEarlyQuitAsSuccess(t *testing.T) {
	// head exits after one line, closing the pipe while the renderer is still
	// writing, which is what happens when a user presses q in less.
	var b bytes.Buffer

	err := ui.RunPager(context.Background(), "head -n 1", &b, func(w io.Writer) error {
		for i := 0; i < 100000; i++ {
			if _, err := fmt.Fprintf(w, "line %d\n", i); err != nil {
				return err
			}
		}
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, "line 0\n", b.String())
}

func TestRunPagerTreatsCancellationAsSuccess(t *testing.T) {
	// Cancelling the context mid-render kills the pager, which is what
	// happens on ctrl-c. Neither the killed pager nor the resulting write
	// failures should be reported as an error.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := ui.RunPager(ctx, "sleep 30", &bytes.Buffer{}, func(w io.Writer) error {
		if _, err := io.WriteString(w, "first line\n"); err != nil {
			return err
		}
		cancel()
		for i := 0; i < 100000; i++ {
			if _, err := fmt.Fprintf(w, "line %d\n", i); err != nil {
				return err
			}
		}
		return nil
	})

	require.NoError(t, err)
}

func TestRunPagerStillReportsRealRenderErrors(t *testing.T) {
	boom := errors.New("boom")

	err := ui.RunPager(context.Background(), "cat", &bytes.Buffer{}, func(io.Writer) error { return boom })

	assert.ErrorIs(t, err, boom)
}

func TestRunPagerFallsBackWhenPagerMissing(t *testing.T) {
	var b bytes.Buffer

	err := ui.RunPager(context.Background(), "definitely-not-a-real-pager-binary", &b, func(w io.Writer) error {
		_, err := io.WriteString(w, "hello\n")
		return err
	})

	require.NoError(t, err)
	assert.Equal(t, "hello\n", b.String())
}

func TestRunPagerReportsPagerFailure(t *testing.T) {
	err := ui.RunPager(context.Background(), "false", &bytes.Buffer{}, func(w io.Writer) error {
		_, err := io.WriteString(w, "hello\n")
		if err != nil {
			return err
		}
		return nil
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "pager exited with an error")
}

func ptr[T any](v T) *T { return &v }

// unsetEnv makes key absent for the duration of the test and restores it after.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "") // registers restoration of the original value
	require.NoError(t, os.Unsetenv(key))
}
