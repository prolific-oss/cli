package credentials_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
)

// createTempCredentialsFile creates a temporary CSV file with credentials for testing
func createTempCredentialsFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	credFile := filepath.Join(tmpDir, "credentials.csv")
	err := os.WriteFile(credFile, []byte(content), 0600)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return credFile
}

// setupMockController creates a new gomock controller with cleanup
func setupMockController(t *testing.T) *gomock.Controller {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	return ctrl
}
