package hook

import (
	"fmt"
	"io"
	"strings"

	"github.com/benmatselby/prolificli/client"
	"github.com/spf13/cobra"
)

// NewEventTypeCommand creates a new `hook event-types` command to give you details about
// your which events you can register subscriptions for.
func NewEventTypeCommand(commandName string, client client.API, w io.Writer) *cobra.Command {

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "List of event types you can subscribe to",
		RunE: func(cmd *cobra.Command, args []string) error {

			err := RenderEventTypes(client, w)
			if err != nil {
				return fmt.Errorf("Error: %s\n", strings.ReplaceAll(err.Error(), "\n", ""))
			}

			return nil
		},
	}

	return cmd
}

// RenderEventTypes will show all of the event types that can be registered.
func RenderEventTypes(client client.API, w io.Writer) error {

	eventTypes, err := client.GetHookEventTypes()
	if err != nil {
		return err
	}

	for _, event := range eventTypes.Results {
		fmt.Fprintln(w, event)
	}

	return nil
}
