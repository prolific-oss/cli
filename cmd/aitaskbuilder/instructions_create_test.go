package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewInstructionsCreateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewInstructionsCreateCommand(c, os.Stdout)

	use := "create"
	short := "Create instructions for an AI Task Builder batch"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewInstructionsCreateCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e12345"

	response := client.CreateAITaskBuilderInstructionsResponse{
		Message: "Instructions created successfully",
	}

	instructions := client.CreateAITaskBuilderInstructionsPayload{
		Instructions: []client.Instruction{
			{
				Type:        "free_text",
				CreatedBy:   "Sean",
				Description: "Please explain your decision.",
			},
		},
	}

	c.EXPECT().CreateAITaskBuilderInstructions(batchID, instructions).Return(&response, nil)

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewInstructionsCreateCommand(c, writer)

	instructionsJSON := `[{"type":"free_text","created_by":"Sean","description":"Please explain your decision."}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error; got %s", err.Error())
	}

	writer.Flush()

	expectedOutput := "Successfully created instructions for batch " + batchID
	if !strings.Contains(buf.String(), expectedOutput) {
		t.Fatalf("expected output to contain '%s'; got %s", expectedOutput, buf.String())
	}

	expectedMessage := "Message: Instructions created successfully"
	if !strings.Contains(buf.String(), expectedMessage) {
		t.Fatalf("expected output to contain '%s'; got %s", expectedMessage, buf.String())
	}
}

func TestNewInstructionsCreateCommandWithFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e54321"

	// Create a temporary file with instructions
	tmpDir := t.TempDir()
	instructionsFile := filepath.Join(tmpDir, "instructions.json")
	instructionsContent := `[
		{
			"type": "multiple_choice",
			"created_by": "Sean",
			"description": "Choose the LLM response which is more accurate.",
			"options": [
				{
					"label": "Response 1",
					"value": "response1"
				},
				{
					"label": "Response 2",
					"value": "response2"
				}
			]
		}
	]`

	err := os.WriteFile(instructionsFile, []byte(instructionsContent), 0600)
	if err != nil {
		t.Fatalf("failed to create temporary file: %s", err.Error())
	}

	response := client.CreateAITaskBuilderInstructionsResponse{
		Message: "Instructions created successfully",
	}

	instructions := client.CreateAITaskBuilderInstructionsPayload{
		Instructions: []client.Instruction{
			{
				Type:        "multiple_choice",
				CreatedBy:   "Sean",
				Description: "Choose the LLM response which is more accurate.",
				Options: []client.InstructionOption{
					{Label: "Response 1", Value: "response1"},
					{Label: "Response 2", Value: "response2"},
				},
			},
		},
	}

	c.EXPECT().CreateAITaskBuilderInstructions(batchID, instructions).Return(&response, nil)

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewInstructionsCreateCommand(c, writer)

	cmd.SetArgs([]string{
		"-b", batchID,
		"-f", instructionsFile,
	})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error; got %s", err.Error())
	}

	writer.Flush()

	expectedOutput := "Successfully created instructions for batch " + batchID
	if !strings.Contains(buf.String(), expectedOutput) {
		t.Fatalf("expected output to contain '%s'; got %s", expectedOutput, buf.String())
	}
}

func TestNewInstructionsCreateCommandAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "12344894-65b3-779e-aaf6-348698e23634"

	instructions := client.CreateAITaskBuilderInstructionsPayload{
		Instructions: []client.Instruction{
			{
				Type:        "free_text",
				CreatedBy:   "John",
				Description: "Please explain your choice.",
			},
		},
	}

	c.EXPECT().CreateAITaskBuilderInstructions(batchID, instructions).Return(nil, errors.New("API error"))

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewInstructionsCreateCommand(c, writer)

	instructionsJSON := `[{"type":"free_text","created_by":"John","description":"Please explain your choice."}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "API error") {
		t.Fatalf("expected error to contain 'API error'; got %s", err.Error())
	}
}

func TestNewInstructionsCreateCommandMissingBatchID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewInstructionsCreateCommand(c, writer)

	instructionsJSON := `[{"type":"free_text","created_by":"Sean","description":"Please provide your explanation."}]`

	cmd.SetArgs([]string{
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "required flag") {
		t.Fatalf("expected error about required flag; got %s", err.Error())
	}
}

func TestNewInstructionsCreateCommandMissingInstructions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01951234-65b3-779e-aaf6-348698e23634"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewInstructionsCreateCommand(c, writer)

	cmd.SetArgs([]string{
		"-b", batchID,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "either instructions file (-f) or JSON string (-j) must be provided") {
		t.Fatalf("expected error about missing instructions; got %s", err.Error())
	}
}

func TestNewInstructionsCreateCommandBothFileAndJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-123e-aaf6-348698e23634"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewInstructionsCreateCommand(c, writer)

	instructionsJSON := `[{"type":"free_text","created_by":"Sean","description":"Please explain your reasoning."}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-f", "instructions.json",
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "cannot specify both instructions file (-f) and JSON string (-j)") {
		t.Fatalf("expected error about both file and JSON; got %s", err.Error())
	}
}

func TestNewInstructionsCreateCommandInvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-89b3-123e-aaf6-348698e23634"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewInstructionsCreateCommand(c, writer)

	invalidJSON := `[{"type":"free_text","created_by":"Sean"description":"Please explain your choice."}]` // missing comma

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", invalidJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "failed to parse instructions JSON") {
		t.Fatalf("expected error about invalid JSON; got %s", err.Error())
	}
}

func TestNewInstructionsCreateCommandInvalidInstructionType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23699"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewInstructionsCreateCommand(c, writer)

	instructionsJSON := `[{"type":"invalid_type","created_by":"Sean","description":"Please explain your choice."}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "invalid type 'invalid_type'") {
		t.Fatalf("expected error about invalid type; got %s", err.Error())
	}
}
