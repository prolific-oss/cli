package aitaskbuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/prolific-oss/cli/client"
	"github.com/spf13/cobra"
)

// BatchInstructionsOptions are the options for creating AI Task Builder instructions.
type BatchInstructionsOptions struct {
	Args             []string
	BatchID          string
	InstructionsFile string
	InstructionsJSON string
}

// NewBatchInstructionsCommand creates a new command for creating AI Task Builder instructions.
func NewBatchInstructionsCommand(client client.API, w io.Writer) *cobra.Command {
	var opts BatchInstructionsOptions

	cmd := &cobra.Command{
		Use:   "instructions",
		Short: "Add instructions to a batch",
		Long: `Add instructions to an AI Task Builder batch

This command adds instructions to a batch that has already been created. Instructions
define the tasks that participants will complete. You can provide instructions either as
a JSON file or as a JSON string directly.

The instructions should be an array of instruction objects with the following types:
- multiple_choice: Instructions with predefined options
- free_text: Instructions requiring text input
- multiple_choice_with_free_text: Instructions with options and text input`,
		Example: `
Add instructions from a file:
$ prolific aitaskbuilder batch instructions -b <batch_id> -f instructions.json

Add instructions with JSON string:
$ prolific aitaskbuilder batch instructions -b <batch_id> -j '[{"type":"free_text","created_by":"Sean","description":"Please explain your choice."}]'

Example instructions.json:
[
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
  },
  {
    "type": "free_text",
    "created_by": "Sean",
    "description": "Please share the reasons for your choice."
  }
]
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			err := createBatchInstructions(client, opts, w)
			if err != nil {
				return fmt.Errorf("error: %s", err.Error())
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.BatchID, "batch-id", "b", "", "Batch ID (required) - The ID of the batch to add instructions to.")
	flags.StringVarP(&opts.InstructionsFile, "file", "f", "", "Path to JSON file containing instructions")
	flags.StringVarP(&opts.InstructionsJSON, "json", "j", "", "JSON string containing instructions")

	_ = cmd.MarkFlagRequired("batch-id")

	return cmd
}

// createBatchInstructions will create instructions for an AI Task Builder batch
func createBatchInstructions(c client.API, opts BatchInstructionsOptions, w io.Writer) error {
	if opts.BatchID == "" {
		return errors.New("batch ID is required")
	}

	// Validate that either file or json is provided, but not both
	if opts.InstructionsFile == "" && opts.InstructionsJSON == "" {
		return errors.New("either instructions file (-f) or JSON string (-j) must be provided")
	}

	if opts.InstructionsFile != "" && opts.InstructionsJSON != "" {
		return errors.New("cannot specify both instructions file (-f) and JSON string (-j)")
	}

	var instructionsData []byte
	var err error

	// Read instructions from file or use provided JSON string
	if opts.InstructionsFile != "" {
		instructionsData, err = os.ReadFile(opts.InstructionsFile)
		if err != nil {
			return fmt.Errorf("failed to read instructions file: %w", err)
		}
	} else {
		instructionsData = []byte(opts.InstructionsJSON)
	}

	// Parse the instructions JSON
	var instructions client.CreateAITaskBuilderInstructionsPayload

	// Try to parse as the new format (object with instructions key) first
	if err := json.Unmarshal(instructionsData, &instructions); err != nil {
		// If that fails, try to parse as array (legacy format) and wrap it
		var instructionArray []client.Instruction
		if err := json.Unmarshal(instructionsData, &instructionArray); err != nil {
			return fmt.Errorf("failed to parse instructions JSON: %w", err)
		}
		instructions = client.CreateAITaskBuilderInstructionsPayload{
			Instructions: instructionArray,
		}
	}

	// Validate instructions
	if err := validateInstructions(instructions); err != nil {
		return fmt.Errorf("invalid instructions: %w", err)
	}

	// Create the instructions
	response, err := c.CreateAITaskBuilderInstructions(opts.BatchID, instructions)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Successfully added instructions to batch %s\n", opts.BatchID)
	if response.Message != "" {
		fmt.Fprintf(w, "Message: %s\n", response.Message)
	}

	return nil
}

// validateInstructions validates the instruction payload
func validateInstructions(instructions client.CreateAITaskBuilderInstructionsPayload) error {
	if len(instructions.Instructions) == 0 {
		return errors.New("at least one instruction must be provided")
	}

	validTypes := map[string]bool{
		"multiple_choice":                true,
		"free_text":                      true,
		"multiple_choice_with_free_text": true,
	}

	for i, instruction := range instructions.Instructions {
		if instruction.Type == "" {
			return fmt.Errorf("instruction %d: type is required", i+1)
		}

		if !validTypes[instruction.Type] {
			return fmt.Errorf("instruction %d: invalid type '%s'. Must be one of: multiple_choice, free_text, multiple_choice_with_free_text", i+1, instruction.Type)
		}

		if instruction.CreatedBy == "" {
			return fmt.Errorf("instruction %d: created_by is required", i+1)
		}

		if instruction.Description == "" {
			return fmt.Errorf("instruction %d: description is required", i+1)
		}

		// Validate type-specific requirements
		if strings.Contains(instruction.Type, "multiple_choice") && len(instruction.Options) == 0 {
			return fmt.Errorf("instruction %d: options are required for type '%s'", i+1, instruction.Type)
		}

		// Validate options if present
		for j, option := range instruction.Options {
			if option.Label == "" {
				return fmt.Errorf("instruction %d, option %d: label is required", i+1, j+1)
			}
			if option.Value == "" {
				return fmt.Errorf("instruction %d, option %d: value is required", i+1, j+1)
			}
			// Heading is required for multiple_choice_with_free_text
			if instruction.Type == "multiple_choice_with_free_text" && option.Heading == "" {
				return fmt.Errorf("instruction %d, option %d: heading is required for type 'multiple_choice_with_free_text'", i+1, j+1)
			}
		}
	}

	return nil
}
