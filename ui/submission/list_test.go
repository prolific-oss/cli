package submission_test

import (
	"bufio"
	"bytes"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/prolificli/client"
	"github.com/prolific-oss/prolificli/mock_client"
	"github.com/prolific-oss/prolificli/model"
	"github.com/prolific-oss/prolificli/ui/submission"
)

func TestCsvRendererRendersInCsvFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := submission.ListUsedOptions{
		StudyID: "1234",
	}

	started, _ := time.Parse("2006-01-02 15:04", "2022-07-24 08:04")

	actualSubmission := model.Submission{
		ID:            "23",
		ParticipantID: "999",
		StartedAt:     started,
		StudyCode:     "ALPHA1",
		Status:        "completed",
	}
	submissionResponse := client.ListSubmissionsResponse{
		Results: []model.Submission{actualSubmission},
	}

	c.
		EXPECT().
		GetSubmissions(gomock.Eq(opts.StudyID)).
		Return(&submissionResponse, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := submission.CsvRenderer{}
	err := renderer.Render(c, opts, writer)

	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	expected := `ParticipantID,StartedAt,TimeTaken,StudyCode,Status,
999,2022-07-24 08:04:00 +0000 UTC,0,ALPHA1,completed,
`

	if b.String() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, b.String())
	}
}

func TestCsvRendererRendersInCsvFormatAndRespectsFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := submission.ListUsedOptions{
		StudyID: "1234",
		Fields:  "ID,Status",
	}

	started, _ := time.Parse("2006-01-02 15:04", "2022-07-24 08:04")

	actualSubmission := model.Submission{
		ID:            "23",
		ParticipantID: "999",
		StartedAt:     started,
		StudyCode:     "ALPHA1",
		Status:        "completed",
	}
	submissionResponse := client.ListSubmissionsResponse{
		Results: []model.Submission{actualSubmission},
	}

	c.
		EXPECT().
		GetSubmissions(gomock.Eq(opts.StudyID)).
		Return(&submissionResponse, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := submission.CsvRenderer{}
	err := renderer.Render(c, opts, writer)

	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	expected := `ID,Status,
23,completed,
`

	if b.String() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, b.String())
	}
}
