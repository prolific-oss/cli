package campaign_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/campaign"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := campaign.NewListCommand("campaigns", c, os.Stdout)

	use := "campaigns"
	short := "Provide details about your campaigns"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewCampaignCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.ListCampaignsResponse{
		Results: []model.Campaign{
			{
				ID:         "444",
				Name:       "Chili Peppers",
				SignupLink: "https://app.prolific.com/register/participant/waitlist/?campaign_code=BLUEBIRD",
			},
			{
				ID:         "555",
				Name:       "Jovi",
				SignupLink: "https://app.prolific.com/register/participant/waitlist/?campaign_code=ORANGEBURST",
			},
		},
		JSONAPIMeta: &client.JSONAPIMeta{
			Meta: struct {
				Count int `json:"count"`
			}{
				Count: 10,
			},
		},
	}

	c.
		EXPECT().
		GetCampaigns(gomock.Eq("991199"), gomock.Eq(10), gomock.Eq(2)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := campaign.NewListCommand("campaign", c, writer)
	_ = cmd.Flags().Set("workspace", "991199")
	_ = cmd.Flags().Set("limit", "10")
	_ = cmd.Flags().Set("offset", "2")
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `ID  Name          Link
444 Chili Peppers https://app.prolific.com/register/participant/waitlist/?campaign_code=BLUEBIRD
555 Jovi          https://app.prolific.com/register/participant/waitlist/?campaign_code=ORANGEBURST

Showing 2 records of 10
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}

func TestNewListCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "I am titanium"

	c.
		EXPECT().
		GetCampaigns(gomock.Eq("991199"), client.DefaultRecordLimit, client.DefaultRecordOffset).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := campaign.NewListCommand("campaign", c, os.Stdout)
	_ = cmd.Flags().Set("workspace", "991199")
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
