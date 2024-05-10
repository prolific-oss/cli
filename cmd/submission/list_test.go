package submission_test

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/submission"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewViewCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := submission.NewListCommand(client, os.Stdout)

	use := "list"
	short := "Provide details about your submissions, requires Study ID"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewSubmissionCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "777111999"

	submissionStart, _ := time.Parse("2006-01-02 15:04", "2022-07-24 08:04")

	response := client.ListSubmissionsResponse{
		Results: []model.Submission{
			{
				ID:            "1122",
				ParticipantID: "919",
				Status:        "APPROVED",
				StartedAt:     submissionStart,
				TimeTaken:     99,
			},
		},
		JSONAPIMeta: &client.JSONAPIMeta{
			Meta: struct {
				Count int `json:"count"`
			}{
				Count: 10,
			},
		},
	}

	c.
		EXPECT().
		GetSubmissions(gomock.Eq(studyID), gomock.Eq(client.DefaultRecordLimit), gomock.Eq(client.DefaultRecordOffset)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewListCommand(c, writer)
	_ = cmd.Flags().Set("study", studyID)
	_ = cmd.Flags().Set("limit", strconv.Itoa(client.DefaultRecordLimit))
	_ = cmd.Flags().Set("offset", strconv.Itoa(client.DefaultRecordOffset))
	_ = cmd.RunE(cmd, []string{studyID})

	writer.Flush()

	expected := `ParticipantID StartedAt                     TimeTaken StudyCode Status
919           2022-07-24 08:04:00 +0000 UTC 99                  APPROVED

Showing 1 record of 10
`

	actual := stripansi.Strip(b.String())
	actual = strings.ReplaceAll(actual, " ", "")
	expected = strings.ReplaceAll(expected, " ", "")

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}
