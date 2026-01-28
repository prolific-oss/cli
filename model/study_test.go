package model_test

import (
	"encoding/json"
	"testing"

	"github.com/prolific-oss/cli/model"
)

const testCompletionCode = "C1234567"

func TestFilterValueReturnsName(t *testing.T) {
	name := "Patterns of migratory birds"
	study := model.Study{
		ID:   "study-id",
		Name: name,
	}

	if study.FilterValue() != name {
		t.Fatalf("expected filter value to be %s, got %s", name, study.FilterValue())
	}
}

func TestTitleIsTheStudyName(t *testing.T) {
	name := "Patterns of migratory birds"
	study := model.Study{
		ID:   "study-id",
		Name: name,
	}

	if study.Title() != name {
		t.Fatalf("expected filter value to be %s, got %s", name, study.Title())
	}
}

func TestDescriptionReturnsADescriptiveString(t *testing.T) {
	study := model.Study{
		ID:                   "study-id",
		Name:                 "Patterns of migratory birds",
		Status:               model.StatusActive,
		StudyType:            "single",
		TotalAvailablePlaces: 515,
	}

	expected := "active - single - 515 places available"
	if study.Description() != expected {
		t.Fatalf("expected filter value to be %s, got %s", expected, study.Description())
	}
}

func TestGetCurrencyCodeCanFigureOutWhichCurrencyToUse(t *testing.T) {
	tt := []struct {
		Name                    string
		PresentmentCurrencyCode string
		CurrencyCode            string
		Expected                string
	}{
		{"Using presentment", "USD", "GBP", "USD"},
		{"Using currency", "", "EUR", "EUR"},
		{"Default", "", "", "GBP"},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			study := model.Study{
				PresentmentCurrencyCode: tc.PresentmentCurrencyCode,
				CurrencyCode:            tc.CurrencyCode,
			}

			actual := study.GetCurrencyCode()

			if actual != tc.Expected {
				t.Fatalf("expected %v, got %v", tc.Expected, actual)
			}
		})
	}
}

func TestCompletionCodeUnmarshal(t *testing.T) {
	jsonData := `{
		"code": "C1234567",
		"code_type": "COMPLETED",
		"actions": [
			{
				"action": "AUTOMATICALLY_APPROVE"
			}
		]
	}`

	var cc model.CompletionCode
	err := json.Unmarshal([]byte(jsonData), &cc)
	if err != nil {
		t.Fatalf("failed to unmarshal CompletionCode: %v", err)
	}

	if cc.Code != testCompletionCode {
		t.Errorf("expected code to be %s, got %s", testCompletionCode, cc.Code)
	}
	if cc.CodeType != "COMPLETED" {
		t.Errorf("expected code_type to be COMPLETED, got %s", cc.CodeType)
	}
	if len(cc.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(cc.Actions))
	}
	if cc.Actions[0].Action != "AUTOMATICALLY_APPROVE" {
		t.Errorf("expected action to be AUTOMATICALLY_APPROVE, got %s", cc.Actions[0].Action)
	}
}

func TestAccessDetailUnmarshal(t *testing.T) {
	jsonData := `{
		"external_url": "https://example.com/task1",
		"total_allocation": 100
	}`

	var ad model.AccessDetail
	err := json.Unmarshal([]byte(jsonData), &ad)
	if err != nil {
		t.Fatalf("failed to unmarshal AccessDetail: %v", err)
	}

	if ad.ExternalURL != "https://example.com/task1" {
		t.Errorf("expected external_url to be https://example.com/task1, got %s", ad.ExternalURL)
	}
	if ad.TotalAllocation != 100 {
		t.Errorf("expected total_allocation to be 100, got %d", ad.TotalAllocation)
	}
}

func TestCreateStudyWithCompletionCodes(t *testing.T) {
	jsonData := `{
		"name": "Test Study",
		"internal_name": "test-study",
		"description": "A test study",
		"prolific_id_option": "url_parameters",
		"completion_codes": [
			{
				"code": "C1234567",
				"code_type": "COMPLETED",
				"actions": [
					{
						"action": "AUTOMATICALLY_APPROVE"
					}
				]
			},
			{
				"code": "C7654321",
				"code_type": "REJECTED",
				"actions": [
					{
						"action": "AUTOMATICALLY_REJECT"
					}
				]
			}
		],
		"total_available_places": 100,
		"estimated_completion_time": 10,
		"reward": 1.5,
		"device_compatibility": ["desktop"]
	}`

	var study model.CreateStudy
	err := json.Unmarshal([]byte(jsonData), &study)
	if err != nil {
		t.Fatalf("failed to unmarshal CreateStudy: %v", err)
	}

	if len(study.CompletionCodes) != 2 {
		t.Fatalf("expected 2 completion codes, got %d", len(study.CompletionCodes))
	}
	if study.CompletionCodes[0].Code != testCompletionCode {
		t.Errorf("expected first code to be %s, got %s", testCompletionCode, study.CompletionCodes[0].Code)
	}
	if study.CompletionCodes[1].CodeType != "REJECTED" {
		t.Errorf("expected second code_type to be REJECTED, got %s", study.CompletionCodes[1].CodeType)
	}
}

