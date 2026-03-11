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

// testCredPoolID is a test fixture representing a credential pool ID in the format {workspace_id}_{uuid}
const testCredPoolID = "679271425fe00981084a5f58_a856d700-c495-11f0-adce-338d4126f6e8"

func TestNewSetCredentialPoolCommandRendersBasicUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := study.NewSetCredentialPoolCommand(client, os.Stdout)

	use := "set-credential-pool <study-id>"
	short := "Set or update the credential pool on a draft study"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewSetCredentialPoolCommandRequiresCredentialPoolID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewSetCredentialPoolCommand(c, writer)
	err := cmd.RunE(cmd, []string{studyID})

	expected := "credential pool ID is required"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected error '%s', got '%v'", expected, err)
	}
}

func TestNewSetCredentialPoolCommandHandlesFailureToUpdateStudy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	updateStudy := model.UpdateStudy{
		CredentialPoolID: testCredPoolID,
	}

	c.
		EXPECT().
		UpdateStudy(gomock.Eq(studyID), gomock.Eq(updateStudy)).
		Return(nil, errors.New("failed to update study")).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewSetCredentialPoolCommand(c, writer)
	_ = cmd.Flags().Set("credential-pool-id", testCredPoolID)
	err := cmd.RunE(cmd, []string{studyID})

	expected := "failed to update study"
	if err.Error() != expected {
		t.Fatalf("expected %s, got %s", expected, err.Error())
	}
}

func TestNewSetCredentialPoolCommandRendersStudyOnceUpdated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	updateStudy := model.UpdateStudy{
		CredentialPoolID: testCredPoolID,
	}

	updatedStudy := model.Study{
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
		UpdateStudy(gomock.Eq(studyID), gomock.Eq(updateStudy)).
		Return(&updatedStudy, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewSetCredentialPoolCommand(c, writer)
	_ = cmd.Flags().Set("credential-pool-id", testCredPoolID)
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
