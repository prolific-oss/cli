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
	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/study"
	"github.com/prolific-oss/cli/config"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
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

func TestViewStudyRendersCredentialPoolID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	actualStudy := model.Study{
		ID:                      studyID,
		Name:                    "Study with credential pool",
		InternalName:            "Study with credential pool",
		Desc:                    "This study demonstrates how to attach a credential pool for participant authentication",
		ExternalStudyURL:        "https://example.com/my-study-id",
		TotalAvailablePlaces:    50,
		EstimatedCompletionTime: 15,
		MaximumAllowedTime:      30,
		Reward:                  600,
		DeviceCompatibility:     []string{"desktop"},
		CredentialPoolID:        testCredPoolID,
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

	expected := fmt.Sprintf(`Study with credential pool
This study demonstrates how to attach a credential pool for participant authentication

ID:                        11223344
Status:
Type:
Total cost:                £0.00
Reward:                    £6.00
Hourly rate:               £0.00
Estimated completion time: 15
Maximum allowed time:      30
Study URL:                 https://example.com/my-study-id
Places taken:              0
Available places:          50
Credential Pool ID:        679271425fe00981084a5f58_a856d700-c495-11f0-adce-338d4126f6e8

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

func TestViewStudyRendersUnderpaying(t *testing.T) {
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
		IsUnderpaying: true,
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
Reward:                    £4.00(Underpaying)
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
