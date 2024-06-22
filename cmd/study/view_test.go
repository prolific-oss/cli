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

func TestNewStudyViewCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := study.NewViewCommand(client, os.Stdout)

	use := "view"
	short := "Provide details about your study, requires a Study ID"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestViewStudyRendersStudy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	actualStudy := model.Study{
		ID:                      studyID,
		Name:                    "My first standard sample",
		InternalName:            "Standard sample",
		Desc:                    "This is my first standard sample study on the Prolific system.",
		ExternalStudyURL:        "https://eggs-experriment.com?participant=",
		TotalAvailablePlaces:    10,
		EstimatedCompletionTime: 10,
		MaximumAllowedTime:      10,
		Reward:                  400,
		DeviceCompatibility:     []string{"desktop", "tablet", "mobile"},
		Filters: []model.Filter{
			{
				FilterID:       "handedness",
				SelectedValues: []string{"left"},
			},
		},
	}

	c.
		EXPECT().
		GetStudy(gomock.Eq(studyID)).
		Return(&actualStudy, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewViewCommand(c, writer)
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
Available places:          10

---

Submissions configuration
Maxsubmissionsperparticipant: 0
Maxconcurrentsubmissions:     0

---

Filters

handedness
-left

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

func TestViewStudyHandlesApiErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	c.
		EXPECT().
		GetStudy(gomock.Eq(studyID)).
		Return(nil, errors.New("unable to get study")).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewViewCommand(c, writer)
	err := cmd.RunE(cmd, []string{studyID})
	writer.Flush()

	expected := "error: unable to get study"
	if err.Error() != expected {
		t.Fatalf("expected %s, got %s", expected, err.Error())
	}
}
