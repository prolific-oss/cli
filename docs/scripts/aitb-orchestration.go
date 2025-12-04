package main

// AITB Orchestration Demo Script
//
// This script demonstrates the complete AI Task Builder workflow from dataset
// creation through study publishing, with optional real-time response monitoring.
//
// USAGE:
//
// 1. Full orchestration demo:
//    go run aitb-orchestration.go "Batch Name" "workspace_id" "project_id"
//
// 2. Full orchestration demo with response tailing:
//    go run aitb-orchestration.go "Batch Name" "workspace_id" "project_id" --tail
//
// 3. Tail-only mode (for split-screen demos):
//    go run aitb-orchestration.go --tail-only "batch_id" "Task-A" "green"
//
// SPLIT-SCREEN DEMO:
// For a split-screen demo showing two task types:
// - Run the full orchestration twice to create two batches
// - In terminal 1: go run aitb-orchestration.go --tail-only "batch_id_1" "Task-A" "green"
// - In terminal 2: go run aitb-orchestration.go --tail-only "batch_id_2" "Task-B" "blue"
// - Arrange terminals side-by-side to show responses streaming in real-time

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// ANSI color codes and formatting
const (
	colorReset        = "\033[0m"
	colorBold         = "\033[1m"
	colorRed          = "\033[31m"
	colorGreen        = "\033[32m"
	colorYellow       = "\033[33m"
	colorBlue         = "\033[34m"
	colorPurple       = "\033[35m"
	colorCyan         = "\033[36m"
	colorWhite        = "\033[37m"
	colorGray         = "\033[90m"
	colorProlificBlue = "\033[38;2;15;43;201m" // Prolific brand blue #0F2BC9

	// Symbols
	symbolSuccess = "✓"
	symbolInfo    = "ℹ"
	symbolArrow   = "→"
	symbolDot     = "•"
)

const (
	taskName             = "Model Response Evaluation"
	taskIntro            = "You will be shown a prompt and two different AI model responses. Your task is to evaluate which response is better and explain your reasoning."
	taskSteps            = "1. Read the prompt carefully\n2. Review both Response 1 and Response 2\n3. Select which response is better\n4. Answer the accuracy questions for each response\n5. Provide detailed justification for your evaluation"
	csvFilePath          = "docs/examples/pairwise-evaluation.csv"
	instructionsFile     = "docs/examples/pairwise-evaluation-instructions.json"
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

// printSectionHeader prints a formatted section header with Enter prompt
func printSectionHeader(stepNum int, title string) {
	// Prompt to continue
	fmt.Printf("\n%s%s▶ Press Enter to continue to Step %d...%s ", colorBold, colorCyan, stepNum, colorReset)
	_, _ = fmt.Scanln() // Wait for Enter key

	fmt.Printf("\n%s%s%s Step %d: %s %s\n", colorBold, colorBlue, symbolArrow, stepNum, title, colorReset)
	fmt.Printf("%s%s%s\n", colorGray, strings.Repeat("─", 60), colorReset)
}

// printCommand prints a command in a formatted way
func printCommand(cmd string) {
	fmt.Printf("%s%s Command:%s %s%s\n", colorGray, symbolDot, colorReset, colorCyan, cmd)
	fmt.Print(colorReset)
}

// printSuccess prints a success message
func printSuccess(message string) {
	fmt.Printf("%s%s %s%s\n", colorGreen, symbolSuccess, message, colorReset)
}

// printInfo prints an info message
func printInfo(label, value string) {
	fmt.Printf("%s%s%s %s:%s %s%s%s\n", colorBold, colorWhite, symbolInfo, label, colorReset, colorCyan, value, colorReset)
}

// showSpinner displays an animated spinner for a duration
func showSpinner(duration time.Duration, message string) {
	spinChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(duration)
	i := 0

	for {
		select {
		case <-timeout:
			fmt.Printf("\r%s\n", strings.Repeat(" ", 80)) // Clear the line
			return
		case <-ticker.C:
			fmt.Printf("\r%s%s%s %s %s", colorYellow, spinChars[i%len(spinChars)], colorReset, message, colorGray)
			i++
		}
	}
}

// printDatasetPreview displays the CSV data in a formatted table
func printDatasetPreview(filePath string) error {
	// Read the CSV file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read CSV file: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	fmt.Printf("\n%s%s📋 Dataset Preview:%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s%s%s\n", colorGray, strings.Repeat("─", 80), colorReset)

	// Parse and display rows (limit to first 3 data rows + header for demo)
	maxRows := 4 // header + 3 data rows
	if len(lines) > maxRows {
		lines = lines[:maxRows]
	}

	for i, line := range lines {
		// Parse CSV line (simple parsing - assumes no commas in quoted fields for demo)
		fields := parseCSVLine(line)

		if i == 0 {
			// Header row
			fmt.Printf("%s%s│%s ", colorBold, colorYellow, colorReset)
			for j, field := range fields {
				if j > 0 {
					fmt.Printf("%s│%s ", colorYellow, colorReset)
				}
				// Truncate long headers
				displayField := field
				if len(field) > 25 {
					displayField = field[:22] + "..."
				}
				fmt.Printf("%s%-25s%s", colorBold+colorYellow, displayField, colorReset)
			}
			fmt.Printf(" %s│%s\n", colorYellow, colorReset)
			fmt.Printf("%s%s%s\n", colorYellow, strings.Repeat("─", 80), colorReset)
		} else {
			// Data row
			fmt.Printf("%s│%s ", colorGray, colorReset)
			for j, field := range fields {
				if j > 0 {
					fmt.Printf("%s│%s ", colorGray, colorReset)
				}
				// Truncate long fields
				displayField := field
				if len(field) > 25 {
					displayField = field[:22] + "..."
				}
				fmt.Printf("%-25s", displayField)
			}
			fmt.Printf(" %s│%s\n", colorGray, colorReset)
		}
	}

	fmt.Printf("%s%s%s\n", colorGray, strings.Repeat("─", 80), colorReset)

	// Show total row count
	totalRows, _ := countCSVRows(filePath)
	if totalRows > 3 {
		fmt.Printf("%s  ... and %d more rows%s\n", colorGray, totalRows-3, colorReset)
	}
	fmt.Printf("%s  Total: %d data rows%s\n\n", colorCyan, totalRows, colorReset)

	return nil
}

// parseCSVLine parses a CSV line handling quoted fields
func parseCSVLine(line string) []string {
	var fields []string
	var currentField strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		char := line[i]

		if char == '"' {
			inQuotes = !inQuotes
		} else if char == ',' && !inQuotes {
			fields = append(fields, strings.TrimSpace(currentField.String()))
			currentField.Reset()
		} else {
			currentField.WriteByte(char)
		}
	}

	fields = append(fields, strings.TrimSpace(currentField.String()))
	return fields
}

