package study_test

import (
	"testing"

	"github.com/benmatselby/prolificli/cmd/study"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/golang/mock/gomock"
)

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := study.NewStudyCommand(client)

	use := "study"
	short := "Study related commands"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}
