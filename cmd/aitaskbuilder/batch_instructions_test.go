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
	"github.com/prolific-oss/cli/model"
)

func TestNewBatchInstructionsCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, os.Stdout)

	use := "instructions"
	short := "Add instructions to a batch"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewBatchInstructionsCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e12345"

	response := client.CreateAITaskBuilderInstructionsResponse{
		model.Instruction{
			ID:          "inst-123",
			Type:        "free_text",
			BatchID:     batchID,
			CreatedBy:   "Sean",
			CreatedAt:   "2024-09-18T07:50:15.055Z",
			Description: "Please explain your decision.",
		},
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

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

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

	expectedOutput := "Successfully added 1 instruction(s) to batch " + batchID
	if !strings.Contains(buf.String(), expectedOutput) {
		t.Fatalf("expected output to contain '%s'; got %s", expectedOutput, buf.String())
	}

	expectedID := "ID: inst-123"
	if !strings.Contains(buf.String(), expectedID) {
		t.Fatalf("expected output to contain '%s'; got %s", expectedID, buf.String())
	}
}

func TestNewBatchInstructionsCommandWithFile(t *testing.T) {
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
		model.Instruction{
			ID:          "inst-456",
			Type:        "multiple_choice",
			BatchID:     batchID,
			CreatedBy:   "Sean",
			CreatedAt:   "2024-09-18T07:50:15.055Z",
			Description: "Choose the LLM response which is more accurate.",
			Options: []model.InstructionOption{
				{Label: "Response 1", Value: "response1"},
				{Label: "Response 2", Value: "response2"},
			},
		},
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

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	cmd.SetArgs([]string{
		"-b", batchID,
		"-f", instructionsFile,
	})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error; got %s", err.Error())
	}

	writer.Flush()

	expectedOutput := "Successfully added 1 instruction(s) to batch " + batchID
	if !strings.Contains(buf.String(), expectedOutput) {
		t.Fatalf("expected output to contain '%s'; got %s", expectedOutput, buf.String())
	}
}

func TestNewBatchInstructionsCommandAPIError(t *testing.T) {
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

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

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

func TestNewBatchInstructionsCommandMissingBatchID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

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

func TestNewBatchInstructionsCommandMissingInstructions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01951234-65b3-779e-aaf6-348698e23634"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	cmd.SetArgs([]string{
		"-b", batchID,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), aitaskbuilder.ErrInstructionInputRequired) {
		t.Fatalf("expected error about missing instructions; got %s", err.Error())
	}
}

func TestNewBatchInstructionsCommandBothFileAndJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-123e-aaf6-348698e23634"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

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

	if !strings.Contains(err.Error(), aitaskbuilder.ErrBothInstructionInputsProvided) {
		t.Fatalf("expected error about both file and JSON; got %s", err.Error())
	}
}

func TestNewBatchInstructionsCommandInvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-89b3-123e-aaf6-348698e23634"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

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

func TestNewBatchInstructionsCommandInvalidInstructionType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23699"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

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

