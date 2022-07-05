package requirement_test

import (
	"os"
	"testing"

	requirement "github.com/benmatselby/prolificli/cmd/requirements"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/golang/mock/gomock"
)

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := requirement.NewListCommand(client, os.Stdout)

	use := "list"
	short := "List all eligibility requirements"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}
