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

func TestNewUpdateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := credentials.NewUpdateCommand(c, os.Stdout)

	use := "update <credential-pool-id> [credentials]"
	short := "Update an existing credential pool"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestUpdateCredentialPool(t *testing.T) {
	credentialsString := "user1,pass1\\nuser2,pass2\\nuser3,pass3"
	credentialPoolID := "pool123456"

	tests := []struct {
		name           string
		args           []string
		mockReturn     *client.CreateCredentialPoolResponse
		mockError      error
		expectedOutput string
		expectedError  string
	}{
		{
			name: "successful update with string argument",
			args: []string{credentialPoolID, credentialsString},
			mockReturn: &client.CreateCredentialPoolResponse{
				CredentialPoolID: credentialPoolID,
			},
			mockError: nil,
			expectedOutput: `Credential pool updated successfully
Credential Pool ID: pool123456
`,
			expectedError: "",
		},
		{
			name:           "missing credentials error",
			args:           []string{credentialPoolID},
			mockReturn:     nil,
			mockError:      nil,
			expectedOutput: "",
			expectedError:  "credentials must be provided either as an argument or via -f flag",
		},
		{
			name:           "missing credential pool ID",
			args:           []string{},
			mockReturn:     nil,
			mockError:      nil,
			expectedOutput: "",
			expectedError:  "accepts between 1 and 2 arg(s), received 0",
		},
		{
			name:           "service unavailable",
			args:           []string{credentialPoolID, credentialsString},
			mockReturn:     nil,
			mockError:      errors.New("request failed with status 502: credentials service unavailable"),
			expectedOutput: "",
			expectedError:  "request failed with status 502: credentials service unavailable",
		},
		{
			name:           "not found error",
			args:           []string{credentialPoolID, credentialsString},
			mockReturn:     nil,
			mockError:      errors.New("request failed with status 404: credential pool not found"),
			expectedOutput: "",
			expectedError:  "request failed with status 404: credential pool not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			// Only expect API call if we have enough args and credentials
			if len(tt.args) > 1 {
				c.EXPECT().
					UpdateCredentialPool(tt.args[0], gomock.Any()).
					Return(tt.mockReturn, tt.mockError).
					Times(1)
			}

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := credentials.NewUpdateCommand(c, writer)

			var err error
			// For missing args test, need to use Execute() to trigger Cobra's argument validation
			if len(tt.args) == 0 {
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
				t.Fatalf("expected output:\\n'%s'\\n\\ngot:\\n'%s'", tt.expectedOutput, actual)
			}
		})
	}
}

func TestUpdateCredentialPoolFromFile(t *testing.T) {
	ctrl := setupMockController(t)
	c := mock_client.NewMockAPI(ctrl)

	credContent := "user1,pass1\\nuser2,pass2"
	credFile := createTempCredentialsFile(t, credContent)

	credentialPoolID := "pool789"
	c.EXPECT().
		UpdateCredentialPool(credentialPoolID, credContent).
		Return(&client.CreateCredentialPoolResponse{
			CredentialPoolID: credentialPoolID,
		}, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	cmd := credentials.NewUpdateCommand(c, writer)
	cmd.SetArgs([]string{credentialPoolID, "-f", credFile})
	err := cmd.Execute()
	writer.Flush()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedOutput := `Credential pool updated successfully
Credential Pool ID: pool789
`
	actual := b.String()
	if actual != expectedOutput {
		t.Fatalf("expected output:\\n'%s'\\n\\ngot:\\n'%s'", expectedOutput, actual)
	}
}
