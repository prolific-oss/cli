package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/aitaskbuilder"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewCreateDatasetCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewCreateDatasetCommand(c, os.Stdout)

	use := "create"
	short := "Create a Dataset"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewCreateDatasetCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := "workspace-123"
	payload := client.CreateAITaskBuilderDatasetPayload{
		Name: "Test Dataset",
	}

	response := client.CreateAITaskBuilderDatasetResponse{
		ID:                  "dataset-456",
		Name:                "Test Dataset",
		CreatedAt:           "2024-01-15T10:30:00Z",
		CreatedBy:           "user-789",
		Status:              "READY",
		TotalDatapointCount: 0,
		WorkspaceID:         "workspace-123",
	}

	c.
		EXPECT().
		CreateAITaskBuilderDataset(gomock.Eq(workspaceID), gomock.Eq(payload)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewCreateDatasetCommand(c, writer)
	_ = cmd.Flags().Set("name", "Test Dataset")
	_ = cmd.Flags().Set("workspace-id", "workspace-123")
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `ID: dataset-456
Name: Test Dataset
Created At: 2024-01-15T10:30:00Z
Created By: user-789
Status: READY
Total Datapoint Count: 0
Workspace ID: workspace-123
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewCreateDatasetCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := "invalid-workspace"
	payload := client.CreateAITaskBuilderDatasetPayload{
		Name: "Test Dataset",
	}

	errorMessage := "workspace not found"

	c.
		EXPECT().
		CreateAITaskBuilderDataset(gomock.Eq(workspaceID), gomock.Eq(payload)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := aitaskbuilder.NewCreateDatasetCommand(c, os.Stdout)
	_ = cmd.Flags().Set("name", "Test Dataset")
	_ = cmd.Flags().Set("workspace-id", "invalid-workspace")
	err := cmd.RunE(cmd, nil)

	expected := errorMessage

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewCreateDatasetCommandRequiresName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewCreateDatasetCommand(c, os.Stdout)
	_ = cmd.Flags().Set("workspace-id", "workspace-123")
	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Fatal("expected error when name is missing")
	}

	if !cmd.Flags().Changed("name") {
		expected := aitaskbuilder.ErrNameRequired
		if err.Error() != ""+expected {
			t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
		}
	}
}

func TestNewCreateDatasetCommandRequiresWorkspaceID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewCreateDatasetCommand(c, os.Stdout)
	_ = cmd.Flags().Set("name", "Test Dataset")
	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Fatal("expected error when workspace-id is missing")
	}

	if !cmd.Flags().Changed("workspace-id") {
		expected := aitaskbuilder.ErrWorkspaceIDRequired
		if err.Error() != ""+expected {
			t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
		}
	}
}
