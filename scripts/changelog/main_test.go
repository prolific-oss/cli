package main

import (
	"testing"
)

func TestExtractSection(t *testing.T) {
	changelog := `# CHANGELOG

## next

<!-- Add manual release notes here. -->

## 0.0.60

### Features

- Added feature A
- Added feature B

## 0.0.59

- Bug fix C
`

	tests := []struct {
		name          string
		section       string
		stripComments bool
		want          string
	}{
		{
			name:          "extract next section",
			section:       "next",
			stripComments: false,
			want:          "<!-- Add manual release notes here. -->\n",
		},
		{
			name:          "extract next section with strip comments",
			section:       "next",
			stripComments: true,
			want:          "",
		},
		{
			name:          "extract versioned section",
			section:       "0.0.60",
			stripComments: false,
			want: `### Features

- Added feature A
- Added feature B
`,
		},
		{
			name:          "section not found returns empty",
			section:       "0.0.99",
			stripComments: false,
			want:          "",
		},
		{
			name:          "version with dots matches exactly",
			section:       "0.0.6",
			stripComments: false,
			want:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractSection(changelog, tt.section, tt.stripComments)
			if got != tt.want {
				t.Errorf("ExtractSection(%q, strip=%v)\ngot:  %q\nwant: %q",
					tt.section, tt.stripComments, got, tt.want)
			}
		})
	}
}

func TestExtractSection_StripMultilineComments(t *testing.T) {
	changelog := `# CHANGELOG

## next

<!-- This is a
multi-line comment -->

Some actual content here.

## 0.0.1

- Initial
`

	got := ExtractSection(changelog, "next", true)
	want := "Some actual content here.\n"
	if got != want {
		t.Errorf("got:  %q\nwant: %q", got, want)
	}
}

func TestExtractSection_LastSection(t *testing.T) {
	changelog := `# CHANGELOG

## 0.0.1

- Initial release
`

	got := ExtractSection(changelog, "0.0.1", false)
	want := "- Initial release\n"
	if got != want {
		t.Errorf("got:  %q\nwant: %q", got, want)
	}
}

func TestMergeNotes(t *testing.T) {
	tests := []struct {
		name      string
		manual    string
		generated string
		fallback  string
		want      string
	}{
		{
			name:      "both present",
			manual:    "Manual notes",
			generated: "Generated notes",
			fallback:  "fallback",
			want:      "Manual notes\n\nGenerated notes\n",
		},
		{
			name:      "only manual",
			manual:    "Manual notes",
			generated: "",
			fallback:  "fallback",
			want:      "Manual notes\n",
		},
		{
			name:      "only generated",
			manual:    "",
			generated: "Generated notes",
			fallback:  "fallback",
			want:      "Generated notes\n",
		},
		{
			name:      "neither present uses fallback",
			manual:    "",
			generated: "",
			fallback:  "- Maintenance and dependency updates",
			want:      "- Maintenance and dependency updates\n",
		},
		{
			name:      "whitespace-only inputs use fallback",
			manual:    "   ",
			generated: "  \n  ",
			fallback:  "fallback text",
			want:      "fallback text\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeNotes(tt.manual, tt.generated, tt.fallback)
			if got != tt.want {
				t.Errorf("MergeNotes(%q, %q, %q)\ngot:  %q\nwant: %q",
					tt.manual, tt.generated, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestUpdateChangelog(t *testing.T) {
	tests := []struct {
		name      string
		changelog string
		version   string
		notes     string
		want      string
	}{
		{
			name: "standard case with existing next section",
			changelog: `# CHANGELOG

## next

<!-- Add manual release notes here. -->

## 0.0.59

- Old entry
`,
			version: "0.0.60",
			notes:   "- New feature A\n- New feature B",
			want: `# CHANGELOG

## next

<!-- Add manual release notes here. They will be merged into the generated changelog at release time. -->

## 0.0.60

- New feature A
- New feature B

## 0.0.59

- Old entry
`,
		},
		{
			name: "next section with content gets reset",
			changelog: `# CHANGELOG

## next

Some manual notes here.

## 0.0.59

- Old entry
`,
			version: "0.0.60",
			notes:   "- Release notes",
			want: `# CHANGELOG

## next

<!-- Add manual release notes here. They will be merged into the generated changelog at release time. -->

## 0.0.60

- Release notes

## 0.0.59

- Old entry
`,
		},
		{
			name: "no next section inserts before first version",
			changelog: `# CHANGELOG

## 0.0.59

- Old entry
`,
			version: "0.0.60",
			notes:   "- New feature",
			want: `# CHANGELOG

## next

<!-- Add manual release notes here. They will be merged into the generated changelog at release time. -->

## 0.0.60

- New feature

## 0.0.59

- Old entry
`,
		},
		{
			name: "preserves existing entries",
			changelog: `# CHANGELOG

## next

<!-- comment -->

## 0.0.59

- Feature A

## 0.0.58

- Feature B
`,
			version: "0.0.60",
			notes:   "- Feature C",
			want: `# CHANGELOG

## next

<!-- Add manual release notes here. They will be merged into the generated changelog at release time. -->

## 0.0.60

- Feature C

## 0.0.59

- Feature A

## 0.0.58

- Feature B
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UpdateChangelog(tt.changelog, tt.version, tt.notes)
			if got != tt.want {
				t.Errorf("UpdateChangelog(version=%q)\ngot:\n%s\nwant:\n%s", tt.version, got, tt.want)
			}
		})
	}
}
