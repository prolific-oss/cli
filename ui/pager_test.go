package ui_test

import (
	"bytes"
	"errors"
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
		{name: "defaults to less", want: "less -FRX"},
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

	err := ui.Page(&b, func(w io.Writer) error {
		_, err := io.WriteString(w, "hello\n")
		return err
	})

	require.NoError(t, err)
	assert.Equal(t, "hello\n", b.String())
}

func TestPagePropagatesRenderError(t *testing.T) {
	boom := errors.New("boom")

	err := ui.Page(&bytes.Buffer{}, func(io.Writer) error { return boom })

	assert.ErrorIs(t, err, boom)
}

func ptr[T any](v T) *T { return &v }

// unsetEnv makes key absent for the duration of the test and restores it after.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "") // registers restoration of the original value
	require.NoError(t, os.Unsetenv(key))
}
