package study_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/benmatselby/prolificli/cmd/study"
	"github.com/benmatselby/prolificli/config"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewIncreasePlacesCommandRendersBasicUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := study.NewIncreasePlacesCommand(client, os.Stdout)

	use := "increase-places"
	short := "Increase the total available places on a study"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewIncreasePlacesCommandValidatesUserInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

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
		GetStudy(gomock.Eq(studyID)).
		Return(&actualStudy, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewIncreasePlacesCommand(c, writer)
	_ = cmd.Flags().Set("places", "9")
	err := cmd.RunE(cmd, []string{studyID})

	expected := "study currently has 10 places, and you cannot decrease the available places to 9"
	if err.Error() != expected {
		t.Fatalf("expected %s, got %s", expected, err.Error())
	}
}

func TestNewIncreasePlacesCommandHandlesGetStudyFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"
	c.
		EXPECT().
		GetStudy(gomock.Eq(studyID)).
		Return(nil, errors.New("failed to get study")).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewIncreasePlacesCommand(c, writer)
	_ = cmd.Flags().Set("places", "9")
	err := cmd.RunE(cmd, []string{studyID})

	expected := "error: failed to get study"
	if err.Error() != expected {
		t.Fatalf("expected %s, got %s", expected, err.Error())
	}
}

func TestNewIncreasePlacesCommandHandlesFailureToUpdateStudy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

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

	updateStudy := model.UpdateStudy{
		TotalAvailablePlaces: 11,
	}

	c.
		EXPECT().
		GetStudy(gomock.Eq(studyID)).
		Return(&actualStudy, nil).
		AnyTimes()

	c.
		EXPECT().
		UpdateStudy(gomock.Eq(studyID), gomock.Eq(updateStudy)).
		Return(nil, errors.New("failed to update study")).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewIncreasePlacesCommand(c, writer)
	_ = cmd.Flags().Set("places", "11")
	err := cmd.RunE(cmd, []string{studyID})

	expected := "failed to update study"
	if err.Error() != expected {
		t.Fatalf("expected %s, got %s", expected, err.Error())
	}
}

func TestNewIncreasePlacesCommandRendersStudyOnceUpdated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

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

	updateStudy := model.UpdateStudy{
		TotalAvailablePlaces: 11,
	}

	updatedStudy := model.Study{
		ID:                      studyID,
		Name:                    "My first standard sample",
		InternalName:            "Standard sample",
		Desc:                    "This is my first standard sample study on the Prolific system.",
		ExternalStudyURL:        "https://eggs-experriment.com?participant=",
		TotalAvailablePlaces:    11,
		EstimatedCompletionTime: 10,
		MaximumAllowedTime:      10,
		Reward:                  400,
		DeviceCompatibility:     []string{"desktop", "tablet", "mobile"},
	}

	c.
		EXPECT().
		GetStudy(gomock.Eq(studyID)).
		Return(&actualStudy, nil).
		AnyTimes()

	c.
		EXPECT().
		UpdateStudy(gomock.Eq(studyID), gomock.Eq(updateStudy)).
		Return(&updatedStudy, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewIncreasePlacesCommand(c, writer)
	_ = cmd.Flags().Set("places", "11")
	_ = cmd.RunE(cmd, []string{studyID})
	writer.Flush()

	expected := fmt.Sprintf(`My first standard sample
This is my first standard sample study on the Prolific system.

ID:                        11223344
Status:
Type:
Total cost:                £0.00
Reward:                    £4.00
Hourly rate:               £0.00
Estimated completion time: 10
Maximum allowed time:      10
Study URL:                 https://eggs-experriment.com?participant=
Places taken:              0
Available places:          11

---

Submissions configuration
Maxsubmissionsperparticipant: 0
Maxconcurrentsubmissions:     0

---

Filters
Nofiltersaredefinedforthisstudy.

---

View study in the application: %s/researcher/studies/11223344
`, config.GetApplicationURL())

	actual := stripansi.Strip(b.String())
	actual = strings.ReplaceAll(actual, " ", "")
	expected = strings.ReplaceAll(expected, " ", "")

	if actual != expected {
		t.Fatalf("expected \n'%s'\ngot\n'%s'", expected, actual)
	}
}
