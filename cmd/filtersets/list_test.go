package filtersets_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/client"
	"github.com/benmatselby/prolificli/cmd/filtersets"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := filtersets.NewListCommand("list", client, os.Stdout)

	use := "list"
	short := "Provide a list of your filter sets"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewListCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := "777111999"

	response := client.ListFilterSetsResponse{
		Results: []model.FilterSet{
			{
				ID:   "1122",
				Name: "Left handed people",
			},
			{
				ID:   "3344",
				Name: "Radiohead fans",
			},
		},
		JSONAPIMeta: &client.JSONAPIMeta{
			Meta: struct {
				Count int `json:"count"`
			}{
				Count: 10,
			},
		},
	}

	c.
		EXPECT().
		GetFilterSets(gomock.Eq(workspaceID), client.DefaultRecordLimit, client.DefaultRecordOffset).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := filtersets.NewListCommand("list", c, writer)
	_ = cmd.Flags().Set("workspace", workspaceID)
	_ = cmd.RunE(cmd, nil)

	writer.Flush()

	expected := `ID   Name
1122 Left handed people
3344 Radiohead fans

Showing 2 records of 10
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}

func TestNewListCommandReturnsErrorIfWorkspaceNotDefined(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := filtersets.NewListCommand("list", c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	expected := "error: please provide a workspace ID"
	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewListCommandHandlesAnAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	workspaceID := "777111999"

	errorMessage := "The city has got be chasing stars"

	c.
		EXPECT().
		GetFilterSets(gomock.Eq(workspaceID), client.DefaultRecordLimit, client.DefaultRecordOffset).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := filtersets.NewListCommand("list", c, writer)
	_ = cmd.Flags().Set("workspace", workspaceID)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)
	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
