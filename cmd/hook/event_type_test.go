package hook_test

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/hook"
	"github.com/benmatselby/prolificli/mock_client"
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

	event := "event.fluxcapacitor.wibble"

	response := client.ListHookEventTypesResponse{
		Results: []string{event},
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

	if strings.Trim(b.String(), "\n") != event {
		t.Fatalf("expected '%s', got '%s'", event, b.String())
	}
}
