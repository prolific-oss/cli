package hook_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/prolificli/cmd/hook"
	"github.com/prolific-oss/prolificli/mock_client"
)

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := hook.NewListCommand("list", c, os.Stdout)

	use := "list"
	short := "Provide details about your hook subscriptions"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewListCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "I am titanium"

	c.
		EXPECT().
		GetHooks(gomock.Eq(true)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := hook.NewListCommand("list", c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewListCommandCanAskForDisabledHooks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "I am titanium"

	c.
		EXPECT().
		GetHooks(gomock.Eq(false)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := hook.NewListCommand("list", c, os.Stdout)
	_ = cmd.Flags().Set("disabled", "true")
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
