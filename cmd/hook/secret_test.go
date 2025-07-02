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
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewListSecretCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := hook.NewListSecretCommand("secrets", c, os.Stdout)

	use := "secrets"
	short := "List your hook secrets"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewListSecretCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.ListSecretsResponse{
		Results: []model.Secret{
			{ID: "63722971f9cc073ecc730f6a", Value: "Leicester Square", WorkspaceID: "63722982f9cc073ecc730f6b"},
		},
	}

	c.
		EXPECT().
		GetHookSecrets("").
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := hook.NewListSecretCommand("secrets", c, writer)
	_ = cmd.RunE(cmd, nil)
	writer.Flush()

	expected := `ID                       Secret           Workspace ID
63722971f9cc073ecc730f6a Leicester Square 63722982f9cc073ecc730f6b
`

	actual := b.String()

	if actual != expected {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}

func TestNewListSecretCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "I will try to fix you"
	workspaceID := "workspace-id"

	c.
		EXPECT().
		GetHookSecrets(gomock.Eq(workspaceID)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := hook.NewListSecretCommand("secrets", c, os.Stdout)
	_ = cmd.Flags().Set("workspace", workspaceID)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
