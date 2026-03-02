// Package main provides a CLI tool for changelog manipulation used in the
// release workflow. It replaces inline shell/Python scripts with testable Go.
//
// Usage:
//
//	go run ./scripts/changelog extract --section next --strip-comments
//	go run ./scripts/changelog merge --manual MANUAL.md --generated CLIFF.md --output MERGED.md
//	go run ./scripts/changelog update --version 0.1.0 --notes NOTES.md
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	reHTMLComment    = regexp.MustCompile(`(?s)<!--.*?-->`)
	reBlankLines     = regexp.MustCompile(`\n{3,}`)
	reVersionHeading = regexp.MustCompile(`^## \d`)
	reChangelogTitle = regexp.MustCompile(`(?m)(# CHANGELOG\n+)`)
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: changelog <extract|merge|update|transform> [flags]")
		os.Exit(1)
	}

	cmd := os.Args[1]
	os.Args = append(os.Args[:1], os.Args[2:]...) // shift for flag parsing

	switch cmd {
	case "extract":
		runExtract()
	case "merge":
		runMerge()
	case "update":
		runUpdate()
	case "transform":
		runTransform()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

// resolvePath resolves a relative file path against the working directory and
// verifies it does not escape via traversal (CWE-23). The resolved absolute
// path is returned.
func resolvePath(raw string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}
	abs := filepath.Join(wd, filepath.Clean(raw))
	if !strings.HasPrefix(abs, wd+string(filepath.Separator)) && abs != wd {
		return "", fmt.Errorf("path %q is outside the working directory", raw)
	}
	return abs, nil
}

func runExtract() {
	fs := flag.NewFlagSet("extract", flag.ContinueOnError)
	section := fs.String("section", "", "section heading to extract (e.g. next, 0.0.60)")
	stripComments := fs.Bool("strip-comments", false, "remove HTML comments and resulting blank lines")
	changelog := fs.String("changelog", "CHANGELOG.md", "path to changelog file")
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "extract: %v\n", err)
		os.Exit(1)
	}

	if *section == "" {
		fmt.Fprintln(os.Stderr, "extract: --section is required")
		os.Exit(1)
	}

	path, err := resolvePath(*changelog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "extract: %v\n", err)
		os.Exit(1)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "extract: %v\n", err)
		os.Exit(1)
	}

	result := ExtractSection(string(data), *section, *stripComments)
	fmt.Print(result)
}

func runMerge() {
	fs := flag.NewFlagSet("merge", flag.ContinueOnError)
	manual := fs.String("manual", "", "path to manual notes file")
	generated := fs.String("generated", "", "path to generated notes file")
	fallback := fs.String("fallback", "- Maintenance and dependency updates", "fallback text if both inputs are empty")
	output := fs.String("output", "", "output file path (required)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "merge: %v\n", err)
		os.Exit(1)
	}

	if *output == "" {
		fmt.Fprintln(os.Stderr, "merge: --output is required")
		os.Exit(1)
	}

	manualText := readFileOrEmpty(*manual)
	generatedText := readFileOrEmpty(*generated)

	result := MergeNotes(manualText, generatedText, *fallback)

	outPath, err := resolvePath(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "merge: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(outPath, []byte(result), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "merge: %v\n", err)
		os.Exit(1)
	}
}

func runUpdate() {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	version := fs.String("version", "", "new version number (e.g. 0.1.0)")
	notes := fs.String("notes", "", "path to release notes file")
	changelog := fs.String("changelog", "CHANGELOG.md", "path to changelog file")
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "update: %v\n", err)
		os.Exit(1)
	}

	if *version == "" {
		fmt.Fprintln(os.Stderr, "update: --version is required")
		os.Exit(1)
	}
	if *notes == "" {
		fmt.Fprintln(os.Stderr, "update: --notes is required")
		os.Exit(1)
	}

	changelogPath, err := resolvePath(*changelog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update: %v\n", err)
		os.Exit(1)
	}
	changelogData, err := os.ReadFile(changelogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update: reading changelog: %v\n", err)
		os.Exit(1)
	}

	notesPath, err := resolvePath(*notes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update: %v\n", err)
		os.Exit(1)
	}
	notesData, err := os.ReadFile(notesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update: reading notes: %v\n", err)
		os.Exit(1)
	}

	result := UpdateChangelog(string(changelogData), *version, strings.TrimSpace(string(notesData)))

	if err := os.WriteFile(changelogPath, []byte(result), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "update: writing changelog: %v\n", err)
		os.Exit(1)
	}
}

