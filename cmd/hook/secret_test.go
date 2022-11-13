package hook_test

import (
	"bufio"
	"bytes"
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
			{Value: "Leicester Square"},
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

	expected := fmt.Sprintf("%s\n", response.Results[0].Value)
	if b.String() != expected {
		t.Fatalf("expected '%s', got '%s'", expected, b.String())
	}
}
