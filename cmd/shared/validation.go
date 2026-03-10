package shared

import (
	"github.com/prolific-oss/cli/model"
)

// ExclusiveOptionsInput abstracts the fields needed for exclusive options validation
type ExclusiveOptionsInput struct {
	Options     []OptionInput
	AnswerLimit *int   // nil means not set
	TypeStr     string // instruction type as string
}

// OptionInput represents an option with an exclusive flag
type OptionInput struct {
	Exclusive bool
}

// IsMultipleChoiceType checks if a type string is a multiple choice type
func IsMultipleChoiceType(typeStr string) bool {
	return typeStr == string(model.InstructionTypeMultipleChoice) ||
		typeStr == string(model.InstructionTypeMultipleChoiceWithFreeText)
}

// ValidateExclusiveOptions validates exclusive option constraints.
// Returns an error message (without location prefix) or empty string if valid.
//
// Validation rules:
//  1. Only applies to multiple choice types
//  2. Exclusive options are not allowed with single select (answer_limit == 1)
//  3. At least one non-exclusive option is required when using exclusive options
//
// Note on answer_limit nil handling:
//   - answer_limit == nil (or 0 for int types) means "not specified"
//   - The API typically defaults to multi-select behavior when not specified
//   - We only block exclusive options when explicitly single-select (answer_limit == 1)
//   - answer_limit == -1 explicitly means multi-select (unlimited selections)
func ValidateExclusiveOptions(input ExclusiveOptionsInput) string {
	// Skip if not multiple choice type
	if !IsMultipleChoiceType(input.TypeStr) {
		return ""
	}

	// Count exclusive and non-exclusive options
	exclusiveCount := 0
	nonExclusiveCount := 0
	for _, opt := range input.Options {
		if opt.Exclusive {
			exclusiveCount++
		} else {
			nonExclusiveCount++
		}
	}

	// No exclusive options, nothing to validate
	if exclusiveCount == 0 {
		return ""
	}

	// Exclusive options are not allowed with single select (answer_limit == 1)
	if input.AnswerLimit != nil && *input.AnswerLimit == 1 {
		return ErrExclusiveWithSingleSelect
	}

	// At least one non-exclusive option is required when using exclusive options
	if nonExclusiveCount == 0 {
		return ErrNoNonExclusiveOptions
	}

	return ""
}

// ValidateExclusiveOptionsWithIntLimit is a convenience function for types
// where AnswerLimit is an int (zero value means not set).
// It converts the int to a pointer for validation.
func ValidateExclusiveOptionsWithIntLimit(options []OptionInput, answerLimit int, typeStr string) string {
	var limitPtr *int
	if answerLimit != 0 {
		limitPtr = &answerLimit
	}
	return ValidateExclusiveOptions(ExclusiveOptionsInput{
		Options:     options,
		AnswerLimit: limitPtr,
		TypeStr:     typeStr,
	})
}
