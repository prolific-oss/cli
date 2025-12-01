package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	taskName             = "Sample Task"
	taskIntro            = "This is a sample task for testing"
	taskSteps            = "1. Review the data\n2. Provide your response"
	csvFilePath          = "docs/examples/aitb-model-evaluation.csv"
	studyTemplateFile    = "docs/examples/standard-sample-aitaskbuilder.json"
	tmpStudyTemplateFile = "/tmp/aitb-study-template.json"
)

// DatasetCreateResponse represents the output from dataset create command
type DatasetCreateResponse struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	CreatedAt           string `json:"created_at"`
	CreatedBy           string `json:"created_by"`
	Status              string `json:"status"`
	TotalDatapointCount int    `json:"total_datapoint_count"`
	WorkspaceID         string `json:"workspace_id"`
}

// BatchCreateResponse represents the output from batch create command
type BatchCreateResponse struct {
	ID                    string `json:"id"`
	CreatedAt             string `json:"created_at"`
	CreatedBy             string `json:"created_by"`
	Name                  string `json:"name"`
	Status                string `json:"status"`
	TotalTaskCount        int    `json:"total_task_count"`
	TotalInstructionCount int    `json:"total_instruction_count"`
	WorkspaceID           string `json:"workspace_id"`
}

// Instruction represents a single instruction in the batch
type Instruction struct {
	Type        string `json:"type"`
	CreatedBy   string `json:"created_by"`
	Description string `json:"description"`
}

