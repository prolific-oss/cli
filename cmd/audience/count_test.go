package audience_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/audience"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
	"github.com/spf13/cobra"
)

func TestNewCountCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := audience.NewCountCommand(c, os.Stdout)

	use := "count"
	short := "Count participants matching a set of filters"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestCountIsNestedUnderAudience(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := audience.NewAudienceCommand(c, os.Stdout)

	sub, _, err := cmd.Find([]string{"count"})
	if err != nil {
		t.Fatalf("expected count to be a subcommand of audience: %s", err)
	}

	if sub.Use != "count" {
		t.Fatalf("expected use: count; got %s", sub.Use)
	}
}

func TestCountCommandRendersCountFromTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// Asserted as marshalled JSON rather than a struct, so that the pointer
	// in SelectedRange and the any-typed bounds compare by wire shape.
	var actualPayload client.EligibilityCountPayload

	c.
		EXPECT().
		GetEligibilityCount(gomock.Any()).
		DoAndReturn(func(payload client.EligibilityCountPayload) (*client.EligibilityCountResponse, error) {
			actualPayload = payload
			return &client.EligibilityCountResponse{Count: 1234}, nil
		}).
		Times(1)

	templatePath := mustWriteTempTemplate(t, `{
		"filters": [{"filter_id": "age", "selected_range": {"lower": 30, "upper": 40}}]
	}`)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := audience.NewCountCommand(c, writer)
	_ = cmd.Flags().Set("template-path", templatePath)
	_ = cmd.Flags().Set("workspace", "ws-id")

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	writer.Flush()

	expectedPayload := `{"filters":[{"filter_id":"age","selected_range":{"lower":30,"upper":40}}],"workspace_id":"ws-id"}`
	sentPayload, err := json.Marshal(actualPayload)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if string(sentPayload) != expectedPayload {
		t.Fatalf("expected payload %s, got %s", expectedPayload, sentPayload)
	}

	expected := "Eligible participants: 1234\n"
	if b.String() != expected {
		t.Fatalf("expected %q, got %q", expected, b.String())
	}
}

func TestCountCommandRendersCountFromFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	expectedPayload := client.EligibilityCountPayload{
		Filters:     []model.Filter{{FilterID: "age", SelectedRange: &model.FilterRange{Lower: float64(30), Upper: float64(40)}}},
		WorkspaceID: "ws-id",
	}

	c.
		EXPECT().
		GetEligibilityCount(gomock.Eq(expectedPayload)).
		Return(&client.EligibilityCountResponse{Count: 1234}, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := audience.NewCountCommand(c, writer)
	_ = cmd.Flags().Set("filters", `[{"filter_id":"age","selected_range":{"lower":30,"upper":40}}]`)
	_ = cmd.Flags().Set("workspace", "ws-id")

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	writer.Flush()

	expected := "Eligible participants: 1234\n"
	if b.String() != expected {
		t.Fatalf("expected %q, got %q", expected, b.String())
	}
}

func TestCountCommandHandlesMalformedFlagJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := audience.NewCountCommand(c, writer)
	_ = cmd.Flags().Set("filters", `not-json`)
	_ = cmd.Flags().Set("workspace", "ws-id")

	err := cmd.RunE(cmd, nil)
	if err == nil || !strings.HasPrefix(err.Error(), "error: unable to parse --filters as JSON:") {
		t.Fatalf("expected a --filters parse error, got %v", err)
	}
}

