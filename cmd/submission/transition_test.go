package submission_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/submission"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewTransitionCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := submission.NewTransitionCommand(c, os.Stdout)

	use := "transition"
	short := "Transition the status of a submission"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestTransitionCommandErrorsIfInvalidAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		TransitionSubmission(gomock.Any(), gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewTransitionCommand(c, writer)
	_ = cmd.Flags().Set("action", "FOOBAR")
	err := cmd.RunE(cmd, []string{testSubmissionID})

	writer.Flush()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expected := `error: invalid action "FOOBAR"`
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestTransitionCommandErrorsIfRejectWithoutMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		TransitionSubmission(gomock.Any(), gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewTransitionCommand(c, writer)
	_ = cmd.Flags().Set("action", "REJECT")
	_ = cmd.Flags().Set("rejection-category", "FAILED_INSTRUCTIONS")
	err := cmd.RunE(cmd, []string{testSubmissionID})

	writer.Flush()

	expected := "error: message is required when rejecting a submission"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestTransitionCommandErrorsIfRejectWithoutCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		TransitionSubmission(gomock.Any(), gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewTransitionCommand(c, writer)
	_ = cmd.Flags().Set("action", "REJECT")
	_ = cmd.Flags().Set("message", "This is a rejection message that needs to be at least 100 characters long to satisfy the API requirement for rejections")
	err := cmd.RunE(cmd, []string{testSubmissionID})

	writer.Flush()

	expected := "error: rejection-category is required when rejecting a submission"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestTransitionCommandTransitionsSubmission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	submissionID := testSubmissionID
	action := "APPROVE"

	r := client.TransitionSubmissionResponse{
		ID:          submissionID,
		Status:      "APPROVED",
		Participant: "participant-123",
		StudyID:     "study-456",
	}

	c.
		EXPECT().
		TransitionSubmission(gomock.Eq(submissionID), gomock.Eq(client.TransitionSubmissionPayload{
			Action: action,
		})).
		Return(&r, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewTransitionCommand(c, writer)
	_ = cmd.Flags().Set("action", action)
	err := cmd.RunE(cmd, []string{submissionID})
	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	if !strings.Contains(actual, submissionID) {
		t.Fatalf("expected output to contain submission ID %s, got\n'%s'\n", submissionID, actual)
	}
	if !strings.Contains(actual, "APPROVED") {
		t.Fatalf("expected output to contain status APPROVED, got\n'%s'\n", actual)
	}
	if !strings.Contains(actual, "participant-123") {
		t.Fatalf("expected output to contain participant ID, got\n'%s'\n", actual)
	}
	if !strings.Contains(actual, "study-456") {
		t.Fatalf("expected output to contain study ID, got\n'%s'\n", actual)
	}
}

func TestTransitionCommandTransitionsSubmissionWithRejection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	submissionID := testSubmissionID
	action := "REJECT"
	message := "Your response did not follow the instructions provided in the study description which is why we are rejecting"
	category := "FAILED_INSTRUCTIONS"

	r := client.TransitionSubmissionResponse{
		ID:          submissionID,
		Status:      "REJECTED",
		Participant: "participant-123",
	}

	c.
		EXPECT().
		TransitionSubmission(gomock.Eq(submissionID), gomock.Eq(client.TransitionSubmissionPayload{
			Action:            action,
			Message:           message,
			RejectionCategory: category,
		})).
		Return(&r, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewTransitionCommand(c, writer)
	_ = cmd.Flags().Set("action", action)
	_ = cmd.Flags().Set("message", message)
	_ = cmd.Flags().Set("rejection-category", category)
	err := cmd.RunE(cmd, []string{submissionID})
	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	if !strings.Contains(actual, submissionID) {
		t.Fatalf("expected output to contain submission ID %s, got\n'%s'\n", submissionID, actual)
	}
	if !strings.Contains(actual, "REJECTED") {
		t.Fatalf("expected output to contain status REJECTED, got\n'%s'\n", actual)
	}
}

func TestTransitionCommandErrorsIfInvalidRejectionCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		TransitionSubmission(gomock.Any(), gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewTransitionCommand(c, writer)
	_ = cmd.Flags().Set("action", "REJECT")
	_ = cmd.Flags().Set("message", "This is a rejection message that needs to be at least 100 characters long to satisfy the API requirement for rejections")
	_ = cmd.Flags().Set("rejection-category", "INVALID_CATEGORY")
	err := cmd.RunE(cmd, []string{testSubmissionID})

	writer.Flush()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expected := `error: invalid rejection category "INVALID_CATEGORY"`
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestTransitionCommandErrorsIfCompleteWithoutCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		TransitionSubmission(gomock.Any(), gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewTransitionCommand(c, writer)
	_ = cmd.Flags().Set("action", "COMPLETE")
	err := cmd.RunE(cmd, []string{testSubmissionID})

	writer.Flush()

	expected := "error: completion-code is required when completing a submission"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestTransitionCommandTransitionsSubmissionWithComplete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	submissionID := testSubmissionID
	action := "COMPLETE"
	completionCode := "MY_CODE"

	r := client.TransitionSubmissionResponse{
		ID:          submissionID,
		Status:      "APPROVED",
		Participant: "participant-123",
	}

	c.
		EXPECT().
		TransitionSubmission(gomock.Eq(submissionID), gomock.Eq(client.TransitionSubmissionPayload{
			Action:         action,
			CompletionCode: completionCode,
		})).
		Return(&r, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewTransitionCommand(c, writer)
	_ = cmd.Flags().Set("action", action)
	_ = cmd.Flags().Set("completion-code", completionCode)
	err := cmd.RunE(cmd, []string{submissionID})
	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	if !strings.Contains(actual, submissionID) {
		t.Fatalf("expected output to contain submission ID %s, got\n'%s'\n", submissionID, actual)
	}
	if !strings.Contains(actual, "APPROVED") {
		t.Fatalf("expected output to contain status, got\n'%s'\n", actual)
	}
}

func TestTransitionCommandTransitionsSubmissionWithDynamicPayment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	submissionID := testSubmissionID
	action := "COMPLETE"
	completionCode := "MY_CODE"

	r := client.TransitionSubmissionResponse{
		ID:          submissionID,
		Status:      "APPROVED",
		Participant: "participant-123",
	}

	c.
		EXPECT().
		TransitionSubmission(gomock.Eq(submissionID), gomock.Eq(client.TransitionSubmissionPayload{
			Action:         action,
			CompletionCode: completionCode,
			CompletionCodeData: &client.CompletionCodeData{
				PercentageOfReward:   50,
				MessageToParticipant: "Good job",
			},
		})).
		Return(&r, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewTransitionCommand(c, writer)
	_ = cmd.Flags().Set("action", action)
	_ = cmd.Flags().Set("completion-code", completionCode)
	_ = cmd.Flags().Set("percentage-of-reward", "50")
	_ = cmd.Flags().Set("message-to-participant", "Good job")
	err := cmd.RunE(cmd, []string{submissionID})
	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	if !strings.Contains(actual, submissionID) {
		t.Fatalf("expected output to contain submission ID %s, got\n'%s'\n", submissionID, actual)
	}
}

func TestTransitionCommandReturnsErrorIfTransitionFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "Unable to transition submission"
	submissionID := testSubmissionID

	c.
		EXPECT().
		TransitionSubmission(gomock.Eq(submissionID), gomock.Eq(client.TransitionSubmissionPayload{
			Action: "APPROVE",
		})).
		Return(nil, errors.New(errorMessage)).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewTransitionCommand(c, writer)
	_ = cmd.Flags().Set("action", "APPROVE")
	err := cmd.RunE(cmd, []string{submissionID})

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
