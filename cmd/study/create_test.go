package study_test

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/study"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

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

func TestCreateCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyTemplate := model.CreateStudy{
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

	actualStudy := model.Study{
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

	ls := client.ListSubmissionsResponse{}

	c.
		EXPECT().
		CreateStudy(gomock.Eq(studyTemplate)).
		Return(&actualStudy, nil).
		AnyTimes()

	c.
		EXPECT().
		GetSubmissions(gomock.Eq(actualStudy.ID)).
		Return(&ls, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewCreateCommand(c, writer)
	err := cmd.Flags().Set("template-path", "../../docs/examples/standard-sample.json")
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}
	_ = cmd.RunE(cmd, nil)
	writer.Flush()
}

func TestCreateCommandCanPublish(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyTemplate := model.CreateStudy{
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

	actualStudy := model.Study{
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

	ls := client.ListSubmissionsResponse{}
	tsr := client.TransitionStudyResponse{}

	c.
		EXPECT().
		CreateStudy(gomock.Eq(studyTemplate)).
		Return(&actualStudy, nil).
		MaxTimes(1)

	c.
		EXPECT().
		GetSubmissions(gomock.Eq(actualStudy.ID)).
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
	// _ = cmd.Flags().Set("template-path", "../../docs/examples/standard-sample.json")
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

	studyTemplate := model.CreateStudy{
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
