package study_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/study"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewTestStudyCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := study.NewTestStudyCommand(c, os.Stdout)

	use := "test <study-id>"
	short := "Create a test run of a study"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestTestStudyCallsClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "55667788"

	actualStudy := model.Study{
		ID:                      "8888888",
		Name:                    "Test run of survey",
		InternalName:            "Survey test run",
		Desc:                    "A test run to validate the survey configuration.",
		ExternalStudyURL:        "https://survey.example.com?participant={{%PROLIFIC_PID%}}",
		TotalAvailablePlaces:    5,
		EstimatedCompletionTime: 15,
		MaximumAllowedTime:      20,
		Reward:                  200,
		DeviceCompatibility:     []string{"desktop"},
	}

	c.
		EXPECT().
		TestStudy(gomock.Eq(studyID)).
		Return(&actualStudy, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewTestStudyCommand(c, writer)
	_ = cmd.RunE(cmd, []string{studyID})
	writer.Flush()

	expected := fmt.Sprintf("%s\n", actualStudy.ID)
	actual := b.String()

	if actual != expected {
		t.Fatalf("expected \n'%s'\ngot\n'%s'", expected, actual)
	}
}

func TestTestStudyHandlesApiErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "55667788"

	c.
		EXPECT().
		TestStudy(gomock.Eq(studyID)).
		Return(nil, errors.New("study not found")).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewTestStudyCommand(c, writer)
	err := cmd.RunE(cmd, []string{studyID})
	writer.Flush()

	expected := "error: study not found"
	if err.Error() != expected {
		t.Fatalf("expected %s, got %s", expected, err.Error())
	}
}
