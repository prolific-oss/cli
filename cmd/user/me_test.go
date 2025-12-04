package user_test

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/user"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewMeCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := user.NewMeCommand(client, os.Stdout)

	use := "whoami"
	short := "View details about your account"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestRenderMe(t *testing.T) {
	tt := []struct {
		name   string
		output string
		err    error
	}{
		{name: "can return a list of sprints", output: `Bald Rick
ID:                616
Email:             baldrick@turnip.co

`, err: nil},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			me := client.MeResponse{
				ID:               "616",
				FirstName:        "Bald",
				LastName:         "Rick",
				Email:            "baldrick@turnip.co",
				CurrencyCode:     "GBP",
				AvailableBalance: 1000,
				Balance:          850,
			}

			c.
				EXPECT().
				GetMe().
				Return(&me, nil).
				AnyTimes()

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := user.NewMeCommand(c, writer)
			err := cmd.RunE(cmd, nil)

			if err != nil {
				t.Fatalf("did not expect error, got %v", err)
			}

			writer.Flush()

			if strings.ReplaceAll(b.String(), " ", "") != strings.ReplaceAll(tc.output, " ", "") {
				t.Fatalf("expected \n'%s'\ngot\n'%s'", tc.output, b.String())
			}
		})
	}
}

func TestRenderMeHandlesFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		GetMe().
		Return(nil, fmt.Errorf("Failure to look within")).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := user.NewMeCommand(c, writer)
	err := cmd.RunE(cmd, nil)

	if err.Error() != "Failure to look within" {
		t.Fatalf("expected a specific error, got %v", err)
	}

	writer.Flush()
}
