package filtersets_test

import (
	"os"
	"testing"

	"github.com/benmatselby/prolificli/cmd/filtersets"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/golang/mock/gomock"
)

func TestNewFilterSetCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := filtersets.NewFilterSetCommand(client, os.Stdout)

	use := "filter-sets"
	short := "Manage and view your filter sets"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}
