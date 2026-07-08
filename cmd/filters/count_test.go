package filters_test

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
	filters "github.com/prolific-oss/cli/cmd/filters"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func writeCountTemplateFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "filters.json")
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		t.Fatalf("unable to write template file: %s", err)
	}
	return path
}

func TestNewCountCommandSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	templatePath := writeCountTemplateFile(t, `{
		"workspace_id": "ws123",
		"filters": [
			{
				"filter_id": "fluent_languages",
				"selected_values": ["1"]
			}
		]
	}`)

	c.
		EXPECT().
		GetEligibleCount(gomock.Any()).
		DoAndReturn(func(payload client.EligibilityCountPayload) (*client.EligibilityCountResponse, error) {
			if payload.WorkspaceID != "ws123" {
				t.Errorf("expected workspace_id 'ws123'; got '%s'", payload.WorkspaceID)
			}
			if len(payload.Filters) != 1 || payload.Filters[0].FilterID != "fluent_languages" {
				t.Errorf("unexpected filters payload: %+v", payload.Filters)
			}
			return &client.EligibilityCountResponse{Count: 1500}, nil
		}).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewCountCommand(c, w)
	cmd.SetArgs([]string{"-t", templatePath})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	w.Flush()

	expected := "Eligible participants: 1500\n"
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewCountCommandWithWorkspaceOverride(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	templatePath := writeCountTemplateFile(t, `{
		"workspace_id": "old-ws",
		"filters": [
			{
				"filter_id": "approval_rate",
				"selected_range": {
					"lower": 80,
					"upper": 100
				}
			}
		]
	}`)

	c.
		EXPECT().
		GetEligibleCount(gomock.Any()).
		DoAndReturn(func(payload client.EligibilityCountPayload) (*client.EligibilityCountResponse, error) {
			if payload.WorkspaceID != "new-ws" {
				t.Errorf("expected workspace_id override 'new-ws'; got '%s'", payload.WorkspaceID)
			}
			if len(payload.Filters) != 1 || payload.Filters[0].FilterID != "approval_rate" {
				t.Errorf("unexpected filters payload: %+v", payload.Filters)
			}
			return &client.EligibilityCountResponse{Count: 800}, nil
		}).
		Times(1)

	cmd := filters.NewCountCommand(c, os.Stdout)
	cmd.SetArgs([]string{"-t", templatePath, "-w", "new-ws"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestNewCountCommandMissingTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := filters.NewCountCommand(c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	expected := "error: a template file is required, use -t to specify the path"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%v'\n", expected, err)
	}
}

func TestNewCountCommandAllowsEmptyFiltersInTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	templatePath := writeCountTemplateFile(t, `{"workspace_id":"ws123","filters":[]}`)

	c.
		EXPECT().
		GetEligibleCount(client.EligibilityCountPayload{
			WorkspaceID: "ws123",
			Filters:     []model.Filter{},
		}).
		Return(&client.EligibilityCountResponse{Count: 338670}, nil).
		Times(1)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	cmd := filters.NewCountCommand(c, w)
	cmd.SetArgs([]string{"-t", templatePath})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	w.Flush()

	expected := "Eligible participants: 338670\n"
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewCountCommandHandlesAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	templatePath := writeCountTemplateFile(t, `{
		"filters": [
			{
				"filter_id": "fluent_languages",
				"selected_values": ["1"]
			}
		]
	}`)

	errorMessage := "API error: bad request"

	c.
		EXPECT().
		GetEligibleCount(client.EligibilityCountPayload{
			Filters: []model.Filter{
				{FilterID: "fluent_languages", SelectedValues: []string{"1"}},
			},
		}).
		Return(nil, errors.New(errorMessage)).
		Times(1)

	cmd := filters.NewCountCommand(c, os.Stdout)
	cmd.SetArgs([]string{"-t", templatePath})
	err := cmd.Execute()

	expected := fmt.Sprintf("error: %s", errorMessage)
	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%v'\n", expected, err)
	}
}
