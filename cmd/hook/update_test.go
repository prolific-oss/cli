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

func TestNewUpdateSubscriptionCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := hook.NewUpdateSubscriptionCommand(c, os.Stdout)

	if cmd.Use != "update <subscription-id>" {
		t.Fatalf("expected use: update <subscription-id>; got %s", cmd.Use)
	}

	short := "Update a hook subscription"
	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewUpdateSubscriptionCommandRejectsEnableAndDisable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := hook.NewUpdateSubscriptionCommand(c, os.Stdout)
	_ = cmd.Flags().Set("enable", "true")
	_ = cmd.Flags().Set("disable", "true")
	err := cmd.RunE(cmd, []string{"sub-id-123"})

	expected := "--enable and --disable are mutually exclusive"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected error %q; got %v", expected, err)
	}
}

func TestNewUpdateSubscriptionCommandHandlesAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		UpdateHookSubscription("sub-id-123", client.UpdateHookPayload{}).
		Return(nil, errors.New("something went wrong"))

	cmd := hook.NewUpdateSubscriptionCommand(c, os.Stdout)
	err := cmd.RunE(cmd, []string{"sub-id-123"})

	expected := "error: something went wrong"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected error %q; got %v", expected, err)
	}
}

func TestNewUpdateSubscriptionCommandUpdatesFields(t *testing.T) {
	tests := []struct {
		name            string
		flagsToSet      map[string]string
		expectedPayload client.UpdateHookPayload
	}{
		{
			name:            "no flags — empty payload",
			flagsToSet:      map[string]string{},
			expectedPayload: client.UpdateHookPayload{},
		},
		{
			name:            "disable only",
			flagsToSet:      map[string]string{"disable": "true"},
			expectedPayload: client.UpdateHookPayload{IsEnabled: new(false)},
		},
		{
			name:            "enable only",
			flagsToSet:      map[string]string{"enable": "true"},
			expectedPayload: client.UpdateHookPayload{IsEnabled: new(true)},
		},
		{
			name:            "update target URL",
			flagsToSet:      map[string]string{"target-url": "https://new.example.com/hook/"},
			expectedPayload: client.UpdateHookPayload{TargetURL: new("https://new.example.com/hook/")},
		},
		{
			name:            "update event type",
			flagsToSet:      map[string]string{"event-type": "study.status.change"},
			expectedPayload: client.UpdateHookPayload{EventType: new("study.status.change")},
		},
		{
			name: "update all fields and enable",
			flagsToSet: map[string]string{
				"event-type": "study.status.change",
				"target-url": "https://new.example.com/hook/",
				"enable":     "true",
			},
			expectedPayload: client.UpdateHookPayload{
				EventType: new("study.status.change"),
				TargetURL: new("https://new.example.com/hook/"),
				IsEnabled: new(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			updated := &model.Hook{
				ID:          "sub-id-123",
				EventType:   "study.status.change",
				TargetURL:   "https://example.com/hook/",
				IsEnabled:   true,
				WorkspaceID: "workspace-id",
			}

			c.EXPECT().
				UpdateHookSubscription("sub-id-123", tt.expectedPayload).
				Return(updated, nil)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := hook.NewUpdateSubscriptionCommand(c, writer)
			for flag, val := range tt.flagsToSet {
				_ = cmd.Flags().Set(flag, val)
			}
			_ = cmd.RunE(cmd, []string{"sub-id-123"})

			writer.Flush()

			actual := b.String()
			expected := fmt.Sprintf(`Subscription updated successfully
ID:           %s
Event Type:   %s
Target URL:   %s
Enabled:      %v
Workspace ID: %s
`, updated.ID, updated.EventType, updated.TargetURL, updated.IsEnabled, updated.WorkspaceID)

			if actual != expected {
				t.Fatalf("expected\n%q\ngot\n%q", expected, actual)
			}
		})
	}
}
