package study_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/study"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

// Helper function to create string pointers for tests
func stringPtr(s string) *string {
	return &s
}

var studyTemplate = model.CreateStudy{
	Name:                    "My first standard sample",
	InternalName:            "Standard sample",
	Description:             "This is my first standard sample study on the Prolific system.",
	ExternalStudyURL:        stringPtr("https://eggs-experriment.com?participant={{%PROLIFIC_PID%}}"),
	ProlificIDOption:        "url_parameters",
	CompletionCode:          "COMPLE01",
	TotalAvailablePlaces:    10,
	EstimatedCompletionTime: 10,
	MaximumAllowedTime:      10,
	Reward:                  400,
	DeviceCompatibility:     []string{"desktop", "tablet", "mobile"},
	PeripheralRequirements:  []string{"audio", "camera", "download", "microphone"},
	SubmissionsConfig: struct {
		MaxSubmissionsPerParticipant int `json:"max_submissions_per_participant" mapstructure:"max_submissions_per_participant"`
		MaxConcurrentSubmissions     int `json:"max_concurrent_submissions" mapstructure:"max_concurrent_submissions"`
	}{
		MaxSubmissionsPerParticipant: -1,
		MaxConcurrentSubmissions:     0,
	},
}

var actualStudy = model.Study{
	Name:                    "My first standard sample",
	InternalName:            "Standard sample",
	Desc:                    "This is my first standard sample study on the Prolific system.",
	ExternalStudyURL:        "https://eggs-experriment.com?participant={{%PROLIFIC_PID%}}",
	TotalAvailablePlaces:    10,
	EstimatedCompletionTime: 10,
	MaximumAllowedTime:      10,
	Reward:                  400,
	DeviceCompatibility:     []string{"desktop", "tablet", "mobile"},
}

func TestNewCreateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := study.NewCreateCommand(client, os.Stdout)

	use := "create"
	short := "Creation of studies"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestCreateCommandHandlesFailureToReadConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	ls := client.ListSubmissionsResponse{}

	c.
		EXPECT().
		CreateStudy(gomock.Eq(studyTemplate)).
		Return(&actualStudy, nil).
		AnyTimes()

	c.
		EXPECT().
		GetSubmissions(gomock.Eq(actualStudy.ID), gomock.Eq(client.DefaultRecordLimit), gomock.Eq(client.DefaultRecordOffset)).
		Return(&ls, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewCreateCommand(c, writer)
	_ = cmd.Flags().Set("template-path", "broken-path.json")

	err := cmd.RunE(cmd, nil)
	writer.Flush()

	expected := "error: open broken-path.json: no such file or directory"
	if err.Error() != expected {
		t.Fatalf("expected %s, got %s", expected, err.Error())
	}
}

func TestCreateCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	ls := client.ListSubmissionsResponse{}

	c.
		EXPECT().
		CreateStudy(gomock.Eq(studyTemplate)).
		Return(&actualStudy, nil).
		AnyTimes()

	c.
		EXPECT().
		GetSubmissions(gomock.Eq(actualStudy.ID), gomock.Eq(client.DefaultRecordLimit), gomock.Eq(client.DefaultRecordOffset)).
		Return(&ls, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewCreateCommand(c, writer)
	_ = cmd.Flags().Set("template-path", "../../docs/examples/standard-sample.json")

	_ = cmd.RunE(cmd, nil)
	writer.Flush()
}

