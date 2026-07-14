package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const defaultWorkspaceID = "679271425fe00981084a5f58" // DCT Workspace

var fieldPattern = regexp.MustCompile(`(?m)^([A-Za-z ]+):\s*(\S+)$`)

func main() {
	repoRoot, err := repoRoot()
	if err != nil {
		fatal(err)
	}

	workspaceID := defaultWorkspaceID
	if len(os.Args) > 1 {
		workspaceID = os.Args[1]
	}

	cliBinary := filepath.Join(os.TempDir(), "prolific-cli")
	if err := buildCLI(repoRoot, cliBinary); err != nil {
		fatal(err)
	}

	schema := `{"fields":{"question":{"type":"text","label":"Question"},"clip":{"type":"audio_url","label":"Audio clip"}}}`
	batchItems := `[{"rows":[{"columns":[{"items":[{"type":"dataset_field","field":"question"},{"type":"dataset_field","field":"clip"},{"type":"free_text","description":"Please describe what you heard."}]}]}]}]`

	// Create a V4 dataset whose schema includes an audio_url field.
	output, err := run(repoRoot, cliBinary,
		"aitaskbuilder", "dataset", "create",
		"-n", "Audio URL Test Dataset",
		"-w", workspaceID,
		"--strict",
		"--schema", schema,
	)
	if err != nil {
		fatal(err)
	}

	datasetID, err := extractField(output, "ID")
	if err != nil {
		fatal(err)
	}
	fmt.Printf("Created dataset: %s\n", datasetID)

	csvPath := filepath.Join(os.TempDir(), "audio-dataset.csv")
	csvContents := strings.Join([]string{
		"question,clip",
		`"What emotion is being expressed?","https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3"`,
		`"Transcribe the spoken word.","https://www.w3schools.com/html/horse.mp3"`,
		"",
	}, "\n")
	if err := os.WriteFile(csvPath, []byte(csvContents), 0o600); err != nil {
		fatal(fmt.Errorf("failed to write CSV fixture: %w", err))
	}

	invalidCSVPath := filepath.Join(os.TempDir(), "audio-dataset-invalid.csv")
	invalidCSVContents := strings.Join([]string{
		"question,clip",
		`"This row should fail validation.","https://example.com/not-audio.txt"`,
		"",
	}, "\n")
	if err := os.WriteFile(invalidCSVPath, []byte(invalidCSVContents), 0o600); err != nil {
		fatal(fmt.Errorf("failed to write invalid CSV fixture: %w", err))
	}

	// Attempt an invalid upload first to confirm audio_url extension validation rejects non-audio URLs.
	if _, err := runAllowFailure(repoRoot, cliBinary, "aitaskbuilder", "dataset", "upload", "-d", datasetID, "-f", invalidCSVPath); err == nil {
		fatal(fmt.Errorf("expected invalid audio URL upload to fail"))
	} else {
		fmt.Println("Confirmed invalid audio URL upload was rejected.")
	}

	// Upload valid sample data containing supported audio URL extensions.
	if _, err := run(repoRoot, cliBinary, "aitaskbuilder", "dataset", "upload", "-d", datasetID, "-f", csvPath); err != nil {
		fatal(err)
	}

	// Check dataset status for extra visibility while running the manual flow.
	_, _ = runAllowFailure(repoRoot, cliBinary, "aitaskbuilder", "dataset", "check", "-d", datasetID)

	// Create a batch linked to the dataset and define the participant layout via batch_items.
	// batch_items replaces the deprecated standalone instructions endpoint, and this
	// manual flow intentionally references the audio_url dataset field even though
	// the checked-in OpenAPI file has not caught up with the backend yet.
	output, err = run(repoRoot, cliBinary,
		"aitaskbuilder", "batch", "create",
		"-n", "Audio URL Test Batch",
		"-w", workspaceID,
		"-d", datasetID,
		"--task-name", "Audio Review Task",
		"--task-introduction", "Listen to the audio clip and answer the question.",
		"--task-steps", "1. Listen to the clip\\n2. Answer the question",
		"--batch-items-json", batchItems,
	)
	if err != nil {
		fatal(err)
	}

	batchID, err := extractField(output, "ID")
	if err != nil {
		fatal(err)
	}
	fmt.Printf("Created batch: %s\n", batchID)

	// Set up the batch so the persisted preview route has a task group to open.
	if _, err := run(repoRoot, cliBinary,
		"aitaskbuilder", "batch", "setup",
		"-b", batchID,
		"-d", datasetID,
		"--tasks-per-group", "1",
	); err != nil {
		fatal(err)
	}

	// Preview the batch through the researcher preview URL flow.
	if _, err := run(repoRoot, cliBinary, "aitaskbuilder", "batch", "preview", batchID); err != nil {
		fatal(err)
	}

	fmt.Println("\nDone. Review output above for correctness.")
}

func repoRoot() (string, error) {
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to resolve script path")
	}

	return filepath.Abs(filepath.Join(filepath.Dir(filePath), "..", ".."))
}

func buildCLI(repoRoot, cliBinary string) error {
	fmt.Println("Building CLI...")
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", cliBinary, ".")
	_, err := runWithCheck(true, repoRoot, "go", cmd)
	return err
}

func run(repoRoot, cliBinary string, args ...string) (string, error) {
	// #nosec G702 -- cliBinary is the local binary built by this script, not shell-expanded user input.
	cmd := exec.CommandContext(context.Background(), cliBinary, args...)
	return runWithCheck(true, repoRoot, cliBinary, cmd)
}

func runAllowFailure(repoRoot, cliBinary string, args ...string) (string, error) {
	// #nosec G702 -- cliBinary is the local binary built by this script, not shell-expanded user input.
	cmd := exec.CommandContext(context.Background(), cliBinary, args...)
	return runWithCheck(false, repoRoot, cliBinary, cmd)
}

func runWithCheck(check bool, repoRoot, command string, cmd *exec.Cmd) (string, error) {
	fmt.Printf("\n$ %s %s\n", command, strings.Join(cmd.Args[1:], " "))

	cmd.Dir = repoRoot
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	text := string(output)
	fmt.Print(text)

	if err != nil && check {
		return text, fmt.Errorf("command failed: %w", err)
	}

	return text, err
}

func extractField(output, label string) (string, error) {
	matches := fieldPattern.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		if len(match) >= 3 && match[1] == label {
			return match[2], nil
		}
	}

	return "", fmt.Errorf("could not find %q in output", label)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
