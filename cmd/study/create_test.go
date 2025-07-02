package study_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/study"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

var studyTemplate = model.CreateStudy{
	Name:                    "My first standard sample",
	InternalName:            "Standard sample",
	Description:             "This is my first standard sample study on the Prolific system.",
	ExternalStudyURL:        "https://eggs-experriment.com?participant={{%PROLIFIC_PID%}}",
	ProlificIDOption:        "question",
	CompletionCode:          "COMPLE01",
	CompletionOption:        "code",
	TotalAvailablePlaces:    10,
	EstimatedCompletionTime: 10,
	MaximumAllowedTime:      10,
	Reward:                  400,
	DeviceCompatibility:     []string{"desktop", "tablet", "mobile"},
	PeripheralRequirements:  []string{"audio", "camera", "download", "microphone"},
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
