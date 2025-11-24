package study_test

import (
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
