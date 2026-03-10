package shared

import (
	"testing"

	"github.com/prolific-oss/cli/model"
	"github.com/stretchr/testify/assert"
)

func TestIsMultipleChoiceType(t *testing.T) {
	tests := []struct {
		name     string
		typeStr  string
		expected bool
	}{
		{
			name:     "multiple_choice returns true",
			typeStr:  string(model.InstructionTypeMultipleChoice),
			expected: true,
		},
		{
			name:     "multiple_choice_with_free_text returns true",
			typeStr:  string(model.InstructionTypeMultipleChoiceWithFreeText),
			expected: true,
		},
		{
			name:     "free_text returns false",
			typeStr:  string(model.InstructionTypeFreeText),
			expected: false,
		},
		{
			name:     "free_text_with_unit returns false",
			typeStr:  string(model.InstructionTypeFreeTextWithUnit),
			expected: false,
		},
		{
			name:     "file_upload returns false",
			typeStr:  string(model.InstructionTypeFileUpload),
			expected: false,
		},
		{
			name:     "empty string returns false",
			typeStr:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMultipleChoiceType(tt.typeStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateExclusiveOptions(t *testing.T) {
	intPtr := func(v int) *int { return &v }

	tests := []struct {
		name     string
		input    ExclusiveOptionsInput
		expected string
	}{
		{
			name: "non-multiple choice type skips validation",
			input: ExclusiveOptionsInput{
				TypeStr: string(model.InstructionTypeFreeText),
				Options: []OptionInput{{Exclusive: true}},
			},
			expected: "",
		},
		{
			name: "no exclusive options passes validation",
			input: ExclusiveOptionsInput{
				TypeStr: string(model.InstructionTypeMultipleChoice),
				Options: []OptionInput{
					{Exclusive: false},
					{Exclusive: false},
				},
			},
			expected: "",
		},
		{
			name: "exclusive option with multi-select passes",
			input: ExclusiveOptionsInput{
				TypeStr:     string(model.InstructionTypeMultipleChoice),
				AnswerLimit: intPtr(-1),
				Options: []OptionInput{
					{Exclusive: false},
					{Exclusive: true},
				},
			},
			expected: "",
		},
		{
			name: "exclusive option with nil answer_limit passes",
			input: ExclusiveOptionsInput{
				TypeStr:     string(model.InstructionTypeMultipleChoice),
				AnswerLimit: nil,
				Options: []OptionInput{
					{Exclusive: false},
					{Exclusive: true},
				},
			},
			expected: "",
		},
		{
			name: "exclusive option with single select fails",
			input: ExclusiveOptionsInput{
				TypeStr:     string(model.InstructionTypeMultipleChoice),
				AnswerLimit: intPtr(1),
				Options: []OptionInput{
					{Exclusive: false},
					{Exclusive: true},
				},
			},
			expected: ErrExclusiveWithSingleSelect,
		},
		{
			name: "all options exclusive fails",
			input: ExclusiveOptionsInput{
				TypeStr:     string(model.InstructionTypeMultipleChoice),
				AnswerLimit: intPtr(-1),
				Options: []OptionInput{
					{Exclusive: true},
					{Exclusive: true},
				},
			},
			expected: ErrNoNonExclusiveOptions,
		},
		{
			name: "multiple_choice_with_free_text type validates correctly",
			input: ExclusiveOptionsInput{
				TypeStr:     string(model.InstructionTypeMultipleChoiceWithFreeText),
				AnswerLimit: intPtr(1),
				Options: []OptionInput{
					{Exclusive: true},
				},
			},
			expected: ErrExclusiveWithSingleSelect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateExclusiveOptions(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateExclusiveOptionsWithIntLimit(t *testing.T) {
	tests := []struct {
		name        string
		options     []OptionInput
		answerLimit int
		typeStr     string
		expected    string
	}{
		{
			name: "zero answer_limit treated as not set",
			options: []OptionInput{
				{Exclusive: false},
				{Exclusive: true},
			},
			answerLimit: 0,
			typeStr:     string(model.InstructionTypeMultipleChoice),
			expected:    "",
		},
		{
			name: "answer_limit 1 with exclusive fails",
			options: []OptionInput{
				{Exclusive: false},
				{Exclusive: true},
			},
			answerLimit: 1,
			typeStr:     string(model.InstructionTypeMultipleChoice),
			expected:    ErrExclusiveWithSingleSelect,
		},
		{
			name: "answer_limit -1 with exclusive passes",
			options: []OptionInput{
				{Exclusive: false},
				{Exclusive: true},
			},
			answerLimit: -1,
			typeStr:     string(model.InstructionTypeMultipleChoice),
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateExclusiveOptionsWithIntLimit(tt.options, tt.answerLimit, tt.typeStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}
