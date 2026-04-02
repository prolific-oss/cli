package hook_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/hook"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

const testWorkspaceID = "63722982f9cc073ecc730f6b"
const errorMessage = "something went wrong"

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
			{ID: "63722971f9cc073ecc730f6a", Value: "Leicester Square", WorkspaceID: testWorkspaceID},
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

func TestNewCreateSecretCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := hook.NewCreateSecretCommand(c, os.Stdout)

	use := "create-secret"
	short := "Create a hook secret"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestNewCreateSecretCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := testWorkspaceID
	response := model.Secret{
		ID:          "63722971f9cc073ecc730f6a",
		Value:       "cGNqFPb6y0RT3XO9XVSessBDYIbHQ-...",
		WorkspaceID: workspaceID,
	}

	c.
		EXPECT().
		CreateHookSecret(client.CreateSecretPayload{WorkspaceID: workspaceID}).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := hook.NewCreateSecretCommand(c, writer)
	_ = cmd.Flags().Set("workspace", workspaceID)
	_ = cmd.Flags().Set("delete-old-secret", "true")
	_ = cmd.RunE(cmd, nil)
	writer.Flush()

	expected := fmt.Sprintf("Secret created successfully\nID:           %s\nSecret:       %s\nWorkspace ID: %s\n",
		response.ID, response.Value, response.WorkspaceID)

	actual := b.String()

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
	}
}

func TestNewCreateSecretCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := testWorkspaceID

	c.
		EXPECT().
		CreateHookSecret(client.CreateSecretPayload{WorkspaceID: workspaceID}).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := hook.NewCreateSecretCommand(c, os.Stdout)
	_ = cmd.Flags().Set("workspace", workspaceID)
	_ = cmd.Flags().Set("delete-old-secret", "true")
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewCreateSecretCommandConfirmationPrompt(t *testing.T) {
	workspaceID := testWorkspaceID
	response := model.Secret{
		ID:          "63722971f9cc073ecc730f6a",
		Value:       "cGNqFPb6y0RT3XO9XVSessBDYIbHQ-...",
		WorkspaceID: workspaceID,
	}

	tests := []struct {
		name           string
		input          string
		expectAPICall  bool
		expectedOutput string
	}{
		{
			name:          "user answers y proceeds",
			input:         "y\n",
			expectAPICall: true,
			expectedOutput: fmt.Sprintf(
				"This command will delete the old secret. Are you sure? (y/N): Secret created successfully\nID:           %s\nSecret:       %s\nWorkspace ID: %s\n",
				response.ID, response.Value, response.WorkspaceID,
			),
		},
		{
			name:          "user answers Y proceeds",
			input:         "Y\n",
			expectAPICall: true,
			expectedOutput: fmt.Sprintf(
				"This command will delete the old secret. Are you sure? (y/N): Secret created successfully\nID:           %s\nSecret:       %s\nWorkspace ID: %s\n",
				response.ID, response.Value, response.WorkspaceID,
			),
		},
		{
			name:           "user answers n cancels",
			input:          "n\n",
			expectAPICall:  false,
			expectedOutput: "This command will delete the old secret. Are you sure? (y/N): Secret creation cancelled.\n",
		},
		{
			name:           "empty input cancels",
			input:          "\n",
			expectAPICall:  false,
			expectedOutput: "This command will delete the old secret. Are you sure? (y/N): Secret creation cancelled.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			if tt.expectAPICall {
				c.EXPECT().
					CreateHookSecret(client.CreateSecretPayload{WorkspaceID: workspaceID}).
					Return(&response, nil).
					Times(1)
			}

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := hook.NewCreateSecretCommand(c, writer)
			cmd.SetIn(strings.NewReader(tt.input))
			_ = cmd.Flags().Set("workspace", workspaceID)
			_ = cmd.RunE(cmd, nil)
			writer.Flush()

			actual := b.String()
			if actual != tt.expectedOutput {
				t.Fatalf("expected\n'%s'\ngot\n'%s'\n", tt.expectedOutput, actual)
			}
		})
	}
}

func TestNewListSecretCommandHandlesErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

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
