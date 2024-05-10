package hook_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/hook"
	"github.com/benmatselby/prolificli/config"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
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
		GetHooks(gomock.Eq(""), gomock.Eq(true), gomock.Eq(client.DefaultRecordLimit), gomock.Eq(client.DefaultRecordOffset)).
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
		GetHooks(gomock.Eq(""), gomock.Eq(false), gomock.Eq(client.DefaultRecordLimit), gomock.Eq(client.DefaultRecordOffset)).
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

func TestNewListCommandCallsTheAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := "workspace-id"
	response := client.ListHooksResponse{
		Results: []model.Hook{
			{
				ID:          "hook-id",
				EventType:   "wibble",
				TargetURL:   config.GetApplicationURL(),
				IsEnabled:   true,
				WorkspaceID: "workspace-id",
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
		GetHooks(gomock.Eq(workspaceID), gomock.Eq(false), gomock.Eq(44), gomock.Eq(33)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := hook.NewListCommand("list", c, writer)
	_ = cmd.Flags().Set("disabled", "true")
	_ = cmd.Flags().Set("workspace", workspaceID)
	_ = cmd.Flags().Set("limit", "44")
	_ = cmd.Flags().Set("offset", "33")
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	actual := b.String()
	expected := fmt.Sprintf(`ID      Event  Target URL               Enabled Workspace ID
hook-id wibble %s true    workspace-id

Showing 1 record of 10
`, config.GetApplicationURL())

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}
