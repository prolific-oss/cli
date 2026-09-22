package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/mattn/go-isatty"
)

// defaultPager is used when neither PROLIFIC_PAGER nor PAGER is set.
//
//	-F quit immediately if the output fits on one screen
//	-R pass ANSI colour sequences through so highlights render
//	-X don't clear the screen on exit, so output stays in the scrollback
const defaultPager = "less -FRX"

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
// interactive terminal and a pager is configured. Otherwise, or if the pager
// cannot be started, render writes straight to w.
//
// The pager reads from a pipe as render produces output, so the first screen
// appears immediately and the user can scroll through the rest.
func Page(w io.Writer, render func(io.Writer) error) error {
	pager := ResolvePager()
	if pager == "" || !IsTerminal(w) {
		return render(w)
	}

	parts := strings.Fields(pager)
	cmd := exec.CommandContext(context.Background(), parts[0], parts[1:]...) //nolint:gosec // pager comes from the user's own environment
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	cmd.Env = pagerEnv(os.Environ())

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return render(w)
	}
	if err := cmd.Start(); err != nil {
		return render(w)
	}

	renderErr := render(stdin)
	closeErr := stdin.Close()
	waitErr := cmd.Wait()

	if renderErr != nil {
		return renderErr
	}
	if closeErr != nil {
		return closeErr
	}
	if waitErr != nil {
		return fmt.Errorf("pager exited with an error: %w", waitErr)
	}
	return nil
}

// pagerEnv returns env with sensible defaults for less when the user has not
// configured it themselves, so the default flags apply even when PAGER=less.
func pagerEnv(env []string) []string {
	hasLess := false
	for _, kv := range env {
		if strings.HasPrefix(kv, "LESS=") {
			hasLess = true
		}
	}
	if !hasLess {
		env = append(env, "LESS=FRX")
	}
	return env
}
