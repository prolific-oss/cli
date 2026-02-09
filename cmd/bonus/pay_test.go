package bonus_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/bonus"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewPayCommand_Metadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := bonus.NewPayCommand("pay", c, os.Stdout)

	use := "pay"
	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatal("expected Short to be set")
	}
}

func TestPayBonusPayments_SuccessNonInteractive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		PayBonusPayments("bonus-pay-123").
		Return(nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := bonus.NewPayCommand("pay", c, writer)
	_ = cmd.Flags().Set("non-interactive", "true")
	cmd.SetArgs([]string{"bonus-pay-123"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	writer.Flush()
	output := b.String()

	if !strings.Contains(output, "asynchronously") {
		t.Fatalf("expected output to mention asynchronous processing, got:\n%s", output)
	}
}

func TestPayBonusPayments_APIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "bonus already paid"

	c.EXPECT().
		PayBonusPayments("bonus-err-123").
		Return(errors.New(errorMessage)).
		AnyTimes()

	cmd := bonus.NewPayCommand("pay", c, os.Stdout)
	_ = cmd.Flags().Set("non-interactive", "true")
	cmd.SetArgs([]string{"bonus-err-123"})
	err := cmd.Execute()

	expected := fmt.Sprintf("error: %s", errorMessage)
	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%v'", expected, err)
	}
}

func TestPayBonusPayments_Declined(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// PayBonusPayments should NOT be called when user declines
	// (no EXPECT set, so any call would fail the mock controller)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := bonus.NewPayCommand("pay", c, writer)
	cmd.SetArgs([]string{"bonus-decline-123"})
	// Inject "n" response via Cobra's SetIn
	cmd.SetIn(strings.NewReader("n\n"))

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	writer.Flush()
	output := b.String()

	if !strings.Contains(output, "cancelled") && !strings.Contains(output, "Cancelled") {
		t.Fatalf("expected cancellation message, got:\n%s", output)
	}
}
