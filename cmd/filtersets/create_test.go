package filtersets_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/filtersets"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewCreateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := filtersets.NewCreateCommand("create", c, os.Stdout)

	use := "create"
	short := "Create a filter set"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func writeTemplateFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "filter-set.json")
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		t.Fatalf("unable to write template file: %s", err)
	}
	return path
}

func TestNewCreateCommandSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	templatePath := writeTemplateFile(t, `{
		"workspace_id": "ws123",
		"name": "Test filter set",
		"filters": [
			{
				"filter_id": "handedness",
				"selected_values": ["ambidextrous"]
			}
		]
	}`)

	response := client.CreateFilterSetResponse{
		FilterSet: model.FilterSet{
			ID:                       "fs-001",
			Name:                     "Test filter set",
			WorkspaceID:              "ws123",
			EligibleParticipantCount: 500,
		},
	}

	c.
		EXPECT().
		CreateFilterSet(gomock.Any()).
		Return(&response, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := filtersets.NewCreateCommand("create", c, writer)
	cmd.SetArgs([]string{"-t", templatePath})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	writer.Flush()

	expected := "Created filter set: fs-001 (eligible participants: 500)\n"
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewCreateCommandWithNameOverride(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	templatePath := writeTemplateFile(t, `{
		"name": "Original name",
		"filters": [
			{
				"filter_id": "handedness",
				"selected_values": ["left"]
			}
		]
	}`)

	response := client.CreateFilterSetResponse{
		FilterSet: model.FilterSet{
			ID:                       "fs-002",
			Name:                     "Overridden name",
			EligibleParticipantCount: 100,
		},
	}

	c.
		EXPECT().
		CreateFilterSet(gomock.Any()).
		DoAndReturn(func(fs model.FilterSet) (*client.CreateFilterSetResponse, error) {
			if fs.Name != "Overridden name" {
				t.Fatalf("expected name 'Overridden name'; got '%s'", fs.Name)
			}
			return &response, nil
		}).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := filtersets.NewCreateCommand("create", c, writer)
	cmd.SetArgs([]string{"-t", templatePath, "-N", "Overridden name"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	writer.Flush()

	expected := "Created filter set: fs-002 (eligible participants: 100)\n"
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewCreateCommandMissingTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := filtersets.NewCreateCommand("create", c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	expected := "error: a template file is required, use -t to specify the path"
	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewCreateCommandHandlesAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	templatePath := writeTemplateFile(t, `{
		"name": "Test",
		"filters": [
			{
				"filter_id": "age",
				"selected_values": ["19-22"]
			}
		]
	}`)

	errorMessage := "API error: bad request"

	c.
		EXPECT().
		CreateFilterSet(gomock.Any()).
		Return(nil, errors.New(errorMessage)).
		Times(1)

	cmd := filtersets.NewCreateCommand("create", c, os.Stdout)
	cmd.SetArgs([]string{"-t", templatePath})
	err := cmd.Execute()

	expected := fmt.Sprintf("error: %s", errorMessage)
	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%v'\n", expected, err)
	}
}
