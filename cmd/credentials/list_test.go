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

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := credentials.NewListCommand(c, os.Stdout)

	use := "list"
	short := "List credential pools for a workspace"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestListCredentialPools(t *testing.T) {
	workspaceID := "workspace123"

	tests := []struct {
		name           string
		workspaceID    string
		mockReturn     *client.ListCredentialPoolsResponse
		mockError      error
		expectedOutput string
		expectedError  string
	}{
		{
			name:        "successful list with multiple pools",
			workspaceID: workspaceID,
			mockReturn: &client.ListCredentialPoolsResponse{
				CredentialPools: []client.CredentialPoolSummary{
					{
						CredentialPoolID:     "cred_pool_12345",
						TotalCredentials:     100,
						AvailableCredentials: 75,
					},
					{
						CredentialPoolID:     "cred_pool_67890",
						TotalCredentials:     50,
						AvailableCredentials: 50,
					},
				},
			},
			mockError: nil,
			expectedOutput: `Credential Pools for workspace workspace123:

Credential Pool ID: cred_pool_12345
  Total Credentials: 100
  Available Credentials: 75

Credential Pool ID: cred_pool_67890
  Total Credentials: 50
  Available Credentials: 50

`,
			expectedError: "",
		},
		{
			name:        "successful list with single pool",
			workspaceID: workspaceID,
			mockReturn: &client.ListCredentialPoolsResponse{
				CredentialPools: []client.CredentialPoolSummary{
					{
						CredentialPoolID:     "cred_pool_12345",
						TotalCredentials:     100,
						AvailableCredentials: 75,
					},
				},
			},
			mockError: nil,
			expectedOutput: `Credential Pools for workspace workspace123:

Credential Pool ID: cred_pool_12345
  Total Credentials: 100
  Available Credentials: 75

`,
			expectedError: "",
		},
		{
			name:        "no credential pools found",
			workspaceID: workspaceID,
			mockReturn: &client.ListCredentialPoolsResponse{
				CredentialPools: []client.CredentialPoolSummary{},
			},
			mockError:      nil,
			expectedOutput: "No credential pools found for workspace workspace123\n",
			expectedError:  "",
		},
		{
			name:           "workspace ID missing error",
			workspaceID:    "",
			mockReturn:     nil,
			mockError:      nil,
			expectedOutput: "",
			expectedError:  "required flag(s) \"workspace-id\" not set",
		},
		{
			name:           "service unavailable",
			workspaceID:    workspaceID,
			mockReturn:     nil,
			mockError:      errors.New("request failed with status 502: credentials service unavailable"),
			expectedOutput: "",
			expectedError:  "request failed with status 502: credentials service unavailable",
		},
		{
			name:           "forbidden error",
			workspaceID:    workspaceID,
			mockReturn:     nil,
			mockError:      errors.New("request failed with status 403: user does not have permission"),
			expectedOutput: "",
			expectedError:  "request failed with status 403: user does not have permission",
		},
		{
			name:           "bad request - workspace_id missing",
			workspaceID:    workspaceID,
			mockReturn:     nil,
			mockError:      errors.New("request failed with status 400: workspace_id query parameter is required"),
			expectedOutput: "",
			expectedError:  "request failed with status 400: workspace_id query parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			// Only expect API call if we have workspace ID
			if tt.workspaceID != "" {
				c.EXPECT().
					ListCredentialPools(tt.workspaceID).
					Return(tt.mockReturn, tt.mockError).
					Times(1)
			}

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := credentials.NewListCommand(c, writer)
			if tt.workspaceID != "" {
				_ = cmd.Flags().Set("workspace-id", tt.workspaceID)
			}

			var err error
			// For workspace ID missing test, need to use Execute() to trigger Cobra's flag validation
			if tt.workspaceID == "" {
				cmd.SetArgs([]string{})
				err = cmd.Execute()
			} else {
				err = cmd.RunE(cmd, []string{})
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
