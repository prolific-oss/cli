package filtersets_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/benmatselby/prolificli/cmd/filtersets"
	"github.com/benmatselby/prolificli/mock_client"
	"github.com/benmatselby/prolificli/model"
	"github.com/golang/mock/gomock"
)

func TestNewViewCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := filtersets.NewViewCommand("view", c, os.Stdout)

	use := "view"
	short := "Provide details about your filter set"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewViewCommandHandlesNoFiltersInTheSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	filterSetID := "991199"
	response := model.FilterSet{
		ID:                       filterSetID,
		Name:                     "The best participants",
		OrganisationID:           "123",
		WorkspaceID:              "111",
		Version:                  2,
		EligibleParticipantCount: 22222,
		IsDeleted:                false,
		IsLocked:                 true,
	}

	c.
		EXPECT().
		GetFilterSet(gomock.Eq(filterSetID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := filtersets.NewViewCommand("view", c, writer)
	_ = cmd.RunE(cmd, []string{filterSetID})

	writer.Flush()

	expected := `The best participants

Organisation:               123
Workspace:                  111
Version:                    2
Eligible participant count: 22222
Locked:                     true
Deleted:                    false

---

No filters

---

View filter set in the application: https://app.prolific.com/researcher/workspaces/111/screener-sets/991199
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}

func TestNewViewCommandHandlesFiltersInTheSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	filterSetID := "991199"
	response := model.FilterSet{
		ID:                       filterSetID,
		Name:                     "The best participants",
		OrganisationID:           "123",
		WorkspaceID:              "111",
		Version:                  2,
		EligibleParticipantCount: 22222,
		IsDeleted:                false,
		IsLocked:                 true,
		Filters: []model.Filter{
			{
				FilterID:       "123",
				SelectedValues: []string{"a", "b", "c"},
			},
			{
				FilterID:       "456",
				SelectedValues: []string{"d", "e", "f"},
			},
		},
	}

	c.
		EXPECT().
		GetFilterSet(gomock.Eq(filterSetID)).
		Return(&response, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := filtersets.NewViewCommand("view", c, writer)
	_ = cmd.RunE(cmd, []string{filterSetID})

	writer.Flush()

	expected := `The best participants

Organisation:               123
Workspace:                  111
Version:                    2
Eligible participant count: 22222
Locked:                     true
Deleted:                    false

---

Filter ID: 123
Selected values:
- a
- b
- c

Filter ID: 456
Selected values:
- d
- e
- f

---

View filter set in the application: https://app.prolific.com/researcher/workspaces/111/screener-sets/991199
`
	actual := b.String()
	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, b.String())
	}
}

func TestNewViewCommandHandlesErrorsFromTheCliParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "please provide a filter set ID"

	cmd := filtersets.NewViewCommand("view", c, os.Stdout)
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestNewViewCommandHandlesErrorsFromTheAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	filterSetID := "91919191"
	errorMessage := "API says no"

	c.
		EXPECT().
		GetFilterSet(gomock.Eq(filterSetID)).
		Return(nil, errors.New(errorMessage)).
		AnyTimes()

	cmd := filtersets.NewViewCommand("view", c, os.Stdout)
	err := cmd.RunE(cmd, []string{filterSetID})

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}
