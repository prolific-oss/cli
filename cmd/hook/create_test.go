package hook_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/hook"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewCreateSubscriptionCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := hook.NewCreateSubscriptionCommand(c, os.Stdout)

	if cmd.Use != "create" {
		t.Fatalf("expected use: create; got %s", cmd.Use)
	}

	short := "Create a hook subscription"
	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewCreateSubscriptionCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "something went wrong"
	payload := client.CreateHookPayload{
		WorkspaceID: "workspace-id",
		EventType:   "study.status.change",
		TargetURL:   "https://example.com/hook",
	}

	c.EXPECT().
		CreateHookSubscription(gomock.Eq(payload)).
		Return(nil, "", errors.New(errorMessage))

	cmd := hook.NewCreateSubscriptionCommand(c, os.Stdout)
	_ = cmd.Flags().Set("workspace", payload.WorkspaceID)
	_ = cmd.Flags().Set("event-type", payload.EventType)
	_ = cmd.Flags().Set("target-url", payload.TargetURL)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)
	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewCreateSubscriptionCommandHandlesConfirmationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	payload := client.CreateHookPayload{
		WorkspaceID: "workspace-id",
		EventType:   "study.status.change",
		TargetURL:   "https://example.com/hook",
	}
	created := &model.Hook{
		ID:          "sub-id-123",
		EventType:   payload.EventType,
		TargetURL:   payload.TargetURL,
		IsEnabled:   false,
		WorkspaceID: payload.WorkspaceID,
	}
	const secret = "x-hook-secret-value" //nolint:gosec

	c.EXPECT().
		CreateHookSubscription(gomock.Eq(payload)).
		Return(created, secret, nil)

	c.EXPECT().
		ConfirmHookSubscription(gomock.Eq(created.ID), gomock.Eq(secret)).
		Return(nil, errors.New("confirmation failed"))

	cmd := hook.NewCreateSubscriptionCommand(c, os.Stdout)
	_ = cmd.Flags().Set("workspace", payload.WorkspaceID)
	_ = cmd.Flags().Set("event-type", payload.EventType)
	_ = cmd.Flags().Set("target-url", payload.TargetURL)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("subscription created (ID: %s) but confirmation failed: confirmation failed", created.ID)
	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewCreateSubscriptionCommandCallsTheAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	payload := client.CreateHookPayload{
		WorkspaceID: "workspace-id",
		EventType:   "study.status.change",
		TargetURL:   "https://example.com/hook",
	}
	created := &model.Hook{
		ID:          "sub-id-123",
		EventType:   payload.EventType,
		TargetURL:   payload.TargetURL,
		IsEnabled:   false,
		WorkspaceID: payload.WorkspaceID,
	}
	confirmed := &model.Hook{
		ID:          created.ID,
		EventType:   payload.EventType,
		TargetURL:   payload.TargetURL,
		IsEnabled:   true,
		WorkspaceID: payload.WorkspaceID,
	}
	const secret = "x-hook-secret-value" //nolint:gosec

	c.EXPECT().
		CreateHookSubscription(gomock.Eq(payload)).
		Return(created, secret, nil)

	c.EXPECT().
		ConfirmHookSubscription(gomock.Eq(created.ID), gomock.Eq(secret)).
		Return(confirmed, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := hook.NewCreateSubscriptionCommand(c, writer)
	_ = cmd.Flags().Set("workspace", payload.WorkspaceID)
	_ = cmd.Flags().Set("event-type", payload.EventType)
	_ = cmd.Flags().Set("target-url", payload.TargetURL)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	actual := b.String()
	expected := fmt.Sprintf(`Subscription created successfully
ID:           %s
Event Type:   %s
Target URL:   %s
Enabled:      %v
Workspace ID: %s
`, confirmed.ID, confirmed.EventType, confirmed.TargetURL, confirmed.IsEnabled, confirmed.WorkspaceID)

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}
