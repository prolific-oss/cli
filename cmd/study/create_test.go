package study_test

import (
	"bufio"
	"bytes"
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
	cmd.Flags().Set("template-path", "../../docs/examples/standard-sample.json")
	cmd.Run(cmd, nil)
	writer.Flush()
}
