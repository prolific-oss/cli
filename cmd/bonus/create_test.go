package bonus_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/bonus"
	"github.com/prolific-oss/cli/mock_client"
)

// setupCreateMock creates a mock API with a CreateBonusPayments expectation
// and returns the mock, a buffered writer, and a buffer for output capture.
func setupCreateMock(t *testing.T, response *client.CreateBonusPaymentsResponse) (*mock_client.MockAPI, *bufio.Writer, *bytes.Buffer) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		CreateBonusPayments(gomock.Any()).
		Return(response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	return c, writer, &b
}

var defaultCreateResponse = &client.CreateBonusPaymentsResponse{
	ID:          "bonus-abc-123",
	Study:       "study-xyz",
	Amount:      975,
	Fees:        146,
	VAT:         29,
	TotalAmount: 1150,
}

// T014: TestNewCreateCommand_Metadata
func TestNewCreateCommand_Metadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := bonus.NewCreateCommand("create", c, os.Stdout)

	use := "create"
	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatal("expected Short to be set")
	}
}

// T015: TestCreateBonusPayments_SuccessInline
func TestCreateBonusPayments_SuccessInline(t *testing.T) {
	c, writer, b := setupCreateMock(t, defaultCreateResponse)

	cmd := bonus.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("bonus", "pid1,4.25")
	_ = cmd.Flags().Set("bonus", "pid2,5.50")
	cmd.SetArgs([]string{"study-xyz"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	writer.Flush()
	output := b.String()

	if !strings.Contains(output, "bonus-abc-123") {
		t.Fatalf("expected output to contain bonus ID, got:\n%s", output)
	}
	if !strings.Contains(output, "study-xyz") {
		t.Fatalf("expected output to contain study ID, got:\n%s", output)
	}
}

// T016: TestCreateBonusPayments_SuccessFile
func TestCreateBonusPayments_SuccessFile(t *testing.T) {
	response := &client.CreateBonusPaymentsResponse{
		ID:          "bonus-file-123",
		Study:       "study-xyz",
		Amount:      425,
		Fees:        64,
		VAT:         13,
		TotalAmount: 502,
	}

	c, writer, b := setupCreateMock(t, response)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "bonuses.csv")
	if err := os.WriteFile(filePath, []byte("pid1,4.25\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	cmd := bonus.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("file", filePath)
	cmd.SetArgs([]string{"study-xyz"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	writer.Flush()
	output := b.String()

	if !strings.Contains(output, "bonus-file-123") {
		t.Fatalf("expected output to contain bonus ID, got:\n%s", output)
	}
}

// T017: TestCreateBonusPayments_APIError
func TestCreateBonusPayments_APIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "invalid participant ID"

	c.EXPECT().
		CreateBonusPayments(gomock.Any()).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := bonus.NewCreateCommand("create", c, os.Stdout)
	_ = cmd.Flags().Set("bonus", "bad-pid,4.25")
	cmd.SetArgs([]string{"study-xyz"})
	err := cmd.Execute()

	expected := fmt.Sprintf("error: %s", errorMessage)
	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%v'", expected, err)
	}
}

// T018: TestCreateBonusPayments_MutualExclusivity
func TestCreateBonusPayments_MutualExclusivity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := bonus.NewCreateCommand("create", c, os.Stdout)
	_ = cmd.Flags().Set("bonus", "pid1,4.25")
	_ = cmd.Flags().Set("file", "/some/file.csv")
	cmd.SetArgs([]string{"study-xyz"})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for mutual exclusivity violation")
	}

	if !strings.Contains(err.Error(), "cannot use both") {
		t.Fatalf("expected mutual exclusivity error, got: %s", err.Error())
	}
}

// T019: TestCreateBonusPayments_NonInteractive
func TestCreateBonusPayments_NonInteractive(t *testing.T) {
	response := &client.CreateBonusPaymentsResponse{
		ID:          "bonus-ni-123",
		Study:       "study-xyz",
		Amount:      975,
		Fees:        146,
		VAT:         29,
		TotalAmount: 1150,
	}

	c, writer, b := setupCreateMock(t, response)

	cmd := bonus.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("bonus", "pid1,4.25")
	_ = cmd.Flags().Set("non-interactive", "true")
	cmd.SetArgs([]string{"study-xyz"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	writer.Flush()
	output := b.String()

	// Non-interactive: first line should be the bonus ID for pipe extraction
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("expected output to have at least one line")
	}
	if lines[0] != "bonus-ni-123" {
		t.Fatalf("expected first line to be bonus ID 'bonus-ni-123', got: '%s'", lines[0])
	}
}

// T020: TestCreateBonusPayments_CSVOutput
func TestCreateBonusPayments_CSVOutput(t *testing.T) {
	response := &client.CreateBonusPaymentsResponse{
		ID:          "bonus-csv-123",
		Study:       "study-xyz",
		Amount:      975,
		Fees:        146,
		VAT:         29,
		TotalAmount: 1150,
	}

	c, writer, b := setupCreateMock(t, response)

	cmd := bonus.NewCreateCommand("create", c, writer)
	_ = cmd.Flags().Set("bonus", "pid1,4.25")
	_ = cmd.Flags().Set("csv", "true")
	cmd.SetArgs([]string{"study-xyz"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	writer.Flush()
	output := b.String()

	// CSV output should contain comma-separated values
	if !strings.Contains(output, ",") {
		t.Fatalf("expected CSV output with commas, got:\n%s", output)
	}
	if !strings.Contains(output, "bonus-csv-123") {
		t.Fatalf("expected CSV output to contain bonus ID, got:\n%s", output)
	}
}
