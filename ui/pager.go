package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mattn/go-isatty"
)

// defaultPager is used when neither PROLIFIC_PAGER nor PAGER is set.
const defaultPager = "less"

// defaultLessFlags are applied when the pager is less and the user has not
// configured LESS themselves:
//
//	F quit immediately if the output fits on one screen
//	R pass ANSI colour sequences through so highlights render
//	X don't clear the screen on exit, so output stays in the scrollback
const defaultLessFlags = "FRX"

// ResolvePager returns the pager command line to use, or "" to disable paging.
// PROLIFIC_PAGER takes precedence over PAGER. Setting either to an empty
// string, or to "cat", disables paging, matching git's behaviour.
func ResolvePager() string {
	for _, key := range []string{"PROLIFIC_PAGER", "PAGER"} {
		if value, ok := os.LookupEnv(key); ok {
			value = strings.TrimSpace(value)
			if value == "" || value == "cat" {
				return ""
			}
			return value
		}
	}
	return defaultPager
}

// IsTerminal reports whether w is attached to an interactive terminal.
func IsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

// Page runs render, sending its output through the user's pager when w is an
// interactive terminal and a pager is configured. Otherwise render writes
// straight to w. Cancelling ctx terminates the pager.
func Page(ctx context.Context, w io.Writer, render func(io.Writer) error) error {
	pager := ResolvePager()
	if pager == "" || !IsTerminal(w) {
		return render(w)
	}
	return RunPager(ctx, pager, w, render)
}

// RunPager starts the pager command line and feeds it render's output as it
// is produced, so the first screen appears before rendering finishes. If the
// pager cannot be started, render writes straight to w instead.
//
// The user quitting the pager early (for example pressing q in less) closes
// the pipe, which surfaces to render as a broken-pipe write error. That is a
// normal exit and is not reported as a failure.
func RunPager(ctx context.Context, pager string, w io.Writer, render func(io.Writer) error) error {
	parts := strings.Fields(pager)
	if len(parts) == 0 {
		return render(w)
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...) //nolint:gosec // pager comes from the user's own environment
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	cmd.Env = pagerEnv(parts[0])

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return render(w)
	}
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return render(w)
	}

	renderErr := render(stdin)
	_ = stdin.Close()
	waitErr := cmd.Wait()

	if renderErr != nil && !isBrokenPipe(renderErr) {
		return renderErr
	}
	if waitErr != nil {
		return fmt.Errorf("pager exited with an error: %w", waitErr)
	}
	return nil
}

// isBrokenPipe reports whether err is the result of writing to a pipe whose
// reader has gone away.
func isBrokenPipe(err error) bool {
	return errors.Is(err, syscall.EPIPE) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed)
}

// pagerEnv returns the environment for the pager. When the pager is less and
// the user has not set LESS, the default flags are supplied so the standard
// behaviour applies however less was selected.
func pagerEnv(pagerBinary string) []string {
	env := os.Environ()
	if filepath.Base(pagerBinary) != "less" {
		return env
	}
	if _, set := os.LookupEnv("LESS"); set {
		return env
	}
	return append(env, "LESS="+defaultLessFlags)
}