func TestNewBatchInstructionsCommandWithMultipleChoiceWithUnit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e12346"

	response := client.CreateAITaskBuilderInstructionsResponse{
		model.Instruction{
			ID:          "inst-789",
			Type:        "multiple_choice_with_unit",
			BatchID:     batchID,
			CreatedBy:   "Sean",
			CreatedAt:   "2024-09-18T07:50:15.055Z",
			Description: "What is your height?",
			Options: []model.InstructionOption{
				{Label: "150", Value: "150"},
				{Label: "160", Value: "160"},
			},
			UnitOptions: []model.UnitOption{
				{Label: "CM", Value: "cm"},
				{Label: "Inches", Value: "in"},
			},
			DefaultUnit: "cm",
		},
	}

	instructions := client.CreateAITaskBuilderInstructionsPayload{
		Instructions: []client.Instruction{
			{
				Type:        "multiple_choice_with_unit",
				CreatedBy:   "Sean",
				Description: "What is your height?",
				Options: []client.InstructionOption{
					{Label: "150", Value: "150"},
					{Label: "160", Value: "160"},
				},
				UnitOptions: []client.UnitOption{
					{Label: "CM", Value: "cm"},
					{Label: "Inches", Value: "in"},
				},
				DefaultUnit: "cm",
			},
		},
	}

	c.EXPECT().CreateAITaskBuilderInstructions(batchID, instructions).Return(&response, nil)

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	instructionsJSON := `[{
		"type": "multiple_choice_with_unit",
		"created_by": "Sean",
		"description": "What is your height?",
		"options": [
			{"label": "150", "value": "150"},
			{"label": "160", "value": "160"}
		],
		"unit_options": [
			{"label": "CM", "value": "cm"},
			{"label": "Inches", "value": "in"}
		],
		"default_unit": "cm"
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error; got %s", err.Error())
	}

	writer.Flush()

	expectedOutput := "Successfully added 1 instruction(s) to batch " + batchID
	if !strings.Contains(buf.String(), expectedOutput) {
		t.Fatalf("expected output to contain '%s'; got %s", expectedOutput, buf.String())
	}

	expectedID := "ID: inst-789"
	if !strings.Contains(buf.String(), expectedID) {
		t.Fatalf("expected output to contain '%s'; got %s", expectedID, buf.String())
	}

	expectedUnitOptions := "Unit Options: 2"
	if !strings.Contains(buf.String(), expectedUnitOptions) {
		t.Fatalf("expected output to contain '%s'; got %s", expectedUnitOptions, buf.String())
	}
}

func TestNewBatchInstructionsCommandMultipleChoiceWithUnitMissingUnitOptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23700"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	// Missing unit_options
	instructionsJSON := `[{
		"type": "multiple_choice_with_unit",
		"created_by": "Sean",
		"description": "What is your height?",
		"options": [
			{"label": "150", "value": "150"},
			{"label": "160", "value": "160"}
		]
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "unit_options requires at least 2 options") {
		t.Fatalf("expected error about missing unit_options; got %s", err.Error())
	}
}

func TestNewBatchInstructionsCommandMultipleChoiceWithUnitInsufficientUnitOptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23701"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	// Only 1 unit_option (need at least 2)
	instructionsJSON := `[{
		"type": "multiple_choice_with_unit",
		"created_by": "Sean",
		"description": "What is your height?",
		"options": [
			{"label": "150", "value": "150"},
			{"label": "160", "value": "160"}
		],
		"unit_options": [
			{"label": "CM", "value": "cm"}
		]
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "unit_options requires at least 2 options") {
		t.Fatalf("expected error about insufficient unit_options; got %s", err.Error())
	}
}

func TestNewBatchInstructionsCommandMultipleChoiceWithUnitMissingOptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23702"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	// Missing options (required for multiple_choice_with_unit)
	instructionsJSON := `[{
		"type": "multiple_choice_with_unit",
		"created_by": "Sean",
		"description": "What is your height?",
		"unit_options": [
			{"label": "CM", "value": "cm"},
			{"label": "Inches", "value": "in"}
		]
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "options are required for type 'multiple_choice_with_unit'") {
		t.Fatalf("expected error about missing options; got %s", err.Error())
	}
}

func TestNewBatchInstructionsCommandMultipleChoiceWithUnitInvalidDefaultUnit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23703"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	// default_unit doesn't match any unit_options value
	instructionsJSON := `[{
		"type": "multiple_choice_with_unit",
		"created_by": "Sean",
		"description": "What is your height?",
		"options": [
			{"label": "150", "value": "150"},
			{"label": "160", "value": "160"}
		],
		"unit_options": [
			{"label": "CM", "value": "cm"},
			{"label": "Inches", "value": "in"}
		],
		"default_unit": "meters"
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "default_unit 'meters' must match one of the unit_options values") {
		t.Fatalf("expected error about invalid default_unit; got %s", err.Error())
	}
}

func TestNewBatchInstructionsCommandMultipleChoiceWithUnitMissingDefaultUnit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	batchID := "01954894-65b3-779e-aaf6-348698e23703"

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	cmd := aitaskbuilder.NewBatchInstructionsCommand(c, writer)

	// default_unit is missing
	instructionsJSON := `[{
		"type": "multiple_choice_with_unit",
		"created_by": "Sean",
		"description": "What is your height?",
		"options": [
			{"label": "150", "value": "150"},
			{"label": "160", "value": "160"}
		],
		"unit_options": [
			{"label": "CM", "value": "cm"},
			{"label": "Inches", "value": "in"}
		]
	}]`

	cmd.SetArgs([]string{
		"-b", batchID,
		"-j", instructionsJSON,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error; got nil")
	}

	if !strings.Contains(err.Error(), "default_unit is required for type 'multiple_choice_with_unit'") {
		t.Fatalf("expected error about missing default_unit; got %s", err.Error())
	}
}
