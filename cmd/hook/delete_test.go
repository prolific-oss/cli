package hook_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/hook"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewDeleteSubscriptionCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := hook.NewDeleteSubscriptionCommand(c, os.Stdout)

	if cmd.Use != "delete <subscription-id>" {
		t.Fatalf("expected use: delete <subscription-id>; got %s", cmd.Use)
	}

	short := "Delete a hook subscription"
	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewDeleteSubscriptionCommandHandlesAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		DeleteHookSubscription("sub-id-123").
		Return(errors.New("something went wrong"))

	cmd := hook.NewDeleteSubscriptionCommand(c, os.Stdout)
	err := cmd.RunE(cmd, []string{"sub-id-123"})

	expected := "error: something went wrong"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected error %q; got %v", expected, err)
	}
}

func TestNewDeleteSubscriptionCommandDeletesSuccessfully(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		DeleteHookSubscription("sub-id-123").
		Return(nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := hook.NewDeleteSubscriptionCommand(c, writer)
	err := cmd.RunE(cmd, []string{"sub-id-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	writer.Flush()

	expected := "Subscription sub-id-123 deleted successfully\n"
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n%q\ngot\n%q", expected, actual)
	}
}