func TestCreateCommandCanPublish(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	ls := client.ListSubmissionsResponse{}
	tsr := client.TransitionStudyResponse{}

	c.
		EXPECT().
		CreateStudy(gomock.Eq(studyTemplate)).
		Return(&actualStudy, nil).
		MaxTimes(1)

	c.
		EXPECT().
		GetSubmissions(gomock.Eq(actualStudy.ID), gomock.Eq(client.DefaultRecordLimit), gomock.Eq(client.DefaultRecordOffset)).
		Return(&ls, nil).
		MaxTimes(1)

	c.
		EXPECT().
		TransitionStudy(gomock.Eq(actualStudy.ID), gomock.Eq(model.TransitionStudyPublish)).
		Return(&tsr, nil).
		MaxTimes(1)

	c.
		EXPECT().
		GetStudy(gomock.Eq(actualStudy.ID)).
		Return(&actualStudy, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewCreateCommand(c, writer)
	_ = cmd.Flags().Set("template-path", "../../docs/examples/standard-sample.json")
	_ = cmd.Flags().Set("publish", "true")
	_ = cmd.RunE(cmd, nil)
	writer.Flush()
}

func TestCommandFailsIfNoPathSpecified(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewCreateCommand(c, writer)
	_ = cmd.Flags().Set("publish", "true")
	err := cmd.RunE(cmd, nil)

	if err.Error() != "error: Can only create via a template YAML file at the moment" {
		t.Fatalf("Expected a specific error.")
	}

	writer.Flush()
}

func TestCreateCommandHandlesAnErrorFromTheAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		CreateStudy(gomock.Eq(studyTemplate)).
		Return(nil, fmt.Errorf("Whoopsie daisy")).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewCreateCommand(c, writer)
	_ = cmd.Flags().Set("template-path", "../../docs/examples/standard-sample.json")
	_ = cmd.Flags().Set("publish", "true")
	err := cmd.RunE(cmd, nil)

	if err.Error() != "error: Whoopsie daisy" {
		t.Fatalf("Expected a specific error, got %v", err)
	}
	writer.Flush()
}

func TestCreateCommandCanHandleErrorsWhenGettingStudy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	ls := client.ListSubmissionsResponse{}
	tsr := client.TransitionStudyResponse{}

	c.
		EXPECT().
		CreateStudy(gomock.Eq(studyTemplate)).
		Return(&actualStudy, nil).
		MaxTimes(1)

	c.
		EXPECT().
		GetSubmissions(gomock.Eq(actualStudy.ID), gomock.Eq(client.DefaultRecordLimit), gomock.Eq(client.DefaultRecordOffset)).
		Return(&ls, nil).
		MaxTimes(1)

	c.
		EXPECT().
		TransitionStudy(gomock.Eq(actualStudy.ID), gomock.Eq(model.TransitionStudyPublish)).
		Return(&tsr, nil).
		MaxTimes(1)

	c.
		EXPECT().
		GetStudy(gomock.Eq(actualStudy.ID)).
		Return(nil, errors.New("could not get study")).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewCreateCommand(c, writer)
	_ = cmd.Flags().Set("template-path", "../../docs/examples/standard-sample.json")
	_ = cmd.Flags().Set("publish", "true")
	err := cmd.RunE(cmd, nil)
	writer.Flush()

	expected := "error: could not get study"
	if err.Error() != expected {
		t.Fatalf("expected %s; got %v", expected, err.Error())
	}
}

func TestCreateCommandCanHandleErrorsWhenPublishing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	ls := client.ListSubmissionsResponse{}

	c.
		EXPECT().
		CreateStudy(gomock.Eq(studyTemplate)).
		Return(&actualStudy, nil).
		MaxTimes(1)

	c.
		EXPECT().
		GetSubmissions(gomock.Eq(actualStudy.ID), gomock.Eq(client.DefaultRecordLimit), gomock.Eq(client.DefaultRecordOffset)).
		Return(&ls, nil).
		MaxTimes(1)

	c.
		EXPECT().
		TransitionStudy(gomock.Eq(actualStudy.ID), gomock.Eq(model.TransitionStudyPublish)).
		Return(nil, errors.New("could not publish")).
		MaxTimes(1)

	c.
		EXPECT().
		GetStudy(gomock.Eq(actualStudy.ID)).
		Return(&actualStudy, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewCreateCommand(c, writer)
	_ = cmd.Flags().Set("template-path", "../../docs/examples/standard-sample.json")
	_ = cmd.Flags().Set("publish", "true")
	err := cmd.RunE(cmd, nil)
	writer.Flush()

	expected := "error: could not publish"
	if err.Error() != expected {
		t.Fatalf("expected %s; got %v", expected, err.Error())
	}
}

