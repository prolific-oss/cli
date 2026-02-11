package aitaskbuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

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
- multiple_choice_with_free_text: Instructions with options and text input
- free_text_with_unit: Instructions requiring text input with unit selection (e.g., weight with kg/lbs)`,
		Example: `
Add instructions from a file:
$ prolific aitaskbuilder batch instructions -b <batch_id> -f instructions.json

Add instructions with JSON string:
$ prolific aitaskbuilder batch instructions -b <batch_id> -j '[{"type":"free_text","created_by":"Sean","description":"Please explain your choice."}]'

For a comprehensive example file with all instruction types, see:
docs/examples/batch-instructions.json
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
		return errors.New(ErrBatchIDRequired)
	}

	// Validate that either file or json is provided, but not both
	if opts.InstructionsFile == "" && opts.InstructionsJSON == "" {
		return errors.New(ErrInstructionInputRequired)
	}

	if opts.InstructionsFile != "" && opts.InstructionsJSON != "" {
		return errors.New(ErrBothInstructionInputsProvided)
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

	// Output the created instructions
	fmt.Fprintf(w, "Successfully added %d instruction(s) to batch %s\n", len(*response), opts.BatchID)
	for i, instruction := range *response {
		fmt.Fprintf(w, "\nInstruction %d:\n", i+1)
		fmt.Fprintf(w, "  ID: %s\n", instruction.ID)
		fmt.Fprintf(w, "  Type: %s\n", instruction.Type)
		fmt.Fprintf(w, "  Description: %s\n", instruction.Description)
		fmt.Fprintf(w, "  Created At: %s\n", instruction.CreatedAt)
		if len(instruction.Options) > 0 {
			fmt.Fprintf(w, "  Options: %d\n", len(instruction.Options))
		}
		if len(instruction.UnitOptions) > 0 {
			fmt.Fprintf(w, "  Unit Options: %d\n", len(instruction.UnitOptions))
		}
	}

	return nil
}

// validateInstructions validates the instruction payload
func validateInstructions(instructions client.CreateAITaskBuilderInstructionsPayload) error {
	if len(instructions.Instructions) == 0 {
		return errors.New(ErrAtLeastOneInstructionRequired)
	}

	validTypes := map[client.InstructionType]bool{
		client.InstructionTypeMultipleChoice:             true,
		client.InstructionTypeFreeText:                   true,
		client.InstructionTypeMultipleChoiceWithFreeText: true,
		client.InstructionTypeFreeTextWithUnit:           true,
	}

	for i, instruction := range instructions.Instructions {
		if err := validateInstructionBasicFields(instruction, i, validTypes); err != nil {
			return err
		}

		if err := validateInstructionTypeSpecificFields(instruction, i); err != nil {
			return err
		}

		if err := validateInstructionOptions(instruction, i); err != nil {
			return err
		}
	}

	return nil
}

// validateInstructionBasicFields validates required basic fields
func validateInstructionBasicFields(instruction client.Instruction, index int, validTypes map[client.InstructionType]bool) error {
	if instruction.Type == "" {
		return fmt.Errorf("instruction %d: type is required", index+1)
	}

	if !validTypes[instruction.Type] {
		return fmt.Errorf("instruction %d: invalid type '%s'. Must be one of: multiple_choice, free_text, multiple_choice_with_free_text, free_text_with_unit", index+1, instruction.Type)
	}

	if instruction.CreatedBy == "" {
		return fmt.Errorf("instruction %d: created_by is required", index+1)
	}

	if instruction.Description == "" {
		return fmt.Errorf("instruction %d: description is required", index+1)
	}

	return nil
}

// validateInstructionTypeSpecificFields validates fields specific to instruction type
func validateInstructionTypeSpecificFields(instruction client.Instruction, index int) error {
	// Validate options requirement for choice-based types
	if instruction.Type == client.InstructionTypeMultipleChoice ||
		instruction.Type == client.InstructionTypeMultipleChoiceWithFreeText {
		if len(instruction.Options) == 0 {
			return fmt.Errorf("instruction %d: options are required for type '%s'", index+1, instruction.Type)
		}
	}

	// Validate unit_options for free_text_with_unit
	if instruction.Type == client.InstructionTypeFreeTextWithUnit {
		return validateFreeTextWithUnit(instruction, index)
	}

	return nil
}

// validateFreeTextWithUnit validates free_text_with_unit specific fields
func validateFreeTextWithUnit(instruction client.Instruction, index int) error {
	if err := validateUnitOptions(instruction.UnitOptions, index); err != nil {
		return err
	}

	// Validate unit_position is required and has valid value
	if instruction.UnitPosition == "" {
		return fmt.Errorf("instruction %d: unit_position is required for type 'free_text_with_unit'", index+1)
	}
	if instruction.UnitPosition != "prefix" && instruction.UnitPosition != "suffix" {
		return fmt.Errorf("instruction %d: unit_position must be either 'prefix' or 'suffix', got '%s'", index+1, instruction.UnitPosition)
	}

	// Validate default_unit if provided (optional for free_text_with_unit)
	if instruction.DefaultUnit != "" {
		return validateDefaultUnit(instruction.DefaultUnit, instruction.UnitOptions, index)
	}

	return nil
}

// validateUnitOptions validates unit options array
func validateUnitOptions(unitOptions []client.UnitOption, index int) error {
	if len(unitOptions) < 2 {
		return fmt.Errorf("instruction %d: unit_options requires at least 2 options", index+1)
	}

	for j, unitOption := range unitOptions {
		if unitOption.Label == "" {
			return fmt.Errorf("instruction %d, unit_option %d: label is required", index+1, j+1)
		}
		if unitOption.Value == "" {
			return fmt.Errorf("instruction %d, unit_option %d: value is required", index+1, j+1)
		}
	}

	return nil
}

// validateDefaultUnit validates that default_unit matches one of the unit_options values
func validateDefaultUnit(defaultUnit string, unitOptions []client.UnitOption, index int) error {
	validUnit := false
	for _, opt := range unitOptions {
		if opt.Value == defaultUnit {
			validUnit = true
			break
		}
	}
	if !validUnit {
		return fmt.Errorf("instruction %d: default_unit '%s' must match one of the unit_options values", index+1, defaultUnit)
	}
	return nil
}

// validateInstructionOptions validates instruction options
func validateInstructionOptions(instruction client.Instruction, index int) error {
	for j, option := range instruction.Options {
		if option.Label == "" {
			return fmt.Errorf("instruction %d, option %d: label is required", index+1, j+1)
		}
		if option.Value == "" {
			return fmt.Errorf("instruction %d, option %d: value is required", index+1, j+1)
		}
		// Heading is required for multiple_choice_with_free_text
		if instruction.Type == client.InstructionTypeMultipleChoiceWithFreeText && option.Heading == "" {
			return fmt.Errorf("instruction %d, option %d: heading is required for type 'multiple_choice_with_free_text'", index+1, j+1)
		}
	}
	return nil
}