// readFileOrEmpty reads a file and returns its trimmed content, or empty string
// if the path is empty or the file doesn't exist.
func readFileOrEmpty(raw string) string {
	if raw == "" {
		return ""
	}
	path, err := resolvePath(raw)
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// ExtractSection extracts the content between a ## <section> heading and the
// next ## heading. If stripComments is true, HTML comments and resulting blank
// lines are removed. The section is matched literally (not as a regex).
func ExtractSection(changelog, section string, stripComments bool) string {
	lines := strings.Split(changelog, "\n")
	heading := "## " + section

	var out []string
	found := false
	for _, line := range lines {
		if found {
			if strings.HasPrefix(line, "## ") {
				break
			}
			out = append(out, line)
		} else if strings.TrimSpace(line) == heading {
			found = true
		}
	}

	if !found {
		return ""
	}

	result := strings.Join(out, "\n")

	if stripComments {
		result = reHTMLComment.ReplaceAllString(result, "")
		result = reBlankLines.ReplaceAllString(result, "\n\n")
	}

	result = strings.TrimSpace(result)
	if result != "" {
		result += "\n"
	}
	return result
}

// MergeNotes combines manual and generated notes with a blank line separator.
// If both are empty, the fallback text is returned.
func MergeNotes(manual, generated, fallback string) string {
	manual = strings.TrimSpace(manual)
	generated = strings.TrimSpace(generated)

	var parts []string
	if manual != "" {
		parts = append(parts, manual)
	}
	if generated != "" {
		parts = append(parts, generated)
	}

	if len(parts) == 0 {
		return fallback + "\n"
	}
	return strings.Join(parts, "\n\n") + "\n"
}

// reMarker matches the [hash:...][type:...] prefix emitted by cliff.toml.
var reMarker = regexp.MustCompile(`^\- \[hash:([a-f0-9]+)\]\[type:([^\]]+)\] (.+)$`)

// areaMapping maps file path prefixes to user-facing area names.
// Files not matching any prefix are ignored — commits that only touch
// unrecognised paths are excluded from the changelog.
var areaMapping = []struct {
	prefix string
	area   string
}{
	{"cmd/aitaskbuilder/", "AI Task Builder"},
	{"cmd/study/", "Study"},
	{"cmd/bonus/", "Bonus Payments"},
	{"cmd/message/", "Messaging"},
	{"cmd/collection/", "Collections"},
	{"cmd/credentials/", "Credentials"},
	{"cmd/campaign/", "Campaigns"},
	{"cmd/workspace/", "Workspaces"},
	{"cmd/project/", "Projects"},
	{"cmd/submission/", "Submissions"},
	{"cmd/participantgroup/", "Participant Groups"},
	{"cmd/hook/", "Hooks"},
	{"cmd/filters/", "Filters"},
	{"cmd/filtersets/", "Filters"},
	{"cmd/requirements/", "Requirements"},
	{"cmd/user/", "User"},
	{"model/", "Core"},
	{"client/", "Core"},
}

// parsedEntry represents a single changelog entry parsed from cliff output.
type parsedEntry struct {
	hash    string
	typ     string // "Features", "Bug Fixes", etc.
	message string
}

// ParseMarkerLine parses a cliff output line with [hash:...][type:...] markers.
// Returns the parsed entry and true if the line matched, or zero value and false.
func ParseMarkerLine(line string) (parsedEntry, bool) {
	m := reMarker.FindStringSubmatch(strings.TrimSpace(line))
	if m == nil {
		return parsedEntry{}, false
	}
	return parsedEntry{hash: m[1], typ: m[2], message: m[3]}, true
}

// AreaForFiles determines the primary area for a set of changed file paths.
// Only files matching a prefix in areaMapping are counted. Returns "" when
// no files match any known area.
func AreaForFiles(files []string) string {
	counts := make(map[string]int)
	for _, f := range files {
		for _, am := range areaMapping {
			if strings.HasPrefix(f, am.prefix) {
				counts[am.area]++
				break
			}
		}
	}
	if len(counts) == 0 {
		return ""
	}
	best := ""
	bestCount := 0
	// Sort keys for deterministic tie-breaking.
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if counts[k] > bestCount {
			best = k
			bestCount = counts[k]
		}
	}
	return best
}

// gitDiffTreeFunc is the function used to get changed files for a commit hash.
// It is a variable so tests can override it.
var gitDiffTreeFunc = gitDiffTree

