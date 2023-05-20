package hook_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/prolificli/client"
	"github.com/prolific-oss/prolificli/cmd/hook"
	"github.com/prolific-oss/prolificli/mock_client"
	"github.com/prolific-oss/prolificli/model"
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

func TestNewEventTypeCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "The sun went down on us"

	c.
		EXPECT().
		GetHookEventTypes().
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := hook.NewEventTypeCommand("event-types", c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