func TestCreateStudyWithAccessDetails(t *testing.T) {
	jsonData := `{
		"name": "Test Study",
		"internal_name": "test-study",
		"description": "A test study",
		"prolific_id_option": "url_parameters",
		"access_details": [
			{
				"external_url": "https://example.com/task1",
				"total_allocation": 50
			},
			{
				"external_url": "https://example.com/task2",
				"total_allocation": 50
			}
		],
		"total_available_places": 100,
		"estimated_completion_time": 10,
		"reward": 1.5,
		"device_compatibility": ["desktop"]
	}`

	var study model.CreateStudy
	err := json.Unmarshal([]byte(jsonData), &study)
	if err != nil {
		t.Fatalf("failed to unmarshal CreateStudy: %v", err)
	}

	if len(study.AccessDetails) != 2 {
		t.Fatalf("expected 2 access details, got %d", len(study.AccessDetails))
	}
	if study.AccessDetails[0].ExternalURL != "https://example.com/task1" {
		t.Errorf("expected first URL to be https://example.com/task1, got %s", study.AccessDetails[0].ExternalURL)
	}
	if study.AccessDetails[1].TotalAllocation != 50 {
		t.Errorf("expected second allocation to be 50, got %d", study.AccessDetails[1].TotalAllocation)
	}
}

func TestCreateStudyBackwardCompatibilityWithCompletionCode(t *testing.T) {
	jsonData := `{
		"name": "Test Study",
		"internal_name": "test-study",
		"description": "A test study",
		"prolific_id_option": "url_parameters",
		"completion_code": "C1234567",
		"completion_option": "code",
		"total_available_places": 100,
		"estimated_completion_time": 10,
		"reward": 1.5,
		"device_compatibility": ["desktop"]
	}`

	var study model.CreateStudy
	err := json.Unmarshal([]byte(jsonData), &study)
	if err != nil {
		t.Fatalf("failed to unmarshal CreateStudy: %v", err)
	}

	if study.CompletionCode != testCompletionCode {
		t.Errorf("expected completion_code to be %s, got %s", testCompletionCode, study.CompletionCode)
	}
	if study.CompletionOption != "code" {
		t.Errorf("expected completion_option to be code, got %s", study.CompletionOption)
	}
}

func TestSubmissionsConfigWithAutoRejectionCategories(t *testing.T) {
	jsonData := `{
		"name": "Test Study",
		"internal_name": "test-study",
		"description": "A test study",
		"prolific_id_option": "url_parameters",
		"total_available_places": 100,
		"estimated_completion_time": 10,
		"reward": 1.5,
		"device_compatibility": ["desktop"],
		"submissions_config": {
			"max_submissions_per_participant": 5,
			"max_concurrent_submissions": 2,
			"auto_rejection_categories": ["EXCEPTIONALLY_FAST"]
		}
	}`

	var study model.CreateStudy
	err := json.Unmarshal([]byte(jsonData), &study)
	if err != nil {
		t.Fatalf("failed to unmarshal CreateStudy: %v", err)
	}

	if study.SubmissionsConfig.MaxSubmissionsPerParticipant != 5 {
		t.Errorf("expected max_submissions_per_participant to be 5, got %d", study.SubmissionsConfig.MaxSubmissionsPerParticipant)
	}
	if len(study.SubmissionsConfig.AutoRejectionCategories) != 1 {
		t.Fatalf("expected 2 auto_rejection_categories, got %d", len(study.SubmissionsConfig.AutoRejectionCategories))
	}
	if study.SubmissionsConfig.AutoRejectionCategories[0] != "EXCEPTIONALLY_FAST" {
		t.Errorf("expected first category to be EXCEPTIONALLY_FAST, got %s", study.SubmissionsConfig.AutoRejectionCategories[0])
	}
}

func TestCreateStudyWithNewOptionalFields(t *testing.T) {
	jsonData := `{
		"name": "Test Study",
		"internal_name": "test-study",
		"description": "A test study",
		"prolific_id_option": "url_parameters",
		"total_available_places": 100,
		"estimated_completion_time": 10,
		"reward": 1.5,
		"device_compatibility": ["desktop"],
		"filter_set_id": "filter-set-123",
		"filter_set_version": 2,
		"is_custom_screening": true,
		"content_warnings": ["VIOLENCE", "EXPLICIT_LANGUAGE"],
		"content_warning_details": "May contain violent imagery",
		"metadata": {
			"project_id": "proj-123",
			"researcher_notes": "Important study"
		},
		"is_external_study_url_secure": true
	}`

	var study model.CreateStudy
	err := json.Unmarshal([]byte(jsonData), &study)
	if err != nil {
		t.Fatalf("failed to unmarshal CreateStudy: %v", err)
	}

	if study.FilterSetID != "filter-set-123" {
		t.Errorf("expected filter_set_id to be filter-set-123, got %s", study.FilterSetID)
	}
	if study.FilterSetVersion != 2 {
		t.Errorf("expected filter_set_version to be 2, got %d", study.FilterSetVersion)
	}
	if !study.IsCustomScreening {
		t.Error("expected is_custom_screening to be true")
	}
	if len(study.ContentWarnings) != 2 {
		t.Fatalf("expected 2 content_warnings, got %d", len(study.ContentWarnings))
	}
	if study.ContentWarningDetails != "May contain violent imagery" {
		t.Errorf("expected content_warning_details to match, got %s", study.ContentWarningDetails)
	}
	if study.Metadata["project_id"] != "proj-123" {
		t.Error("expected metadata project_id to be proj-123")
	}
	if !study.IsExternalStudyURLSecure {
		t.Error("expected is_external_study_url_secure to be true")
	}
}
