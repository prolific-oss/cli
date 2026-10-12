package shared

import (
	"testing"

	"github.com/prolific-oss/cli/ui"

	"github.com/stretchr/testify/assert"
)

// resolveFormat is exercised directly so the terminal branch can be covered
// without attaching the test process to a pty.
func TestResolveFormatOnATerminal(t *testing.T) {
	assert.Equal(t, ui.FormatInteractive, resolveFormat(OutputOptions{}, true))
	assert.Equal(t, ui.FormatTable, resolveFormat(OutputOptions{}, false))
	assert.Equal(t, ui.FormatJSON, resolveFormat(OutputOptions{Json: true}, true))
}
