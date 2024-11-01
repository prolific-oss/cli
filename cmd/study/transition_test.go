package study_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/study"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewTransitionCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := study.NewTransitionCommand(client, os.Stdout)

	use := "transition"
	short := "Transition the status of a study"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestTransitionCommandCallsClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	response := client.TransitionStudyResponse{}

	actualStudy := model.Study{
		ID:                      studyID,
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

	c.
		EXPECT().
		TransitionStudy(gomock.Eq(studyID), gomock.Eq(model.TransitionStudyPause)).
		Return(&response, nil).
		AnyTimes()

	c.
		EXPECT().
		GetStudy(gomock.Eq(studyID)).
		Return(&actualStudy, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewTransitionCommand(c, writer)
	_ = cmd.Flags().Set("action", model.TransitionStudyPause)
	_ = cmd.Flags().Set("silent", "true")
	_ = cmd.RunE(cmd, []string{studyID})
	writer.Flush()

	expected := ""
	actual := b.String()

	if actual != expected {
		t.Fatalf("expected \n'%s'\ngot\n'%s'", expected, actual)
	}
}

func TestTransitionStudyHandlesApiErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	c.
		EXPECT().
		TransitionStudy(gomock.Eq(studyID), gomock.Eq(model.TransitionStudyPause)).
		Return(nil, errors.New("No no no")).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewTransitionCommand(c, writer)
	_ = cmd.Flags().Set("action", model.TransitionStudyPause)
	err := cmd.RunE(cmd, []string{studyID})
	writer.Flush()

	expected := "error: No no no"
	if err.Error() != expected {
		t.Fatalf("expected %s, got %s", expected, err.Error())
	}
}

func TestTransitionStudyHandlesNoActionSpecified(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewTransitionCommand(c, writer)
	err := cmd.RunE(cmd, []string{studyID})
	writer.Flush()

	expected := "error: you must provide an action to transition the study to"
	if err.Error() != expected {
		t.Fatalf("expected %s, got %s", expected, err.Error())
	}
}
