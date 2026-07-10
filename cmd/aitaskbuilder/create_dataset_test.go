package aitaskbuilder_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strings"
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

	workspaceID := workspaceID123
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
		WorkspaceID:         workspaceID123,
		SchemaVersion:       3,
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
	_ = cmd.Flags().Set("workspace-id", workspaceID123)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `ID: dataset-456
Name: Test Dataset
Created At: 2024-01-15T10:30:00Z
Created By: user-789
Status: READY
Total Datapoint Count: 0
Workspace ID: workspace-123
Schema Version: 3
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewCreateDatasetCommandWithSchema(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	strict := true
	workspaceID := workspaceID123
	payload := client.CreateAITaskBuilderDatasetPayload{
		Name: "Test Dataset",
		Schema: &client.DatasetSchema{
			Strict: &strict,
			Fields: map[string]client.DatasetSchemaField{
				"question": {Type: "text", Label: "Question"},
				"group":    {Type: "task_group_id"},
			},
		},
	}

	response := client.CreateAITaskBuilderDatasetResponse{
		ID:            "dataset-456",
		Name:          "Test Dataset",
		CreatedAt:     "2024-01-15T10:30:00Z",
		CreatedBy:     "user-789",
		Status:        "READY",
		WorkspaceID:   workspaceID123,
		SchemaVersion: 4,
	}

	c.
		EXPECT().
		CreateAITaskBuilderDataset(gomock.Eq(workspaceID), gomock.Eq(payload)).
		Return(&response, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewCreateDatasetCommand(c, writer)
	_ = cmd.Flags().Set("name", "Test Dataset")
	_ = cmd.Flags().Set("workspace-id", workspaceID123)
	_ = cmd.Flags().Set("strict", "true")
	_ = cmd.Flags().Set("schema", `{"fields":{"question":{"type":"text","label":"Question"},"group":{"type":"task_group_id"}}}`)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("expected no error; got %v", err)
	}

	writer.Flush()

	expected := `ID: dataset-456
Name: Test Dataset
Created At: 2024-01-15T10:30:00Z
Created By: user-789
Status: READY
Total Datapoint Count: 0
Workspace ID: workspace-123
Schema Version: 4
Strict: true
Schema Fields: 2
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewCreateDatasetCommandWithSchemaDefaultsStrictFalse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	strict := false
	workspaceID := workspaceID123
	payload := client.CreateAITaskBuilderDatasetPayload{
		Name: "Test Dataset",
		Schema: &client.DatasetSchema{
			Strict: &strict,
			Fields: map[string]client.DatasetSchemaField{
				"question": {Type: "text"},
			},
		},
	}

	response := client.CreateAITaskBuilderDatasetResponse{
		ID:            "dataset-456",
		Name:          "Test Dataset",
		CreatedAt:     "2024-01-15T10:30:00Z",
		CreatedBy:     "user-789",
		Status:        "READY",
		WorkspaceID:   workspaceID123,
		SchemaVersion: 4,
	}

	c.
		EXPECT().
		CreateAITaskBuilderDataset(gomock.Eq(workspaceID), gomock.Eq(payload)).
		Return(&response, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := aitaskbuilder.NewCreateDatasetCommand(c, writer)
	_ = cmd.Flags().Set("name", "Test Dataset")
	_ = cmd.Flags().Set("workspace-id", workspaceID123)
	_ = cmd.Flags().Set("schema", `{"fields":{"question":{"type":"text"}}}`)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("expected no error; got %v", err)
	}

	writer.Flush()

	expected := `ID: dataset-456
Name: Test Dataset
Created At: 2024-01-15T10:30:00Z
Created By: user-789
Status: READY
Total Datapoint Count: 0
Workspace ID: workspace-123
Schema Version: 4
Strict: false
Schema Fields: 1
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewCreateDatasetCommandStrictWithoutSchema(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// No API call expected: validation fails before the client is used.

	cmd := aitaskbuilder.NewCreateDatasetCommand(c, os.Stdout)
	_ = cmd.Flags().Set("name", "Test Dataset")
	_ = cmd.Flags().Set("workspace-id", workspaceID123)
	_ = cmd.Flags().Set("strict", "true")
	err := cmd.RunE(cmd, nil)

	expected := aitaskbuilder.ErrStrictRequiresSchema
	if err == nil || err.Error() != expected {
		t.Fatalf("expected error '%s'; got '%v'", expected, err)
	}
}

func TestNewCreateDatasetCommandExecuteShowsUsageForRuntimeErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	writer := bufio.NewWriter(&stdout)

	cmd := aitaskbuilder.NewCreateDatasetCommand(c, writer)
	cmd.SetOut(&stderr)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--name", "Test Dataset", "--workspace-id", workspaceID123, "--strict"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}

	if err.Error() != aitaskbuilder.ErrStrictRequiresSchema {
		t.Fatalf("expected error %q; got %q", aitaskbuilder.ErrStrictRequiresSchema, err.Error())
	}

	writer.Flush()

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output; got %q", stdout.String())
	}

	if !strings.Contains(stderr.String(), "Usage:") || !strings.Contains(stderr.String(), "Error: "+aitaskbuilder.ErrStrictRequiresSchema) {
		t.Fatalf("expected Cobra error and usage output; got %q", stderr.String())
	}
}

func TestNewCreateDatasetCommandStrictSetInBoth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// No API call expected: validation fails before the client is used.

	cmd := aitaskbuilder.NewCreateDatasetCommand(c, os.Stdout)
	_ = cmd.Flags().Set("name", "Test Dataset")
	_ = cmd.Flags().Set("workspace-id", workspaceID123)
	_ = cmd.Flags().Set("strict", "true")
	_ = cmd.Flags().Set("schema", `{"strict":true,"fields":{"question":{"type":"text"}}}`)
	err := cmd.RunE(cmd, nil)

	expected := aitaskbuilder.ErrSchemaStrictSetInBoth
	if err == nil || err.Error() != expected {
		t.Fatalf("expected error '%s'; got '%v'", expected, err)
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

	c.
		EXPECT().
		CreateAITaskBuilderDataset(gomock.Eq(workspaceID), gomock.Eq(payload)).
		Return(nil, errors.New(workspaceNotFoundError)).
		AnyTimes()

	cmd := aitaskbuilder.NewCreateDatasetCommand(c, os.Stdout)
	_ = cmd.Flags().Set("name", "Test Dataset")
	_ = cmd.Flags().Set("workspace-id", "invalid-workspace")
	err := cmd.RunE(cmd, nil)

	expected := workspaceNotFoundError

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewCreateDatasetCommandRequiresName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := aitaskbuilder.NewCreateDatasetCommand(c, os.Stdout)
	_ = cmd.Flags().Set("workspace-id", workspaceID123)
	err := cmd.RunE(cmd, nil)

	if err == nil {
		t.Fatal("expected error when name is missing")
	}

	if !cmd.Flags().Changed("name") {
		expected := aitaskbuilder.ErrNameRequired
		if err.Error() != expected {
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
		if err.Error() != expected {
			t.Fatalf("expected error to contain '%s', got '%s'", expected, err.Error())
		}
	}
}
