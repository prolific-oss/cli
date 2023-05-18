package hook_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/prolificli/cmd/hook"
	"github.com/prolific-oss/prolificli/mock_client"
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
