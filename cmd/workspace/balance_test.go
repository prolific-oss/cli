package workspace_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/workspace"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewBalanceCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := workspace.NewBalanceCommand("balance", c, os.Stdout)

	if cmd.Use != "balance <workspace-id>" {
		t.Fatalf("expected use: balance <workspace-id>; got %s", cmd.Use)
	}

	if cmd.Short != "Show the balance of a workspace" {
		t.Fatalf("expected short: Show the balance of a workspace; got %s", cmd.Short)
	}
}

func TestNewBalanceCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := "abc123"

	response := client.WorkspaceBalanceResponse{
		CurrencyCode:     "USD",
		TotalBalance:     1428,
		AvailableBalance: 1428,
	}
	response.BalanceBreakdown.Rewards = 1428
	response.BalanceBreakdown.Fees = 0
	response.BalanceBreakdown.VAT = 0
	response.AvailableBalanceBreakdown.Rewards = 1428
	response.AvailableBalanceBreakdown.Fees = 0
	response.AvailableBalanceBreakdown.VAT = 0

	c.
		EXPECT().
		GetWorkspaceBalance(workspaceID).
		Return(&response, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := workspace.NewBalanceCommand("balance", c, writer)
	_ = cmd.RunE(cmd, []string{workspaceID})

	writer.Flush()

	expected := `Currency:  USD

Total Balance:  14.28
  Rewards:      14.28
  Fees:         0.00
  VAT:          0.00

Available Balance:  14.28
  Rewards:          14.28
  Fees:             0.00
  VAT:              0.00
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewBalanceCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := "abc123"
	errorMessage := "something went wrong"

	c.
		EXPECT().
		GetWorkspaceBalance(workspaceID).
		Return(nil, errors.New(errorMessage)).
		Times(1)

	cmd := workspace.NewBalanceCommand("balance", c, os.Stdout)
	err := cmd.RunE(cmd, []string{workspaceID})

	expected := fmt.Sprintf("error: %s", errorMessage)
	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
