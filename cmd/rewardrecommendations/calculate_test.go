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
)

func TestNewCalculateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := rewardrecommendations.NewCalculateCommand(c, os.Stdout)

	use := "reward-recommendations"
	short := "Calculate recommended reward rates for participants"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewCalculateCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.RewardRecommendationsResponse{
		CurrencyCode:              "GBP",
		MinRewardPerHour:          8.00,
		EstimatedRewardPerHour:    10.00,
		MaxRewardPerHour:          12.00,
		MinRewardForEstimatedTime: 1.33,
		EstimatedReward:           1.67,
		MaxRewardForEstimatedTime: 2.00,
	}

	c.
		EXPECT().
		GetRewardRecommendations("GBP", 10, []string{}).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := rewardrecommendations.NewCalculateCommand(c, writer)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `Metric                     Value
Currency                   GBP
Min Reward Per Hour        8.00
Estimated Reward Per Hour  10.00
Max Reward Per Hour        12.00
Min Reward For Study       1.33
Estimated Reward For Study 1.67
Max Reward For Study       2.00
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewCalculateCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "API error"

	c.
		EXPECT().
		GetRewardRecommendations("GBP", 10, []string{}).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := rewardrecommendations.NewCalculateCommand(c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
