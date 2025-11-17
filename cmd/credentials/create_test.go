package credentials_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/credentials"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewCreateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := credentials.NewCreateCommand(c, os.Stdout)

	use := "create"
	short := "Create a new credential pool"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestCreateCredentialPool(t *testing.T) {
	credentialsString := "user1,pass1\nuser2,pass2\nuser3,pass3"
	credentialPoolID := "pool123456"
	workspaceID := "workspace123"

	tests := []struct {
		name           string
		args           []string
		workspaceID    string
		mockReturn     *client.CreateCredentialPoolResponse
		mockError      error
		expectedOutput string
		expectedError  string
	}{
		{
			name:        "successful creation with string argument",
			workspaceID: workspaceID,
			args:        []string{credentialsString},
			mockReturn: &client.CreateCredentialPoolResponse{
				CredentialPoolID: credentialPoolID,
			},
			mockError: nil,
			expectedOutput: `Credential pool created successfully
Credential Pool ID: pool123456
`,
			expectedError: "",
		},
		{
			name:           "credentials missing error",
			args:           []string{},
			workspaceID:    workspaceID,
			mockReturn:     nil,
			mockError:      nil,
			expectedOutput: "",
			expectedError:  "credentials must be provided either as an argument or via -f flag",
		},
		{
			name:           "workspace ID missing error",
			args:           []string{credentialsString},
			workspaceID:    "",
			mockReturn:     nil,
			mockError:      nil,
			expectedOutput: "",
			expectedError:  "required flag(s) \"workspace-id\" not set",
		},
		{
			name:           "service unavailable",
			args:           []string{credentialsString},
			workspaceID:    workspaceID,
			mockReturn:     nil,
			mockError:      errors.New("request failed with status 502: credentials service unavailable"),
			expectedOutput: "",
			expectedError:  "request failed with status 502: credentials service unavailable",
		},
		{
			name:           "bad request",
			args:           []string{credentialsString},
			workspaceID:    workspaceID,
			mockReturn:     nil,
			mockError:      errors.New("request failed: study does not have credentials"),
			expectedOutput: "",
			expectedError:  "request failed: study does not have credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			// Only expect API call if we have workspace ID and args
			if len(tt.args) > 0 && tt.workspaceID != "" {
				c.EXPECT().
					CreateCredentialPool(gomock.Any(), tt.workspaceID).
					Return(tt.mockReturn, tt.mockError).
					Times(1)
			}

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := credentials.NewCreateCommand(c, writer)
			if tt.workspaceID != "" {
				_ = cmd.Flags().Set("workspace-id", tt.workspaceID)
			}

			var err error
			// For workspace ID missing test, need to use Execute() to trigger Cobra's flag validation
			if tt.workspaceID == "" {
				cmd.SetArgs(tt.args)
				err = cmd.Execute()
			} else {
				err = cmd.RunE(cmd, tt.args)
			}
			writer.Flush()

			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error '%s', got nil", tt.expectedError)
				}
				if err.Error() != tt.expectedError {
					t.Fatalf("expected error '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			actual := b.String()
			if actual != tt.expectedOutput {
				t.Fatalf("expected output:\n'%s'\n\ngot:\n'%s'", tt.expectedOutput, actual)
			}
		})
	}
}

func TestCreateCredentialPoolFromFile(t *testing.T) {
	ctrl := setupMockController(t)
	c := mock_client.NewMockAPI(ctrl)

	credContent := "user1,pass1\nuser2,pass2"
	credFile := createTempCredentialsFile(t, credContent)
	workspaceID := "workspace789"

	credentialPoolID := "pool789"
	c.EXPECT().
		CreateCredentialPool(credContent, workspaceID).
		Return(&client.CreateCredentialPoolResponse{
			CredentialPoolID: credentialPoolID,
		}, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	cmd := credentials.NewCreateCommand(c, writer)
	cmd.SetArgs([]string{"-w", workspaceID, "-f", credFile})
	err := cmd.Execute()
	writer.Flush()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedOutput := `Credential pool created successfully
Credential Pool ID: pool789
`
	actual := b.String()
	if actual != expectedOutput {
		t.Fatalf("expected output:\n'%s'\n\ngot:\n'%s'", expectedOutput, actual)
	}
}
