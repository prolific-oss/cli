package study_test

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/study"
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
		GetStudy(gomock.Eq(studyID)).
		Return(&actualStudy, nil).
		AnyTimes()

	c.
		EXPECT().
		GetSubmissions(gomock.Eq(studyID)).
		Return(&ls, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewViewCommand(c, writer)
	_ = cmd.RunE(cmd, []string{studyID})
	writer.Flush()

	expected := `
 My first standard sample


This is my first standard sample study on the Prolific system.

ID:                        11223344
Status:
Type:
Total cost:                £0.00
Reward:                    £4.00
Hourly rate:               £0.00
Estimated completion time: 10
Maximum allowed time:      10
Study URL:                 https://eggs-experriment.com?participant={{%PROLIFIC_PID%}}
Places taken:              0
Available places:          10

---

Eligibility requirements

No eligibility requirements are defined for this study.

---

Submissions

No submissions have been submitted for this study yet.
---

View study in the application: https://app.prolific.co/researcher/studies/11223344
`

	actual := stripansi.Strip(b.String())
	actual = strings.ReplaceAll(actual, " ", "")
	expected = strings.ReplaceAll(expected, " ", "")

	if actual != expected {
		t.Fatalf("expected \n'%s'\ngot\n'%s'", expected, actual)
	}
}
