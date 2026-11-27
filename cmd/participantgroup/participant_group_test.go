package participantgroup_test

import (
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/participantgroup"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestNewParticipantCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := participantgroup.NewParticipantCommand(client, os.Stdout)

	use := "participant-group"
	short := "Manage and view your participant groups"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestParticipantGroupAlias(t *testing.T) {
	for _, name := range []string{"participant-group", "participant"} {
		t.Run(name, func(t *testing.T) {
			c := mock_client.NewMockAPI(gomock.NewController(t))
			parent := participantgroup.NewParticipantCommand(c, io.Discard)
			root := &cobra.Command{Use: "prolific"}
			root.AddCommand(parent)
			found, args, err := root.Find([]string{name, "search", "pilot"})
			require.NoError(t, err)
			require.Equal(t, "search", found.Name())
			require.Equal(t, []string{"pilot"}, args)
		})
	}
}
