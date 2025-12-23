package study_test

import (
	"bufio"
	"bytes"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/study"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewViewCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := study.NewListCommand("list", client, os.Stdout)

	use := "list"
	short := "List all of your studies"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestListCommandPassesBothProjectIDAndStatusToGetStudies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock_client.NewMockAPI(ctrl)

	projectID := "6655b8281cc82a88996f0bbb"
	status := model.StatusUnpublished

	studyResponse := client.ListStudiesResponse{
		Results: []model.Study{},
	}

	mockClient.
		EXPECT().
		GetStudies(gomock.Eq(status), gomock.Eq(projectID)).
		Return(&studyResponse, nil).
		Times(1)

	cmd := study.NewListCommand("list", mockClient, os.Stdout)

	_ = cmd.Flags().Set("project", projectID)
	_ = cmd.Flags().Set("status", status)
	_ = cmd.Flags().Set("non-interactive", "true")

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestListCommandFiltersByUnderpaying(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	underpayingStudy := model.Study{
		ID:                      "1234",
		Name:                    "An underpaying study",
		InternalName:            "Underpaying study",
		Desc:                    "An underpaying study.",
		Status:                  model.StatusActive,
		ExternalStudyURL:        "https://eggs-experriment.com?participant={{%PROLIFIC_PID%}}",
		TotalAvailablePlaces:    10,
		EstimatedCompletionTime: 10,
		MaximumAllowedTime:      10,
		Reward:                  400,
		DeviceCompatibility:     []string{"desktop", "tablet", "mobile"},
		IsUnderpaying:           true,
	}
	payingFineStudy := model.Study{
		ID:                      "5678",
		Name:                    "Normal paying study",
		InternalName:            "Normal pay",
		Desc:                    "A normal paying study.",
		Status:                  model.StatusActive,
		ExternalStudyURL:        "https://eggs-experriment.com?participant={{%PROLIFIC_PID%}}",
		TotalAvailablePlaces:    10,
		EstimatedCompletionTime: 10,
		MaximumAllowedTime:      10,
		Reward:                  400,
		DeviceCompatibility:     []string{"desktop", "tablet", "mobile"},
		IsUnderpaying:           false,
	}
	studyResponse := client.ListStudiesResponse{
		Results: []model.Study{underpayingStudy, payingFineStudy},
	}

	c.
		EXPECT().
		GetStudies(gomock.Eq(model.StatusAll), gomock.Eq("")).
		Return(&studyResponse, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewListCommand("list", c, writer)

	_ = cmd.Flags().Set("underpaying", "true")
	_ = cmd.Flags().Set("csv", "true")

	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}
	writer.Flush()

	expected := `ID,Name,Status,
1234,An underpaying study,active,
`

	if b.String() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, b.String())
	}
}