// countCSVRows counts the number of data rows (excluding header)
func countCSVRows(filePath string) (int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) <= 1 {
		return 0, nil
	}

	return len(lines) - 1, nil
}

// printSummaryTable prints a formatted summary table
func printSummaryTable(datasetID, batchID, studyID string) {
	fmt.Printf("\n%s%s%s\n", colorBold, colorGreen, strings.Repeat("═", 60))
	fmt.Printf("                    🎉 ORCHESTRATION COMPLETE 🎉")
	fmt.Printf("\n%s%s\n\n", strings.Repeat("═", 60), colorReset)

	data := []struct {
		label string
		value string
		icon  string
	}{
		{"Dataset ID", datasetID, "📊"},
		{"Batch ID", batchID, "📦"},
		{"Study ID", studyID, "🔬"},
		{"Status", "ACTIVE", "✅"},
	}

	for _, item := range data {
		fmt.Printf("  %s %s%-12s%s %s%s%s\n",
			item.icon,
			colorBold+colorWhite, item.label+":", colorReset,
			colorCyan, item.value, colorReset)
	}

	fmt.Printf("\n%s%s%s\n", colorGreen, strings.Repeat("═", 60), colorReset)

	// Final pause for Q&A
	fmt.Printf("\n%s%sDemo complete! Ready for questions...%s\n\n", colorBold, colorBlue, colorReset)
}