func TestCountCommandRendersJSON(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{
			name:     "exact count",
			count:    1234,
			expected: `{"count":1234}` + "\n",
		},
		{
			name:     "zero count",
			count:    0,
			expected: `{"count":0}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			c.
				EXPECT().
				GetEligibilityCount(gomock.Any()).
				Return(&client.EligibilityCountResponse{Count: tt.count}, nil).
				Times(1)

			templatePath := mustWriteTempTemplate(t, `{"filters": []}`)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := audience.NewCountCommand(c, writer)
			_ = cmd.Flags().Set("template-path", templatePath)
			_ = cmd.Flags().Set("workspace", "ws-id")
			_ = cmd.Flags().Set("json", "true")

			if err := cmd.RunE(cmd, nil); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			writer.Flush()

			if b.String() != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, b.String())
			}
		})
	}
}

func TestCountCommandSendsEmptyFiltersNotNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	expectedPayload := client.EligibilityCountPayload{
		Filters:     []model.Filter{},
		WorkspaceID: "ws-id",
	}

	c.
		EXPECT().
		GetEligibilityCount(gomock.Eq(expectedPayload)).
		Return(&client.EligibilityCountResponse{Count: 0}, nil).
		Times(1)

	// No "filters" key at all — must not leave Filters nil, since the API
	// rejects a null filters field.
	templatePath := mustWriteTempTemplate(t, `{}`)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := audience.NewCountCommand(c, writer)
	_ = cmd.Flags().Set("template-path", templatePath)
	_ = cmd.Flags().Set("workspace", "ws-id")

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestCountCommandValidatesInput(t *testing.T) {
	tests := []struct {
		name          string
		templateJSON  string // empty means -t/--template-path is left unset
		filtersJSON   string
		workspaceID   string
		expectedError string
	}{
		{
			name:          "no template or filters",
			workspaceID:   "ws-id",
			expectedError: "error: a filter template or --filters is required, use -t/--template-path or --filters",
		},
		{
			name:          "both template and filters",
			templateJSON:  `{"filters": []}`,
			filtersJSON:   `[]`,
			workspaceID:   "ws-id",
			expectedError: "error: use only one of -t/--template-path or --filters",
		},
		{
			name:          "missing workspace with template",
			templateJSON:  `{"filters": []}`,
			expectedError: "error: workspace ID is required",
		},
		{
			name:          "missing workspace with filters",
			filtersJSON:   `[]`,
			expectedError: "error: workspace ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)
			cmd := audience.NewCountCommand(c, writer)

			if tt.templateJSON != "" {
				_ = cmd.Flags().Set("template-path", mustWriteTempTemplate(t, tt.templateJSON))
			}
			_ = cmd.Flags().Set("filters", tt.filtersJSON)
			_ = cmd.Flags().Set("workspace", tt.workspaceID)

			err := cmd.RunE(cmd, nil)
			if err == nil || err.Error() != tt.expectedError {
				t.Fatalf("expected %q, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestCountCommandHandlesFailureToReadConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := audience.NewCountCommand(c, writer)
	_ = cmd.Flags().Set("template-path", "broken-path.json")
	_ = cmd.Flags().Set("workspace", "ws-id")

	err := cmd.RunE(cmd, nil)
	writer.Flush()

	expected := "error: open broken-path.json: no such file or directory"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected %q, got %v", expected, err)
	}
}

func TestCountCommandHandlesAPIError(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, c *mock_client.MockAPI, cmd *cobra.Command)
	}{
		{
			name: "eligibility count fails",
			setup: func(t *testing.T, c *mock_client.MockAPI, cmd *cobra.Command) {
				c.EXPECT().GetEligibilityCount(gomock.Any()).Return(nil, errors.New("boom")).Times(1)
				_ = cmd.Flags().Set("template-path", mustWriteTempTemplate(t, `{"filters": []}`))
				_ = cmd.Flags().Set("workspace", "ws-id")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)
			cmd := audience.NewCountCommand(c, writer)

			tt.setup(t, c, cmd)

			err := cmd.RunE(cmd, nil)

			expected := "error: boom"
			if err == nil || err.Error() != expected {
				t.Fatalf("expected %q, got %v", expected, err)
			}
		})
	}
}

func TestRenderCount(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		asJSON   bool
		expected string
	}{
		{
			name:     "exact count",
			count:    1234,
			expected: "Eligible participants: 1234",
		},
		{
			name:     "zero count",
			count:    0,
			expected: "Eligible participants: 0",
		},
		{
			name:     "json exact count",
			count:    1234,
			asJSON:   true,
			expected: `{"count":1234}`,
		},
		{
			name:     "json zero count",
			count:    0,
			asJSON:   true,
			expected: `{"count":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := audience.RenderCount(tt.count, tt.asJSON)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if actual != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}
