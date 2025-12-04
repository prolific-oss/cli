package study_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/study"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

var studyTemplate = model.CreateStudy{
	Name:                    "My first standard sample",
	InternalName:            "Standard sample",
	Description:             "This is my first standard sample study on the Prolific system.",
	ExternalStudyURL:        "https://eggs-experriment.com?participant={{%PROLIFIC_PID%}}",
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

func TestCreateCommandRejectsBothDataCollectionMethodAndExternalStudyURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// Create a temporary test file with both fields set
	tmpFile, err := os.CreateTemp("", "test-study-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	dataCollectionMethod := "AI_TASK_BUILDER"
	testContent := fmt.Sprintf(`{
		"name": "Test Study",
		"internal_name": "Test Study",
		"description": "Test",
		"external_study_url": "https://example.com",
		"data_collection_method": "%s",
		"prolific_id_option": "question",
		"completion_code": "TEST01",
		"total_available_places": 10,
		"estimated_completion_time": 10,
		"reward": 400
	}`, dataCollectionMethod)

	if _, err := tmpFile.Write([]byte(testContent)); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// CreateStudy should never be called because validation should fail first
	c.
		EXPECT().
		CreateStudy(gomock.Any()).
		Times(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewCreateCommand(c, writer)
	_ = cmd.Flags().Set("template-path", tmpFile.Name())
	err = cmd.RunE(cmd, nil)
	writer.Flush()

	expected := "error: data_collection_method and external_study_url are mutually exclusive: only one can be set"
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

func TestCreateCommandAcceptsOnlyExternalStudyURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// This should work - using the existing standard-sample.json which only has external_study_url
	c.
		EXPECT().
		CreateStudy(gomock.Any()).
		Return(&actualStudy, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewCreateCommand(c, writer)
	_ = cmd.Flags().Set("template-path", "../../docs/examples/standard-sample.json")
	err := cmd.RunE(cmd, nil)
	writer.Flush()

	if err != nil {
		t.Fatalf("expected no error when only external_study_url is set, got %v", err)
	}
}

func TestCreateCommandAcceptsOnlyDataCollectionMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// This should work - using the existing AITB example which only has data_collection_method
	c.
		EXPECT().
		CreateStudy(gomock.Any()).
		Return(&actualStudy, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewCreateCommand(c, writer)
	_ = cmd.Flags().Set("template-path", "../../docs/examples/standard-sample-aitaskbuilder.json")
	err := cmd.RunE(cmd, nil)
	writer.Flush()

	if err != nil {
		t.Fatalf("expected no error when only data_collection_method is set, got %v", err)
	}
}
