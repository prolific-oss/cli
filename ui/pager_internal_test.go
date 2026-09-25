package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagerArgs(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  []string
	}{
		{name: "less gets standard flags", parts: []string{"less"}, want: []string{"less", "-FRX"}},
		{name: "less with user flags keeps both", parts: []string{"less", "-S"}, want: []string{"less", "-S", "-FRX"}},
		{name: "less by full path", parts: []string{"/usr/bin/less"}, want: []string{"/usr/bin/less", "-FRX"}},
		{name: "other pagers untouched", parts: []string{"bat", "--paging=always"}, want: []string{"bat", "--paging=always"}},
		{name: "more untouched", parts: []string{"more"}, want: []string{"more"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := append([]string(nil), tt.parts...)
			assert.Equal(t, tt.want, pagerArgs(tt.parts))
			assert.Equal(t, original, tt.parts, "input must not be mutated")
		})
	}
}
