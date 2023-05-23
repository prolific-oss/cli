package requirement_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	requirement "github.com/prolific-oss/cli/cmd/requirements"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := requirement.NewListCommand(client, os.Stdout)

	use := "requirements"
	short := "List all eligibility requirements available for your study"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}
