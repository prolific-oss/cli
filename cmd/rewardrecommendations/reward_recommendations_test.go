package rewardrecommendations_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/rewardrecommendations"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/spf13/viper"
)

func TestNewCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := rewardrecommendations.NewCommand("reward-recommendations", c, os.Stdout)

	if cmd.Use != "reward-recommendations" {
		t.Fatalf("expected use: reward-recommendations; got %s", cmd.Use)
	}

	if cmd.Short != "Calculate recommended participant reward rates" {
		t.Fatalf("expected short: Calculate recommended participant reward rates; got %s", cmd.Short)
	}
}

func TestNewCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.RewardRecommendationsResponse{
		{
			Currency:                 "GBP",
			MinRewardPerHour:         900,
			RecommendedRewardPerHour: 1200,
		},
	}

	c.
		EXPECT().
		GetRewardRecommendations("abc123", "GBP", []string{"mandarin", "spanish"}).
		Return(&response, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := rewardrecommendations.NewCommand("reward-recommendations", c, writer)
	_ = cmd.Flags().Set("workspace", "abc123")
	_ = cmd.Flags().Set("currency", "GBP")
	_ = cmd.Flags().Set("screener-id", "mandarin")
	_ = cmd.Flags().Set("screener-id", "spanish")
	_ = cmd.RunE(cmd, []string{})

	writer.Flush()

	expected := `Currency:                     GBP
Minimum reward per hour:      9.00
Recommended reward per hour:  12.00
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewCommandUsesViperWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	viper.Set("workspace", "viper-workspace-id")
	t.Cleanup(func() { viper.Reset() })

	response := client.RewardRecommendationsResponse{
		{Currency: "USD", MinRewardPerHour: 1200, RecommendedRewardPerHour: 1500},
	}

	c.
		EXPECT().
		GetRewardRecommendations("viper-workspace-id", "USD", []string(nil)).
		Return(&response, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := rewardrecommendations.NewCommand("reward-recommendations", c, writer)
	_ = cmd.Flags().Set("currency", "USD")
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("expected no error; got %s", err)
	}
}

func TestNewCommandErrorsWithNoWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	viper.Reset()

	cmd := rewardrecommendations.NewCommand("reward-recommendations", c, os.Stdout)
	_ = cmd.Flags().Set("currency", "USD")
	err := cmd.RunE(cmd, []string{})

	expected := "error: please provide a workspace ID"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%v'\n", expected, err)
	}
}

func TestNewCommandErrorsWithNoCurrency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := rewardrecommendations.NewCommand("reward-recommendations", c, os.Stdout)
	_ = cmd.Flags().Set("workspace", "abc123")
	err := cmd.RunE(cmd, []string{})

	expected := "error: please provide a currency"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%v'\n", expected, err)
	}
}

func TestNewCommandErrorsWithUnsupportedCurrency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := rewardrecommendations.NewCommand("reward-recommendations", c, os.Stdout)
	_ = cmd.Flags().Set("workspace", "abc123")
	_ = cmd.Flags().Set("currency", "EUR")
	err := cmd.RunE(cmd, []string{})

	expected := "error: currency must be one of: USD, GBP"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%v'\n", expected, err)
	}
}

func TestNewCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "something went wrong"

	c.
		EXPECT().
		GetRewardRecommendations("abc123", "USD", []string(nil)).
		Return(nil, errors.New(errorMessage)).
		Times(1)

	cmd := rewardrecommendations.NewCommand("reward-recommendations", c, os.Stdout)
	_ = cmd.Flags().Set("workspace", "abc123")
	_ = cmd.Flags().Set("currency", "USD")
	err := cmd.RunE(cmd, []string{})

	expected := fmt.Sprintf("error: %s", errorMessage)
	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%v'\n", expected, err)
	}
}

func TestNewCommandErrorsWithEmptyResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.RewardRecommendationsResponse{}

	c.
		EXPECT().
		GetRewardRecommendations("abc123", "USD", []string(nil)).
		Return(&response, nil).
		Times(1)

	cmd := rewardrecommendations.NewCommand("reward-recommendations", c, os.Stdout)
	_ = cmd.Flags().Set("workspace", "abc123")
	_ = cmd.Flags().Set("currency", "USD")
	err := cmd.RunE(cmd, []string{})

	expected := "error: no reward recommendations returned"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%v'\n", expected, err)
	}
}