// gitDiffTree runs git diff-tree to get the list of files changed by a commit.
func gitDiffTree(hash string) ([]string, error) {
	out, err := exec.CommandContext(context.Background(), "git", "diff-tree", "--no-commit-id", "--name-only", "-r", hash).Output()
	if err != nil {
		return nil, fmt.Errorf("git diff-tree %s: %w", hash, err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// areaOrder defines the display order of areas in the transformed output.
// Areas not in this list appear alphabetically after these.
var areaOrder = []string{
	"AI Task Builder",
	"Study",
	"Bonus Payments",
	"Messaging",
	"Collections",
	"Credentials",
	"Campaigns",
	"Workspaces",
	"Projects",
	"Submissions",
	"Participant Groups",
	"Hooks",
	"Filters",
	"Requirements",
	"User",
	"Core",
}

// TransformChangelog takes cliff-generated markdown with [hash:...][type:...]
// markers, resolves each commit's changed files, and re-groups entries by
// subcommand area with features before fixes within each area.
func TransformChangelog(input string, diffTreeFn func(string) ([]string, error)) string {
	type entry struct {
		typ     string
		message string
	}
	grouped := make(map[string][]entry)

	for _, line := range strings.Split(input, "\n") {
		parsed, ok := ParseMarkerLine(line)
		if !ok {
			continue
		}

		files, err := diffTreeFn(parsed.hash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not resolve files for %s: %v\n", parsed.hash, err)
		}
		area := AreaForFiles(files)
		if area == "" {
			continue
		}

		grouped[area] = append(grouped[area], entry{typ: parsed.typ, message: parsed.message})
	}

	if len(grouped) == 0 {
		return ""
	}

	// Build ordered list of areas.
	orderIndex := make(map[string]int)
	for i, a := range areaOrder {
		orderIndex[a] = i
	}
	areas := make([]string, 0, len(grouped))
	for a := range grouped {
		areas = append(areas, a)
	}
	sort.Slice(areas, func(i, j int) bool {
		ii, okI := orderIndex[areas[i]]
		ij, okJ := orderIndex[areas[j]]
		if okI && okJ {
			return ii < ij
		}
		if okI {
			return true
		}
		if okJ {
			return false
		}
		return areas[i] < areas[j]
	})

	var buf strings.Builder
	for i, area := range areas {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString("### " + area + "\n\n")

		entries := grouped[area]
		// Sort: features first, then fixes, then everything else.
		sort.SliceStable(entries, func(a, b int) bool {
			return typePriority(entries[a].typ) < typePriority(entries[b].typ)
		})

		for _, e := range entries {
			buf.WriteString("- " + e.message + "\n")
		}
	}

	return buf.String()
}

// typePriority returns a sort key so features come before fixes, and
// everything else comes after.
func typePriority(typ string) int {
	switch strings.ToLower(typ) {
	case "features":
		return 0
	case "bug fixes":
		return 1
	default:
		return 2
	}
}

func runTransform() {
	fs := flag.NewFlagSet("transform", flag.ContinueOnError)
	input := fs.String("input", "", "path to cliff-generated notes file (required)")
	output := fs.String("output", "", "output file path (required)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "transform: %v\n", err)
		os.Exit(1)
	}

	if *input == "" {
		fmt.Fprintln(os.Stderr, "transform: --input is required")
		os.Exit(1)
	}
	if *output == "" {
		fmt.Fprintln(os.Stderr, "transform: --output is required")
		os.Exit(1)
	}

	inputPath, err := resolvePath(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "transform: %v\n", err)
		os.Exit(1)
	}
	data, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "transform: reading input: %v\n", err)
		os.Exit(1)
	}

	result := TransformChangelog(string(data), gitDiffTreeFunc)

	outPath, err := resolvePath(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "transform: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(outPath, []byte(result), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "transform: writing output: %v\n", err)
		os.Exit(1)
	}
}

// UpdateChangelog inserts a new version entry into the changelog. It resets the
// ## next section and inserts the new ## X.Y.Z entry immediately after it. If
// there is no ## next section, the new entry is inserted before the first
// version heading.
func UpdateChangelog(changelog, version, notes string) string {
	nextSection := "## next\n\n<!-- Add manual release notes here. They will be merged into the generated changelog at release time. -->"
	newEntry := fmt.Sprintf("## %s\n\n%s", version, notes)

	// Try to replace the existing ## next block.
	// Go's regexp doesn't support lookaheads, so we use a line-based approach.
	lines := strings.Split(changelog, "\n")
	nextIdx := -1
	endIdx := -1
	for i, line := range lines {
		if nextIdx < 0 && strings.TrimSpace(line) == "## next" {
			nextIdx = i
		} else if nextIdx >= 0 && endIdx < 0 && reVersionHeading.MatchString(line) {
			endIdx = i
		}
	}

	if nextIdx >= 0 && endIdx >= 0 {
		var result []string
		result = append(result, lines[:nextIdx]...)
		result = append(result, strings.Split(nextSection, "\n")...)
		result = append(result, "")
		result = append(result, strings.Split(newEntry, "\n")...)
		result = append(result, "")
		result = append(result, lines[endIdx:]...)
		return strings.Join(result, "\n")
	}

	// If no ## next section, insert before the first version heading.
	if reChangelogTitle.MatchString(changelog) {
		replacement := "${1}" + nextSection + "\n\n" + newEntry + "\n\n"
		return reChangelogTitle.ReplaceAllString(changelog, replacement)
	}

	// Last resort: prepend.
	return nextSection + "\n\n" + newEntry + "\n\n" + changelog
}
