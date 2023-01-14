package hook_test

import (
	"bufio"
	"bytes"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/hook"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewEventTypeCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := hook.NewEventTypeCommand("event-types", c, os.Stdout)

	use := "event-types"
	short := "List of event types you can subscribe to"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewEventTypeCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.ListHookEventTypesResponse{
		Results: []model.HookEventType{
			{
				EventType:   "wibble",
				Description: "The wibble event",
			},
		},
	}

	c.
		EXPECT().
		GetHookEventTypes().
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := hook.NewEventTypeCommand("event-types", c, writer)
	_ = cmd.RunE(cmd, nil)
	writer.Flush()

	expected := `Event Type Description
wibble     The wibble event
`

	actual := b.String()

	if actual != expected {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}
