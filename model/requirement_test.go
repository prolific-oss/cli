package model_test

import (
	"testing"

	"github.com/benmatselby/prolificli/model"
)

func TestFilterValueReturnsValueDependingOnRules(t *testing.T) {
	tt := []struct {
		Name     string
		Question string
		Title    string
		Expected string
	}{
		{"If query question is defined used it", "Query question", "Query title", "Query question"},
		{"If query question is not defined, and title is, use title", "", "Query title", "Query title"},
		{"If query question and title are not defined, it will be an empty string", "", "", ""},
	}
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			requirement := model.Requirement{
				Query: model.RequirementQuestion{
					Question: tc.Question,
					Title:    tc.Title,
				},
			}

			actual := requirement.FilterValue()

			if actual != tc.Expected {
				t.Fatalf("expected %v, got %v", tc.Expected, actual)
			}
		})
	}
}

func TestTitleReturnsValueDependingOnRules(t *testing.T) {
	tt := []struct {
		Name     string
		Question string
		Title    string
		Expected string
	}{
		{"If query question is defined used it", "Query question", "Query title", "Query question"},
		{"If query question is not defined, and title is, use title", "", "Query title", "Query title"},
		{"If query question and title are not defined, it will be an empty string", "", "", ""},
	}
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			requirement := model.Requirement{
				Query: model.RequirementQuestion{
					Question: tc.Question,
					Title:    tc.Title,
				},
			}

			actual := requirement.Title()

			if actual != tc.Expected {
				t.Fatalf("expected %v, got %v", tc.Expected, actual)
			}
		})
	}
}

func TestDescriptionReturnsValueDependingOnRules(t *testing.T) {
	tt := []struct {
		Name        string
		Category    string
		Description string
		Expected    string
	}{
		{"Use category and description if both exist", "requirement category", "requirement description", "Category: requirement category. requirement description"},
		{"Use only category if description does not exist", "requirement category", "", "Category: requirement category"},
	}
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			requirement := model.Requirement{
				Category: tc.Category,
				Query: model.RequirementQuestion{
					Description: tc.Description,
				},
			}

			actual := requirement.Description()

			if actual != tc.Expected {
				t.Fatalf("expected %v, got %v", tc.Expected, actual)
			}
		})
	}
}
