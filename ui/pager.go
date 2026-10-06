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

// lessFlags are always passed on the command line when the pager is less.
// Command-line options combine with the user's LESS variable rather than
// being overridden by it, so these apply even when LESS is set:
//
//	F quit immediately if the output fits on one screen
//	R pass ANSI colour sequences through so highlights render
//	X don't clear the screen on exit, so output stays in the scrollback
const lessFlags = "-FRX"

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

// RunPager feeds render's output through the pager command line. The pager
// is started lazily on the first byte of output, so if render fails before
// producing anything (for example an API error on the first request) the
// error is returned without the pager ever taking over the screen. Once
// output has started it streams as it is produced, so the first screen
// appears before rendering finishes. If the pager cannot be started, output
// goes straight to w instead.
//
// The user quitting the pager early (for example pressing q in less) closes
// the pipe, which surfaces to render as a broken-pipe write error. Cancelling
// ctx (for example on ctrl-c) kills the pager, which surfaces as a signalled
// exit. Both are normal ways to finish and are not reported as failures.
func RunPager(ctx context.Context, pager string, w io.Writer, render func(io.Writer) error) error {
	parts := strings.Fields(pager)
	if len(parts) == 0 {
		return render(w)
	}

	lp := &lazyPager{ctx: ctx, args: pagerArgs(parts), out: w}
	renderErr := render(lp)

	if lp.cmd == nil {
		// Nothing was written, or the pager could not start and output went
		// straight to w. Either way there is no pager process to wait on.
		return renderErr
	}

	_ = lp.stdin.Close()
	waitErr := lp.cmd.Wait()

	if ctx.Err() != nil {
		// The pager was terminated because the command was cancelled; any
		// render or wait errors are consequences of that, not failures.
		return nil
	}
	if renderErr != nil && !isBrokenPipe(renderErr) {
		return renderErr
	}
	if waitErr != nil {
		return fmt.Errorf("pager exited with an error: %w", waitErr)
	}
	return nil
}

// pagerArgs returns the pager command line to execute. When the pager is
// less, the standard flags are appended so short output exits immediately
// and colours render, whatever the user's LESS variable says.
func pagerArgs(parts []string) []string {
	args := append([]string(nil), parts...)
	if filepath.Base(args[0]) == "less" {
		args = append(args, lessFlags)
	}
	return args
}

// lazyPager is an io.Writer that starts the pager process on the first Write
// and forwards subsequent writes to its stdin. If the pager fails to start,
// all writes fall through to out.
type lazyPager struct {
	ctx      context.Context
	args     []string
	out      io.Writer
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	fallback bool
}

func (lp *lazyPager) Write(p []byte) (int, error) {
	if lp.cmd == nil && !lp.fallback {
		lp.start()
	}
	if lp.fallback {
		return lp.out.Write(p)
	}
	return lp.stdin.Write(p)
}

func (lp *lazyPager) start() {
	cmd := exec.CommandContext(lp.ctx, lp.args[0], lp.args[1:]...) //nolint:gosec // pager comes from the user's own environment
	cmd.Stdout = lp.out
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		lp.fallback = true
		return
	}
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		lp.fallback = true
		return
	}

	lp.cmd = cmd
	lp.stdin = stdin
}

// isBrokenPipe reports whether err is the result of writing to a pipe whose
// reader has gone away.
func isBrokenPipe(err error) bool {
	return errors.Is(err, syscall.EPIPE) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed)
}
