package main

import (
	"strings"
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

func TestParseMarkerLine(t *testing.T) {
	// Valid lines.
	entry, ok := ParseMarkerLine("- [hash:abc12345][type:Features] Add new study command")
	if !ok || entry.hash != "abc12345" || entry.typ != "Features" || entry.message != "Add new study command" {
		t.Errorf("unexpected parse result: %+v, ok=%v", entry, ok)
	}
	entry, ok = ParseMarkerLine("- [hash:def67890][type:Bug Fixes] Fix workspace listing")
	if !ok || entry.hash != "def67890" || entry.typ != "Bug Fixes" || entry.message != "Fix workspace listing" {
		t.Errorf("unexpected parse result: %+v, ok=%v", entry, ok)
	}

	// Non-matching lines.
	for _, line := range []string{"- Added a feature", "", "### Features"} {
		if _, ok := ParseMarkerLine(line); ok {
			t.Errorf("ParseMarkerLine(%q) should not match", line)
		}
	}
}

func TestAreaForFiles(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  string
	}{
		{
			name:  "study command files",
			files: []string{"cmd/study/list.go", "cmd/study/list_test.go"},
			want:  "Study",
		},
		{
			name:  "aitaskbuilder files",
			files: []string{"cmd/aitaskbuilder/create.go"},
			want:  "AI Task Builder",
		},
		{
			name:  "mixed files uses majority area",
			files: []string{"cmd/study/list.go", "cmd/study/get.go", "cmd/workspace/list.go"},
			want:  "Study",
		},
		{
			name:  "model and client map to Core",
			files: []string{"model/study.go", "client/client.go"},
			want:  "Core",
		},
		{
			name:  "filters and filtersets both map to Filters",
			files: []string{"cmd/filters/list.go", "cmd/filtersets/list.go"},
			want:  "Filters",
		},
		{
			name:  "unknown files fall back to Other",
			files: []string{"README.md", "Makefile"},
			want:  "Other",
		},
		{
			name:  "empty file list returns Other",
			files: []string{},
			want:  "Other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AreaForFiles(tt.files)
			if got != tt.want {
				t.Errorf("AreaForFiles(%v) = %q, want %q", tt.files, got, tt.want)
			}
		})
	}
}

func TestTransformChangelog(t *testing.T) {
	noop := func(string) ([]string, error) { return nil, nil }

	t.Run("empty input", func(t *testing.T) {
		if got := TransformChangelog("", noop); got != "" {
			t.Errorf("expected empty output, got: %q", got)
		}
	})

	t.Run("groups by area with features before fixes", func(t *testing.T) {
		filesByHash := map[string][]string{
			"aaa11111": {"cmd/study/list.go", "cmd/study/list_test.go"},
			"bbb22222": {"cmd/study/get.go"},
			"ccc33333": {"cmd/workspace/list.go"},
			"ddd44444": {"cmd/aitaskbuilder/create.go"},
			"eee55555": {"README.md"},
		}
		input := strings.Join([]string{
			"- [hash:aaa11111][type:Features] Add study list pagination",
			"- [hash:bbb22222][type:Bug Fixes] Fix study get error handling",
			"- [hash:ccc33333][type:Features] Add workspace filtering",
			"- [hash:ddd44444][type:Features] Add AI task builder create command",
			"- [hash:eee55555][type:Documentation] Update README",
		}, "\n")

		got := TransformChangelog(input, func(h string) ([]string, error) { return filesByHash[h], nil })

		// Area headings present, markers stripped.
		for _, heading := range []string{"### AI Task Builder", "### Study", "### Workspaces", "### Other"} {
			if !strings.Contains(got, heading) {
				t.Errorf("missing heading %q", heading)
			}
		}
		if strings.Contains(got, "[hash:") || strings.Contains(got, "[type:") {
			t.Error("markers should be stripped from output")
		}

		// AI Task Builder before Study (per areaOrder).
		if strings.Index(got, "### AI Task Builder") > strings.Index(got, "### Study") {
			t.Error("expected AI Task Builder before Study")
		}

		// Features before fixes within Study.
		studySection := got[strings.Index(got, "### Study"):]
		if strings.Index(studySection, "Add study list pagination") > strings.Index(studySection, "Fix study get error handling") {
			t.Error("expected features before fixes within Study")
		}
	})

	t.Run("fixes listed after features in same area", func(t *testing.T) {
		input := strings.Join([]string{
			"- [hash:fff11111][type:Bug Fixes] Fix pagination bug",
			"- [hash:fff22222][type:Features] Add sorting support",
		}, "\n")
		allStudy := func(string) ([]string, error) { return []string{"cmd/study/list.go"}, nil }

		got := TransformChangelog(input, allStudy)
		if strings.Index(got, "Add sorting support") > strings.Index(got, "Fix pagination bug") {
			t.Error("expected feature before bug fix")
		}
	})
}
