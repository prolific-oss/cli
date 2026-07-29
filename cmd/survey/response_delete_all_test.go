package survey_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/survey"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewResponseDeleteAllCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := survey.NewResponseDeleteAllCommand("delete-all", c, os.Stdout)

	if cmd.Use != "delete-all <survey_id>" {
		t.Fatalf("expected use: delete-all <survey_id>; got %s", cmd.Use)
	}

	if cmd.Short != "Delete all responses for a survey" {
		t.Fatalf("expected short: Delete all responses for a survey; got %s", cmd.Short)
	}
}

func TestDeleteAllSurveyResponses_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		DeleteAllSurveyResponses(gomock.Eq(testSurveyID)).
		Return(nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := survey.NewResponseDeleteAllCommand("delete-all", c, writer)
	cmd.SetArgs([]string{testSurveyID})
	err := cmd.Execute()

	writer.Flush()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Deleted all responses for survey: " + testSurveyID + "\n"
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestDeleteAllSurveyResponses_APIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		DeleteAllSurveyResponses(gomock.Eq(testSurveyID)).
		Return(errors.New("not found")).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := survey.NewResponseDeleteAllCommand("delete-all", c, writer)
	cmd.SetArgs([]string{testSurveyID})
	err := cmd.Execute()

	writer.Flush()

	expectedErr := "error: not found"
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("expected error '%s'; got '%v'", expectedErr, err)
	}
}
