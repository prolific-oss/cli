package hook_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/hook"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewEventListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := hook.NewEventListCommand("events", client, os.Stdout)

	use := "events"
	short := "Provide a list of events for your subscription"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewEventListCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	subscriptionID := "777111999"

	hookEventDate, _ := time.Parse("2006-01-02 15:04", "2022-07-24 08:04")

	response := client.ListHookEventsResponse{
		Results: []model.HookEvent{
			{
				ID:          "1122",
				DateCreated: hookEventDate,
				DateUpdated: hookEventDate,
				Status:      "SUCCEEDED",
				ResourceID:  "313",
				EventType:   "",
				TargetURL:   "",
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
		GetEvents(gomock.Eq(subscriptionID), gomock.Eq(44), gomock.Eq(29)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := hook.NewEventListCommand("events", c, writer)
	_ = cmd.Flags().Set("subscription", subscriptionID)
	_ = cmd.Flags().Set("limit", "44")
	_ = cmd.Flags().Set("offset", "29")
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `ID   Created          Updated          Status    Resource ID
1122 24-07-2022 08:04 24-07-2022 08:04 SUCCEEDED 313

Showing 1 record of 10
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}

func TestNewEventListCommandProvidesErrorIfSubmissionNotPassedIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := hook.NewEventListCommand("events", c, os.Stdout)
	error := cmd.RunE(cmd, nil)

	expected := `error: please provide a subscription ID`

	if error.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, error.Error())
	}
}

func TestNewEventListCommandSetsDefaultsForLimitOffset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	subscriptionID := "subscription-id"
	failureMessage := "Get events failed"

	c.
		EXPECT().
		// This is the defaults we have
		GetEvents(gomock.Eq(subscriptionID), gomock.Eq(client.DefaultRecordLimit), gomock.Eq(client.DefaultRecordOffset)).
		Return(nil, errors.New(failureMessage)).
		AnyTimes()

	cmd := hook.NewEventListCommand("events", c, os.Stdout)
	_ = cmd.Flags().Set("subscription", subscriptionID)
	error := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", failureMessage)

	if error.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, error.Error())
	}
}