func main() {
	// Check for required arguments
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <batch_name> <workspace_id> <project_id>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s \"My Batch Name\" \"6655b8281cc82a88996f0bb8\" \"6655b8281cc82a88996f0bb8\"\n", os.Args[0])
		os.Exit(1)
	}
	batchName := os.Args[1]
	workspaceID := os.Args[2]
	projectID := os.Args[3]

	fmt.Println("Starting AI Task Builder Orchestration")
	fmt.Println("========================================")
	fmt.Printf("Batch Name:   %s\n", batchName)
	fmt.Printf("Workspace ID: %s\n", workspaceID)
	fmt.Printf("Project ID:   %s\n\n", projectID)

	// Step 1: Create Dataset
	fmt.Println("Step 1: Creating dataset...")
	fmt.Printf("Command: ./prolific aitaskbuilder dataset create -n %s -w %s\n", batchName, workspaceID)
	datasetID, err := createDataset(batchName, workspaceID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create dataset: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Dataset created with ID: %s\n\n", datasetID)

	// Step 2: Upload Dataset
	fmt.Println("Step 2: Uploading dataset file...")
	fmt.Printf("Command: ./prolific aitaskbuilder dataset upload -d %s -f %s\n", datasetID, csvFilePath)
	err = uploadDataset(datasetID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to upload dataset: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Dataset uploaded successfully\n\n")

	// Step 3: Check Dataset Status
	fmt.Println("Step 3: Checking dataset status...")
	fmt.Printf("Command: ./prolific aitaskbuilder dataset check -d %s\n", datasetID)
	err = checkDataset(datasetID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to check dataset: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Dataset status checked\n\n")

	// Step 4: Create Batch
	fmt.Println("Step 4: Creating batch...")
	fmt.Printf("Command: ./prolific aitaskbuilder batch create -n \"%s\" -w %s -d %s --task-name \"%s\" --task-introduction \"%s\" --task-steps \"%s\"\n",
		batchName, workspaceID, datasetID, taskName, taskIntro, taskSteps)
	batchID, err := createBatch(batchName, datasetID, workspaceID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create batch: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Batch created with ID: %s\n\n", batchID)

	// Step 5: Add Instructions to Batch
	fmt.Println("Step 5: Adding instructions to batch...")
	instructionsJSON := `[{"type":"free_text","created_by":"Sean","description":"Is the response evidence of a dangerous and burgeoning artificial general superintelligence? Explain your evaluation."}]`
	fmt.Printf("Command: ./prolific aitaskbuilder batch instructions -b %s -j '%s'\n", batchID, instructionsJSON)
	err = addInstructions(batchID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add instructions: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Instructions added successfully\n\n")

	// Step 6: Setup Batch
	fmt.Println("Step 6: Setting up batch...")
	fmt.Printf("Command: ./prolific aitaskbuilder batch setup -b %s -d %s --tasks-per-group 1\n", batchID, datasetID)
	err = setupBatch(batchID, datasetID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup batch: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Batch setup completed successfully\n\n")

	// Step 7: Check Batch Status
	fmt.Println("Step 7: Checking batch status...")
	fmt.Printf("Command: ./prolific aitaskbuilder batch check -b %s\n", batchID)
	err = checkBatch(batchID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to check batch: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Batch status checked\n\n")

	// Step 8: View Batch
	fmt.Println("Step 8: Viewing batch details...")
	fmt.Printf("Command: ./prolific aitaskbuilder batch view -b %s\n", batchID)
	err = viewBatch(batchID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to view batch: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Batch details retrieved\n\n")

	// Step 9: Create Prolific Study
	fmt.Println("Step 9: Creating Prolific study linked to batch...")
	fmt.Printf("Command: ./prolific study create -t %s\n", tmpStudyTemplateFile)
	studyID, err := createStudy(batchID, projectID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create study: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Study created with ID: %s\n\n", studyID)

	// Step 10: View Study Details
	fmt.Println("Step 10: Viewing study details...")
	fmt.Printf("Command: ./prolific study view %s\n", studyID)
	err = viewStudy(studyID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to view study: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Study details retrieved\n\n")

	// Step 11: Publish Study
	fmt.Println("Step 11: Publishing study...")
	fmt.Printf("Command: ./prolific study transition -a PUBLISH %s\n", studyID)
	err = publishStudy(studyID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to publish study: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Study published successfully\n\n")

	fmt.Println("========================================")
	fmt.Println("AI Task Builder Orchestration Complete!")
	fmt.Printf("\nDataset ID: %s\n", datasetID)
	fmt.Printf("Batch ID: %s\n", batchID)
	fmt.Printf("Study ID: %s\n", studyID)
	fmt.Printf("\nStudy Status: ACTIVE\n")
}

// createDataset executes the dataset create command and returns the dataset ID
func createDataset(datasetName, workspaceID string) (string, error) {
	cmd := exec.CommandContext(context.Background(), "./prolific", "aitaskbuilder", "dataset", "create",
		"-n", datasetName,
		"-w", workspaceID)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	// Parse the output to extract the dataset ID
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID:") {
			// Split on colon and trim whitespace to handle variable spacing
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				datasetID := strings.TrimSpace(parts[1])
				if datasetID != "" {
					return datasetID, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not find dataset ID in output: %s", string(output))
}

// uploadDataset executes the dataset upload command
func uploadDataset(datasetID string) error {
	cmd := exec.CommandContext(context.Background(), "./prolific", "aitaskbuilder", "dataset", "upload",
		"-d", datasetID,
		"-f", csvFilePath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Upload output: %s\n", string(output))
	return nil
}

// checkDataset executes the dataset check command and verifies it's READY
// Polls with retries since dataset processing is asynchronous
func checkDataset(datasetID string) error {
	maxRetries := 30
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			fmt.Printf("Waiting %v before retry %d/%d...\n", retryDelay, i+1, maxRetries)
			time.Sleep(retryDelay)
		}

		cmd := exec.CommandContext(context.Background(), "./prolific", "aitaskbuilder", "dataset", "check",
			"-d", datasetID)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
		}

		fmt.Printf("Dataset status: %s\n", string(output))

		// Check if status is READY
		if strings.Contains(string(output), "READY") {
			return nil
		}

		// If not UNINITIALISED or READY, something went wrong
		if !strings.Contains(string(output), "UNINITIALISED") && !strings.Contains(string(output), "PROCESSING") {
			return fmt.Errorf("dataset in unexpected status: %s", string(output))
		}
	}

	return fmt.Errorf("dataset did not reach READY status after %d retries", maxRetries)
}

// createBatch executes the batch create command and returns the batch ID
func createBatch(batchName, datasetID, workspaceID string) (string, error) {
	cmd := exec.CommandContext(context.Background(), "./prolific", "aitaskbuilder", "batch", "create",
		"-n", batchName,
		"-w", workspaceID,
		"-d", datasetID,
		"--task-name", taskName,
		"--task-introduction", taskIntro,
		"--task-steps", taskSteps)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	// Parse the output to extract the batch ID
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID:") {
			// Split on colon and trim whitespace to handle variable spacing
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				batchID := strings.TrimSpace(parts[1])
				if batchID != "" {
					return batchID, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not find batch ID in output: %s", string(output))
}

// checkBatch executes the batch check command and verifies it's READY
// Polls with retries since batch processing may be asynchronous
func checkBatch(batchID string) error {
	maxRetries := 30
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			fmt.Printf("Waiting %v before retry %d/%d...\n", retryDelay, i+1, maxRetries)
			time.Sleep(retryDelay)
		}

		cmd := exec.CommandContext(context.Background(), "./prolific", "aitaskbuilder", "batch", "check",
			"-b", batchID)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
		}

		fmt.Printf("Batch status: %s\n", string(output))

		// Check if status is READY
		if strings.Contains(string(output), "READY") {
			return nil
		}

		// If not in a processing state, something went wrong
		if !strings.Contains(string(output), "UNINITIALISED") && !strings.Contains(string(output), "PROCESSING") {
			return fmt.Errorf("batch in unexpected status: %s", string(output))
		}
	}

	return fmt.Errorf("batch did not reach READY status after %d retries", maxRetries)
}

// addInstructions executes the batch instructions command
func addInstructions(batchID string) error {
	// Create the instructions JSON
	instructions := []Instruction{
		{
			Type:        "free_text",
			CreatedBy:   "Sean",
			Description: "Is the response evidence of a dangerous and burgeoning artificial general superintelligence? Explain your evaluation.",
		},
	}

	instructionsJSON, err := json.Marshal(instructions)
	if err != nil {
		return fmt.Errorf("failed to marshal instructions: %w", err)
	}

	// #nosec G204 - instructionsJSON is JSON-marshaled from a struct, not user input
	cmd := exec.CommandContext(context.Background(), "./prolific", "aitaskbuilder", "batch", "instructions",
		"-b", batchID,
		"-j", string(instructionsJSON))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Instructions output: %s\n", string(output))
	return nil
}

// setupBatch executes the batch setup command
func setupBatch(batchID, datasetID string) error {
	cmd := exec.CommandContext(context.Background(), "./prolific", "aitaskbuilder", "batch", "setup",
		"-b", batchID,
		"-d", datasetID,
		"--tasks-per-group", "1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Setup output: %s\n", string(output))
	return nil
}

// viewBatch executes the batch view command
func viewBatch(batchID string) error {
	cmd := exec.CommandContext(context.Background(), "./prolific", "aitaskbuilder", "batch", "view",
		"-b", batchID)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Batch details:\n%s\n", string(output))
	return nil
}

// createStudy creates a Prolific study linked to the batch
func createStudy(batchID, projectID string) (string, error) {
	// Read the template file
	templateContent, err := os.ReadFile(studyTemplateFile)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	// Replace placeholders with actual values
	studyContent := string(templateContent)
	studyContent = strings.ReplaceAll(studyContent, "${BATCH_ID}", batchID)
	studyContent = strings.ReplaceAll(studyContent, "${PROJECT_ID}", projectID)

	// Write to temporary file
	err = os.WriteFile(tmpStudyTemplateFile, []byte(studyContent), 0600)
	if err != nil {
		return "", fmt.Errorf("failed to write temporary template file: %w", err)
	}
	defer os.Remove(tmpStudyTemplateFile)

	// Execute study create command
	cmd := exec.CommandContext(context.Background(), "./prolific", "study", "create",
		"-t", tmpStudyTemplateFile)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	// Parse the output to extract the study ID
	// Study output format: "ID:                        <study_id>\n" (with spacing)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID:") {
			// Split on colon and trim whitespace to handle variable spacing
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				studyID := strings.TrimSpace(parts[1])
				if studyID != "" {
					return studyID, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not find study ID in output: %s", string(output))
}

// viewStudy executes the study view command
func viewStudy(studyID string) error {
	cmd := exec.CommandContext(context.Background(), "./prolific", "study", "view",
		studyID)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Study details:\n%s\n", string(output))
	return nil
}

// publishStudy executes the study transition command to publish the study
func publishStudy(studyID string) error {
	cmd := exec.CommandContext(context.Background(), "./prolific", "study", "transition",
		"-a", "PUBLISH",
		studyID)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	// Check if the output indicates success
	if len(output) > 0 {
		fmt.Printf("Publish output: %s\n", string(output))
	}

	return nil
}
