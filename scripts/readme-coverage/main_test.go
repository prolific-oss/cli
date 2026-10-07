package main

import (
	"os"
	"path/filepath"
	"testing"
)

const table = `package contracttest

var operations = []operation{
	// Workspaces
	{operationID: "workspaces_GetWorkspaces", call: func(c *client.Client) { c.GetWorkspaces(10, 0) }},

	// Studies
	// SPECGAP: SearchStudies hits this operation with search, page and
	// workspace_id, none of which openapi.yaml declares for it.
	{operationID: "studies_GetStudies", call: func(c *client.Client) { c.GetStudies("", "") }},
	{operationID: "studies_GetStudy", skip: "OUTOFSCOPE: no CLI command"},
}
`

func writeTable(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "contract_test.go")
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	return path
}

// A multi-line note about an entry used to be read as a section label, which
// replaced the section heading in the generated README with the note's first line.
func TestParseOperationsIgnoresEntryNotesWhenNamingSections(t *testing.T) {
	entries, err := parseOperations(writeTable(t, table))
	if err != nil {
		t.Fatalf("parseOperations: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	for _, tc := range []struct {
		index           int
		operationID     string
		section         string
		clientMethod    string
		skipTag, reason string
	}{
		{index: 0, operationID: "workspaces_GetWorkspaces", section: "Workspaces", clientMethod: "GetWorkspaces"},
		{index: 1, operationID: "studies_GetStudies", section: "Studies", clientMethod: "GetStudies"},
		{index: 2, operationID: "studies_GetStudy", section: "Studies", skipTag: "OUTOFSCOPE", reason: "no CLI command"},
	} {
		got := entries[tc.index]
		if got.operationID != tc.operationID {
			t.Errorf("entry %d operationID = %q, want %q", tc.index, got.operationID, tc.operationID)
		}
		if got.section != tc.section {
			t.Errorf("entry %d section = %q, want %q", tc.index, got.section, tc.section)
		}
		if got.clientMethod != tc.clientMethod {
			t.Errorf("entry %d clientMethod = %q, want %q", tc.index, got.clientMethod, tc.clientMethod)
		}
		if got.skipTag != tc.skipTag {
			t.Errorf("entry %d skipTag = %q, want %q", tc.index, got.skipTag, tc.skipTag)
		}
		if got.skipReason != tc.reason {
			t.Errorf("entry %d skipReason = %q, want %q", tc.index, got.skipReason, tc.reason)
		}
	}
}

func TestIsSectionLabelRejectsNoteContinuationLines(t *testing.T) {
	// The second line of a SPECGAP note starts with a capital and carries no
	// tag, so only its position inside the comment block rules it out.
	lines := []string{"", "	// SPECGAP: something", "	// Those parameters are not validated.", "	{operationID: \"x\"},"}
	if isSectionLabel(lines, 2, "Those parameters are not validated.") {
		t.Error("a note continuation line must not be read as a section label")
	}
}

func TestIsSectionLabel(t *testing.T) {
	for _, tc := range []struct {
		text string
		want bool
	}{
		{text: "Studies", want: true},
		{text: "Participant Groups", want: true},
		{text: "AI Task Builder", want: true},
		{text: "AI Task Builder — Collections, Exports and Batches", want: true},
		{text: "SPECGAP: SearchStudies hits this operation", want: false},
		{text: "HARNESSGAP: kin-openapi mis-decodes a param", want: false},
	} {
		lines := []string{"", "	// " + tc.text, "	{operationID: \"x\"},"}
		if got := isSectionLabel(lines, 1, tc.text); got != tc.want {
			t.Errorf("isSectionLabel(%q) = %v, want %v", tc.text, got, tc.want)
		}
	}
}
