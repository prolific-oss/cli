package filters_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/filters"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/stretchr/testify/assert"
)

func TestNewFiltersCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := filters.NewFiltersCommand(c, nil)

	assert.Equal(t, "filters", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	var names []string
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}
	assert.ElementsMatch(t, []string{"list", "search"}, names)
}
