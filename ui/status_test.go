package ui

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusIsSilentWhenNotATerminal(t *testing.T) {
	var b bytes.Buffer

	clear := statusOn(&b, "Searching…")
	clear()
	clear()

	assert.Empty(t, b.String())
}