func main() {
	// Check for tail-only mode
	if len(os.Args) >= 3 && os.Args[1] == "--tail-only" {
		batchID := os.Args[2]
		label := "BATCH"
		labelColor := colorGreen

		if len(os.Args) >= 4 {
			label = os.Args[3]
		}
		if len(os.Args) >= 5 {
			// Support custom colors: green, blue, purple, cyan, yellow
			switch os.Args[4] {
			case "blue":
				labelColor = colorBlue
			case "purple":
				labelColor = colorPurple
			case "cyan":
				labelColor = colorCyan
			case "yellow":
				labelColor = colorYellow
			}
		}

		fmt.Printf("\n%s%s════════════════════════════════════════════════════════════╗%s\n", colorBold, labelColor, colorReset)
		fmt.Printf("%s%s            📊 Response Monitor: %s%s\n", colorBold, labelColor, label, colorReset)
		fmt.Printf("%s%s════════════════════════════════════════════════════════════╝%s\n", colorBold, labelColor, colorReset)
		fmt.Printf("%s%sℹ Monitoring Batch ID: %s%s\n", colorBold, colorCyan, batchID, colorReset)
		fmt.Printf("%s%sℹ Press Ctrl+C to stop%s\n\n", colorBold, colorCyan, colorReset)

		// Setup signal handling
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Start tailing
		go tailResponses(ctx, batchID, label, labelColor)

		// Wait for interrupt
		<-sigChan
		fmt.Printf("\n\n%s%s🛑 Shutting down...%s\n", colorBold, colorYellow, colorReset)
		cancel()
		time.Sleep(500 * time.Millisecond)
		fmt.Printf("%s%s✓ Stopped%s\n\n", colorBold, colorGreen, colorReset)
		return
	}

	// Check for required arguments
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  Full demo:  %s <batch_name> <workspace_id> <project_id> [--tail]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Tail only:  %s --tail-only <batch_id> [label] [color]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s \"My Batch\" \"6655b8281cc82a88996f0bb8\" \"6655b8281cc82a88996f0bb8\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --tail-only \"6655b8281cc82a88996f0bb8\" \"Task-A\" \"green\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nColors: green, blue, purple, cyan, yellow\n")
		os.Exit(1)
	}
	batchName := os.Args[1]
	workspaceID := os.Args[2]
	projectID := os.Args[3]

	// Check for --tail flag
	enableTailing := false
	for _, arg := range os.Args[4:] {
		if arg == "--tail" {
			enableTailing = true
			break
		}
	}

	// Print header
	fmt.Println()

	// Read and display ASCII banner
	banner, err := os.ReadFile("docs/scripts/banner.txt")
	if err == nil {
		fmt.Printf("%s%s%s\n", colorProlificBlue, string(banner), colorReset)
	} else {
		// Fallback if banner file not found
		fmt.Printf("%s╔════════════════════════════════════════════════════════════╗\n", colorProlificBlue)
		fmt.Printf("║          AI TASK BUILDER ORCHESTRATION - DEMO             ║\n")
		fmt.Printf("╚════════════════════════════════════════════════════════════╝%s\n", colorReset)
	}
	fmt.Println()
	printInfo("Batch Name", batchName)
	printInfo("Workspace ID", workspaceID)
	printInfo("Project ID", projectID)
	fmt.Printf("\n%s%s💡 Interactive Demo Mode - Press Enter to advance through each step%s\n", colorBold, colorYellow, colorReset)

	// Step 1: Create Dataset
	printSectionHeader(1, "Creating dataset")
	printCommand(fmt.Sprintf("./prolific aitaskbuilder dataset create -n %s -w %s", batchName, workspaceID))
	datasetID, err := createDataset(batchName, workspaceID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Failed to create dataset: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	printSuccess(fmt.Sprintf("Dataset created with ID: %s", datasetID))

	// Step 2: Upload Dataset
	printSectionHeader(2, "Uploading dataset file")
	printCommand(fmt.Sprintf("./prolific aitaskbuilder dataset upload -d %s -f %s", datasetID, csvFilePath))
	err = uploadDataset(datasetID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Failed to upload dataset: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	printSuccess("Dataset uploaded successfully")

	// Show dataset preview
	err = printDatasetPreview(csvFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s⚠ Could not display dataset preview: %v%s\n", colorYellow, err, colorReset)
	}

	// Step 3: Check Dataset Status
	printSectionHeader(3, "Checking dataset status")
	printCommand(fmt.Sprintf("./prolific aitaskbuilder dataset check -d %s", datasetID))
	err = checkDataset(datasetID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Failed to check dataset: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	printSuccess("Dataset status verified")

	// Step 4: Create Batch
	printSectionHeader(4, "Creating batch")
	printCommand(fmt.Sprintf("./prolific aitaskbuilder batch create -n \"%s\" -w %s -d %s --task-name \"%s\" --task-introduction \"%s\" --task-steps \"%s\"",
		batchName, workspaceID, datasetID, taskName, taskIntro, taskSteps))
	batchID, err := createBatch(batchName, datasetID, workspaceID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Failed to create batch: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	printSuccess(fmt.Sprintf("Batch created with ID: %s", batchID))

	// Step 5: Add Instructions to Batch
	printSectionHeader(5, "Adding instructions to batch")
	printCommand(fmt.Sprintf("./prolific aitaskbuilder batch instructions -b %s -j @%s", batchID, instructionsFile))
	err = addInstructions(batchID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Failed to add instructions: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	printSuccess("Instructions added successfully")

	// Step 6: Setup Batch
	printSectionHeader(6, "Setting up batch")
	printCommand(fmt.Sprintf("./prolific aitaskbuilder batch setup -b %s -d %s --tasks-per-group 1", batchID, datasetID))
	err = setupBatch(batchID, datasetID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Failed to setup batch: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	printSuccess("Batch setup completed successfully")

	// Step 7: Check Batch Status
	printSectionHeader(7, "Checking batch status")
	printCommand(fmt.Sprintf("./prolific aitaskbuilder batch check -b %s", batchID))
	err = checkBatch(batchID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Failed to check batch: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	printSuccess("Batch status verified")

	// Step 8: View Batch
	printSectionHeader(8, "Viewing batch details")
	printCommand(fmt.Sprintf("./prolific aitaskbuilder batch view -b %s", batchID))
	err = viewBatch(batchID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Failed to view batch: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	printSuccess("Batch details retrieved")

	// Step 9: Create Prolific Study
	printSectionHeader(9, "Creating Prolific study linked to batch")
	printCommand(fmt.Sprintf("./prolific study create -t %s", tmpStudyTemplateFile))
	studyID, err := createStudy(batchID, projectID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Failed to create study: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	printSuccess(fmt.Sprintf("Study created with ID: %s", studyID))

	// Step 10: View Study Details
	printSectionHeader(10, "Viewing study details")
	printCommand(fmt.Sprintf("./prolific study view %s", studyID))
	err = viewStudy(studyID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Failed to view study: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	printSuccess("Study details retrieved")

	// Step 11: Publish Study
	printSectionHeader(11, "Publishing study")
	printCommand(fmt.Sprintf("./prolific study transition -a PUBLISH %s", studyID))
	err = publishStudy(studyID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Failed to publish study: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	printSuccess("Study published successfully")

	// Print summary
	printSummaryTable(datasetID, batchID, studyID)

	// Step 12 (Optional): Tail Responses
	if enableTailing {
		printSectionHeader(12, "Monitoring task responses in real-time")
		fmt.Printf("\n%s%s💡 This will poll for new responses every 3 seconds%s\n", colorBold, colorYellow, colorReset)
		fmt.Printf("%s%sℹ Press Ctrl+C to stop monitoring%s\n\n", colorBold, colorCyan, colorReset)

		// Setup signal handling for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Start tailing in a goroutine
		go tailResponses(ctx, batchID, "BATCH-1", colorGreen)

		// Wait for interrupt signal
		<-sigChan
		fmt.Printf("\n\n%s%s🛑 Received interrupt signal, shutting down gracefully...%s\n", colorBold, colorYellow, colorReset)
		cancel()
		time.Sleep(500 * time.Millisecond) // Give goroutines time to clean up
		fmt.Printf("%s%s✓ Response monitoring stopped%s\n\n", colorBold, colorGreen, colorReset)
	}
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

	if len(output) > 0 {
		fmt.Printf("%s  %s%s\n", colorGray, strings.TrimSpace(string(output)), colorReset)
	}
	return nil
}

// pollStatus polls a resource until it reaches READY status
func pollStatus(resourceType string, cmdArgs ...string) error {
	maxRetries := 30
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			showSpinner(retryDelay, fmt.Sprintf("Polling %s status (attempt %d/%d)...", resourceType, i+1, maxRetries))
		}

		cmd := exec.CommandContext(context.Background(), "./prolific", cmdArgs...)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
		}

		status := strings.TrimSpace(string(output))
		fmt.Printf("%s  Status: %s%s%s\n", colorGray, colorYellow, status, colorReset)

		// Check if status is READY
		if strings.Contains(status, "READY") {
			fmt.Printf("%s  %s is ready!%s\n", colorGreen, resourceType, colorReset)
			return nil
		}

		// If not UNINITIALISED or READY, something went wrong
		if !strings.Contains(status, "UNINITIALISED") && !strings.Contains(status, "PROCESSING") {
			return fmt.Errorf("%s in unexpected status: %s", resourceType, status)
		}
	}

	return fmt.Errorf("%s did not reach READY status after %d retries", resourceType, maxRetries)
}

// checkDataset executes the dataset check command and verifies it's READY
func checkDataset(datasetID string) error {
	return pollStatus("Dataset", "aitaskbuilder", "dataset", "check", "-d", datasetID)
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
func checkBatch(batchID string) error {
	return pollStatus("Batch", "aitaskbuilder", "batch", "check", "-b", batchID)
}

// addInstructions executes the batch instructions command
func addInstructions(batchID string) error {
	// Read instructions from file
	instructionsJSON, err := os.ReadFile(instructionsFile)
	if err != nil {
		return fmt.Errorf("failed to read instructions file: %w", err)
	}

	// Validate it's valid JSON
	var instructions []map[string]interface{}
	if err := json.Unmarshal(instructionsJSON, &instructions); err != nil {
		return fmt.Errorf("failed to parse instructions JSON: %w", err)
	}

	// #nosec G204 - instructionsJSON is read from a trusted file
	cmd := exec.CommandContext(context.Background(), "./prolific", "aitaskbuilder", "batch", "instructions",
		"-b", batchID,
		"-j", string(instructionsJSON))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	if len(output) > 0 {
		fmt.Printf("%s  %s%s\n", colorGray, strings.TrimSpace(string(output)), colorReset)
	}
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

	if len(output) > 0 {
		fmt.Printf("%s  %s%s\n", colorGray, strings.TrimSpace(string(output)), colorReset)
	}
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

	if len(output) > 0 {
		fmt.Printf("%s  Details:%s\n", colorGray, colorReset)
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			fmt.Printf("%s    %s%s\n", colorGray, line, colorReset)
		}
	}
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
	// defer os.Remove(tmpStudyTemplateFile)

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

	if len(output) > 0 {
		fmt.Printf("%s  Details:%s\n", colorGray, colorReset)
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			fmt.Printf("%s    %s%s\n", colorGray, line, colorReset)
		}
	}
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

	if len(output) > 0 {
		fmt.Printf("%s  %s%s\n", colorGray, strings.TrimSpace(string(output)), colorReset)
	}

	return nil
}

// tailResponses continuously polls and displays new batch responses
func tailResponses(ctx context.Context, batchID, label, labelColor string) {
	seenResponses := make(map[string]bool)
	pollInterval := 3 * time.Second

	// Initial header
	fmt.Printf("\n%s%s[%s] Starting response monitor...%s\n", colorBold, labelColor, label, colorReset)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("%s%s[%s] Stopping response monitor%s\n", colorBold, labelColor, label, colorReset)
			return
		case <-ticker.C:
			cmd := exec.CommandContext(ctx, "./prolific", "aitaskbuilder", "batch", "responses",
				"-b", batchID)

			output, err := cmd.CombinedOutput()
			if err != nil {
				// Silently continue on error - batch might not have responses yet
				continue
			}

			// Parse output to find individual responses
			lines := strings.Split(string(output), "\n")
			var currentResponseID string
			var responseLines []string

			for _, line := range lines {
				if strings.HasPrefix(line, "Response ") && len(responseLines) > 0 {
					// Process previous response
					if currentResponseID != "" && !seenResponses[currentResponseID] {
						printTailedResponse(label, labelColor, responseLines)
						seenResponses[currentResponseID] = true
					}
					responseLines = []string{line}
					currentResponseID = ""
				} else if strings.HasPrefix(strings.TrimSpace(line), "ID: ") && currentResponseID == "" {
					// Extract response ID
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						currentResponseID = strings.TrimSpace(parts[1])
					}
					responseLines = append(responseLines, line)
				} else if line != "" && !strings.HasPrefix(line, "AI Task Builder Batch Responses:") &&
					!strings.HasPrefix(line, "Batch ID:") &&
					!strings.HasPrefix(line, "Total Responses:") {
					responseLines = append(responseLines, line)
				}
			}

			// Process last response
			if currentResponseID != "" && !seenResponses[currentResponseID] && len(responseLines) > 0 {
				printTailedResponse(label, labelColor, responseLines)
				seenResponses[currentResponseID] = true
			}
		}
	}
}

// printTailedResponse prints a formatted response with color coding
func printTailedResponse(label, labelColor string, lines []string) {
	fmt.Printf("\n%s%s┌─ [%s] New Response ─────────────────────────────────%s\n",
		colorBold, labelColor, label, colorReset)

	for _, line := range lines {
		// Highlight important fields
		if strings.Contains(line, "Text:") || strings.Contains(line, "Selected Options:") {
			fmt.Printf("%s%s│%s %s%s\n", labelColor, colorBold, colorReset, line, colorReset)
		} else {
			fmt.Printf("%s│%s %s%s\n", labelColor, colorReset, colorGray, line)
		}
	}

	fmt.Printf("%s%s└────────────────────────────────────────────────────────%s\n",
		colorBold, labelColor, colorReset)
}