func TestValidateStudyConfiguration_ValidAITaskBuilder(t *testing.T) {
	studyData := model.CreateStudy{
		Name:                    "AI Task Builder Study",
		InternalName:            "AI Study",
		Description:             "Test AI Task Builder study",
		DataCollectionMethod:    stringPtr("DC_TOOL"),
		DataCollectionID:        stringPtr("batch-123"),
		AccessDetails:           []model.AccessDetail{{ExternalURL: "https://example.com", TotalAllocation: 10}},
		ProlificIDOption:        "not_required",
		CompletionCode:          "COMP123",
		TotalAvailablePlaces:    10,
		EstimatedCompletionTime: 5,
		MaximumAllowedTime:      10,
		Reward:                  100,
		DeviceCompatibility:     []string{"desktop"},
		PeripheralRequirements:  []string{},
	}

	err := study.ValidateStudyConfiguration(studyData)
	if err != nil {
		t.Fatalf("expected no validation error for valid AI Task Builder study, got: %v", err)
	}
}

func TestValidateStudyConfiguration_ValidTraditional(t *testing.T) {
	studyData := model.CreateStudy{
		Name:                    "Traditional Study",
		InternalName:            "Traditional",
		Description:             "Test traditional study",
		ExternalStudyURL:        stringPtr("https://example.com/study"),
		ProlificIDOption:        "url_parameters",
		CompletionCode:          "COMP123",
		CompletionOption:        "code",
		TotalAvailablePlaces:    10,
		EstimatedCompletionTime: 5,
		MaximumAllowedTime:      10,
		Reward:                  100,
		DeviceCompatibility:     []string{"desktop"},
		PeripheralRequirements:  []string{},
	}

	err := study.ValidateStudyConfiguration(studyData)
	if err != nil {
		t.Fatalf("expected no validation error for valid traditional study, got: %v", err)
	}
}

func TestValidateStudyConfiguration_InvalidMixed(t *testing.T) {
	studyData := model.CreateStudy{
		Name:                    "Invalid Mixed Study",
		InternalName:            "Mixed",
		Description:             "Invalid study with both AI and traditional fields",
		ExternalStudyURL:        stringPtr("https://example.com/study"),
		DataCollectionMethod:    stringPtr("DC_TOOL"),
		DataCollectionID:        stringPtr("batch-123"),
		ProlificIDOption:        "url_parameters",
		CompletionCode:          "COMP123",
		TotalAvailablePlaces:    10,
		EstimatedCompletionTime: 5,
		MaximumAllowedTime:      10,
		Reward:                  100,
		DeviceCompatibility:     []string{"desktop"},
		PeripheralRequirements:  []string{},
	}

	err := study.ValidateStudyConfiguration(studyData)
	if err == nil {
		t.Fatal("expected validation error for mixed study with both AI Task Builder and traditional fields")
	}

	expectedError := "study configuration error: cannot specify both AI Task Builder fields"
	if !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("expected error message to contain '%s', got: %v", expectedError, err.Error())
	}
}

func TestValidateStudyConfiguration_InvalidEmpty(t *testing.T) {
	studyData := model.CreateStudy{
		Name:                    "Invalid Empty Study",
		InternalName:            "Empty",
		Description:             "Invalid study with no required fields",
		ProlificIDOption:        "not_required",
		CompletionCode:          "COMP123",
		TotalAvailablePlaces:    10,
		EstimatedCompletionTime: 5,
		MaximumAllowedTime:      10,
		Reward:                  100,
		DeviceCompatibility:     []string{"desktop"},
		PeripheralRequirements:  []string{},
	}

	err := study.ValidateStudyConfiguration(studyData)
	if err == nil {
		t.Fatal("expected validation error for study with neither external_study_url nor AI Task Builder fields")
	}

	expectedError := "study configuration error: must specify either external_study_url"
	if !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("expected error message to contain '%s', got: %v", expectedError, err.Error())
	}
}
