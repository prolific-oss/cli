package ui

import (
	"fmt"
	"io"
	"os"
)

// Status shows a transient progress message on stderr while a command waits on
// the API, and returns a function that clears it. Nothing is shown unless
// stderr is an interactive terminal, so scripts and piped output are
// unaffected. The returned function is safe to call more than once.
func Status(message string) func() {
	return statusOn(os.Stderr, message)
}

func statusOn(w io.Writer, message string) func() {
	if !IsTerminal(w) {
		return func() {}
	}

	fmt.Fprint(w, RenderDimmed(message))

	cleared := false
	return func() {
		if cleared {
			return
		}
		cleared = true
		// Return to the start of the line and erase it.
		fmt.Fprint(w, "\r\033[K")
	}
}
