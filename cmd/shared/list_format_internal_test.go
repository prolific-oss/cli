package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// resolveFormat is exercised directly so the terminal branch can be covered
// without attaching the test process to a pty.
func TestResolveFormatOnATerminal(t *testing.T) {
	assert.Equal(t, FormatInteractive, resolveFormat(OutputOptions{}, true))
	assert.Equal(t, FormatTable, resolveFormat(OutputOptions{}, false))
	assert.Equal(t, FormatJSON, resolveFormat(OutputOptions{Json: true}, true))
}
