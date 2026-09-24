package audience_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/audience"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func mustWriteTempTemplate(t *testing.T, content string) string {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	return f.Name()
}

func TestNewBreakdownCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := audience.NewBreakdownCommand(c, os.Stdout)

	use := "breakdown"
	short := "Count eligible participants broken down by a filter"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestBreakdownIsNestedUnderAudience(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := audience.NewAudienceCommand(c, os.Stdout)

	sub, _, err := cmd.Find([]string{"breakdown"})
	if err != nil {
		t.Fatalf("expected breakdown to be a subcommand of audience: %s", err)
	}

	if sub.Use != "breakdown" {
		t.Fatalf("expected use: breakdown; got %s", sub.Use)
	}
}

func TestBreakdownCommandRendersBreakdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	expectedPayload := client.FilterBreakdownPayload{
		Filters: []model.Filter{
			{FilterID: "participant_group_allowlist", SelectedValues: []string{"group-id"}},
		},
		BreakdownFilter: model.Filter{FilterID: "handedness", SelectedValues: []string{"0", "1"}},
		WorkspaceID:     "ws-id",
	}

	c.
		EXPECT().
		GetFilterBreakdown(gomock.Eq(expectedPayload)).
		Return(&client.FilterBreakdownResponse{Breakdown: map[string]int{"0": 4, "1": 3, "N/A": 5}}, nil).
		Times(1)

	templatePath := mustWriteTempTemplate(t, `{
		"filters": [{"filter_id": "participant_group_allowlist", "selected_values": ["group-id"]}],
		"breakdown_filter": {"filter_id": "handedness", "selected_values": ["0", "1"]}
	}`)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := audience.NewBreakdownCommand(c, writer)
	_ = cmd.Flags().Set("template-path", templatePath)
	_ = cmd.Flags().Set("workspace", "ws-id")

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	writer.Flush()

	expected := "VALUE   COUNT\n0       4\n1       3\nN/A     5\n"
	if b.String() != expected {
		t.Fatalf("expected %q, got %q", expected, b.String())
	}
}

func TestBreakdownCommandSendsEmptyFiltersNotNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	expectedPayload := client.FilterBreakdownPayload{
		Filters:         []model.Filter{},
		BreakdownFilter: model.Filter{FilterID: "handedness"},
		WorkspaceID:     "ws-id",
	}

	c.
		EXPECT().
		GetFilterBreakdown(gomock.Eq(expectedPayload)).
		Return(&client.FilterBreakdownResponse{Breakdown: map[string]int{}}, nil).
		Times(1)

	// No "filters" key at all — must not leave Filters nil, since the API
	// rejects a null filters field.
	templatePath := mustWriteTempTemplate(t, `{"breakdown_filter": {"filter_id": "handedness"}}`)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := audience.NewBreakdownCommand(c, writer)
	_ = cmd.Flags().Set("template-path", templatePath)
	_ = cmd.Flags().Set("workspace", "ws-id")

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestBreakdownCommandValidatesInput(t *testing.T) {
	tests := []struct {
		name          string
		templateJSON  string // empty means -t/--template-path is left unset
		workspaceID   string
		expectedError string
	}{
		{
			name:          "missing template path",
			workspaceID:   "ws-id",
			expectedError: "error: a filter template is required, use -t/--template-path",
		},
		{
			name:          "missing workspace",
			templateJSON:  `{"breakdown_filter": {"filter_id": "handedness"}}`,
			expectedError: "error: workspace ID is required",
		},
		{
			name:          "missing breakdown filter",
			templateJSON:  `{"filters": []}`,
			workspaceID:   "ws-id",
			expectedError: "error: template must include a breakdown_filter with a filter_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)
			cmd := audience.NewBreakdownCommand(c, writer)

			if tt.templateJSON != "" {
				_ = cmd.Flags().Set("template-path", mustWriteTempTemplate(t, tt.templateJSON))
			}
			_ = cmd.Flags().Set("workspace", tt.workspaceID)

			err := cmd.RunE(cmd, nil)
			if err == nil || err.Error() != tt.expectedError {
				t.Fatalf("expected %q, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestBreakdownCommandHandlesFailureToReadConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := audience.NewBreakdownCommand(c, writer)
	_ = cmd.Flags().Set("template-path", "broken-path.json")
	_ = cmd.Flags().Set("workspace", "ws-id")

	err := cmd.RunE(cmd, nil)
	writer.Flush()

	expected := "error: open broken-path.json: no such file or directory"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected %q, got %v", expected, err)
	}
}

func TestBreakdownCommandHandlesAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		GetFilterBreakdown(gomock.Any()).
		Return(nil, errors.New("boom")).
		Times(1)

	templatePath := mustWriteTempTemplate(t, `{"filters": [], "breakdown_filter": {"filter_id": "handedness"}}`)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := audience.NewBreakdownCommand(c, writer)
	_ = cmd.Flags().Set("template-path", templatePath)
	_ = cmd.Flags().Set("workspace", "ws-id")

	err := cmd.RunE(cmd, nil)

	expected := "error: boom"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected %q, got %v", expected, err)
	}
}

func TestRenderBreakdown(t *testing.T) {
	breakdown := map[string]int{"1": 3, "0": 4, "N/A": 5}

	expected := "VALUE   COUNT\n0       4\n1       3\nN/A     5\n"
	actual := audience.RenderBreakdown(breakdown)

	if actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}
