package hook_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/prolificli/cmd/hook"
	"github.com/prolific-oss/prolificli/mock_client"
)

func TestNewHookCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := hook.NewHookCommand(c, os.Stdout)

	use := "hook"
	short := "Manage and view your hook subscriptions"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}
