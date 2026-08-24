package eligibilitycount_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/eligibilitycount"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewCountCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := eligibilitycount.NewCountCommand(c, os.Stdout)

	use := "eligibility-count"
	short := "Count participants matching a set of filters"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestCountCommandRendersCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	expectedPayload := client.EligibilityCountPayload{
		Filters: []model.Filter{
			{FilterID: "age", SelectedRange: &model.FilterRange{Lower: float64(18), Upper: float64(65)}},
		},
		WorkspaceID: "ws-id",
	}

	c.
		EXPECT().
		GetEligibilityCount(gomock.Eq(expectedPayload)).
		Return(&client.EligibilityCountResponse{Count: 142}, nil).
		Times(1)

	f, err := os.CreateTemp(t.TempDir(), "*.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	if _, err := f.WriteString(`{"filters": [{"filter_id": "age", "selected_range": {"lower": 18, "upper": 65}}]}`); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := eligibilitycount.NewCountCommand(c, writer)
	_ = cmd.Flags().Set("template-path", f.Name())
	_ = cmd.Flags().Set("workspace", "ws-id")

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	writer.Flush()

	expected := "Eligible participants: 142\n"
	if b.String() != expected {
		t.Fatalf("expected %q, got %q", expected, b.String())
	}
}

func TestCountCommandFlagsSubTwentyFiveCounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		GetEligibilityCount(gomock.Any()).
		Return(&client.EligibilityCountResponse{Count: 0}, nil).
		Times(1)

	f, err := os.CreateTemp(t.TempDir(), "*.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	if _, err := f.WriteString(`{"filters": []}`); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := eligibilitycount.NewCountCommand(c, writer)
	_ = cmd.Flags().Set("template-path", f.Name())
	_ = cmd.Flags().Set("workspace", "ws-id")

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	writer.Flush()

	if !strings.Contains(b.String(), "fewer than 25") {
		t.Fatalf("expected sub-25 caveat in output, got %q", b.String())
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
		Return(&client.EligibilityCountResponse{Count: 30}, nil).
		Times(1)

	f, err := os.CreateTemp(t.TempDir(), "*.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	// No "filters" key at all — must not leave Filters nil, since the API
	// rejects a null filters field.
	if _, err := f.WriteString(`{}`); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := eligibilitycount.NewCountCommand(c, writer)
	_ = cmd.Flags().Set("template-path", f.Name())
	_ = cmd.Flags().Set("workspace", "ws-id")

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestCountCommandRequiresTemplatePath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := eligibilitycount.NewCountCommand(c, writer)
	_ = cmd.Flags().Set("workspace", "ws-id")

	err := cmd.RunE(cmd, nil)

	expected := "error: a filter template is required, use -t/--template-path"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected %q, got %v", expected, err)
	}
}

func TestCountCommandRequiresWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	f, err := os.CreateTemp(t.TempDir(), "*.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	if _, err := f.WriteString(`{"filters": []}`); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := eligibilitycount.NewCountCommand(c, writer)
	_ = cmd.Flags().Set("template-path", f.Name())
	_ = cmd.Flags().Set("workspace", "")

	err = cmd.RunE(cmd, nil)

	expected := "error: workspace ID is required"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected %q, got %v", expected, err)
	}
}

func TestCountCommandHandlesFailureToReadConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := eligibilitycount.NewCountCommand(c, writer)
	_ = cmd.Flags().Set("template-path", "broken-path.json")
	_ = cmd.Flags().Set("workspace", "ws-id")

	err := cmd.RunE(cmd, nil)
	writer.Flush()

	expected := "error: open broken-path.json: no such file or directory"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected %q, got %v", expected, err)
	}
}

func TestCountCommandHandlesFailureToUnmarshalConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	f, err := os.CreateTemp(t.TempDir(), "*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	// weightings must be a map; a scalar causes Unmarshal to fail
	if _, err = f.WriteString("filters:\n  - filter_id: age\n    weightings: not-a-map\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := eligibilitycount.NewCountCommand(c, writer)
	_ = cmd.Flags().Set("template-path", f.Name())
	_ = cmd.Flags().Set("workspace", "ws-id")

	err = cmd.RunE(cmd, nil)
	writer.Flush()

	if err == nil || !strings.Contains(err.Error(), "unable to map") {
		t.Fatalf("expected unmarshal error, got %v", err)
	}
}

func TestCountCommandHandlesAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		GetEligibilityCount(gomock.Any()).
		Return(nil, errors.New("boom")).
		Times(1)

	f, err := os.CreateTemp(t.TempDir(), "*.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	if _, err := f.WriteString(`{"filters": []}`); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := eligibilitycount.NewCountCommand(c, writer)
	_ = cmd.Flags().Set("template-path", f.Name())
	_ = cmd.Flags().Set("workspace", "ws-id")

	err = cmd.RunE(cmd, nil)

	expected := "error: boom"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected %q, got %v", expected, err)
	}
}
