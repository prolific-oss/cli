package collection

import (
	"errors"
	"fmt"
	"io"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	studyui "github.com/prolific-oss/cli/ui/study"
	"github.com/spf13/cobra"
)

// PublishOptions is the options for the publish collection command.
type PublishOptions struct {
	Args         []string
	Participants int
	Name         string
	Description  string
}

// NewPublishCommand creates a new `collection publish` command to publish
// a collection as a study.
func NewPublishCommand(c client.API, w io.Writer) *cobra.Command {
	var opts PublishOptions

	cmd := &cobra.Command{
		Use:   "publish <collection-id>",
		Args:  cobra.MinimumNArgs(1),
		Short: "Publish a collection as a study",
		Long: `Publish a collection as a study

This command creates and publishes a study from an AI Task Builder Collection.
The study will be created with the collection's content and made available
to participants.`,
		Example: `
Publish a collection with 100 participants:

$ prolific collection publish 67890abcdef --participants 100

Publish with a custom study name:

$ prolific collection publish 67890abcdef -p 50 --name "My Custom Study"
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 1 || opts.Args[0] == "" {
				return errors.New("please provide a collection ID")
			}

			if opts.Participants <= 0 {
				return errors.New("please provide a valid number of participants using --participants or -p")
			}

			return publishCollection(c, opts, w)
		},
	}

	flags := cmd.Flags()
	flags.IntVarP(&opts.Participants, "participants", "p", 0, "Number of participants required (required)")
	flags.StringVarP(&opts.Name, "name", "n", "", "Study name (defaults to collection's task name)")
	flags.StringVarP(&opts.Description, "description", "d", "", "Study description (defaults to collection's task introduction)")

	return cmd
}

func publishCollection(c client.API, opts PublishOptions, w io.Writer) error {
	collectionID := opts.Args[0]

	// Fetch the collection to get default name/description
	coll, err := c.GetCollection(collectionID)
	if err != nil {
		if shared.IsFeatureNotEnabledError(err) {
			ui.RenderFeatureAccessMessage(FeatureNameAITBCollection, FeatureContactURLAITBCollection)
			return nil
		}
		return fmt.Errorf("failed to get collection: %s", err.Error())
	}

	// Use collection details as defaults if not provided
	studyName := opts.Name
	if studyName == "" && coll.TaskDetails != nil {
		studyName = coll.TaskDetails.TaskName
	}
	if studyName == "" {
		studyName = coll.Name
	}

	studyDescription := opts.Description
	if studyDescription == "" && coll.TaskDetails != nil {
		studyDescription = coll.TaskDetails.TaskIntroduction
	}
	if studyDescription == "" {
		studyDescription = fmt.Sprintf("Study for collection: %s", coll.Name)
	}

	// Create the study with collection-specific configuration
	createStudy := model.CreateStudy{
		Name:                 studyName,
		InternalName:         studyName,
		Description:          studyDescription,
		TotalAvailablePlaces: opts.Participants,
		DataCollectionMethod: model.DataCollectionMethodAITBCollection,
		DataCollectionID:     collectionID,
	}

	study, err := c.CreateStudy(createStudy)
	if err != nil {
		return fmt.Errorf("failed to create study: %s", err.Error())
	}

	// Transition the study to publish
	_, err = c.TransitionStudy(study.ID, model.TransitionStudyPublish)
	if err != nil {
		return fmt.Errorf("failed to publish study: %s", err.Error())
	}

	// Fetch the updated study to get the latest status
	study, err = c.GetStudy(study.ID)
	if err != nil {
		return fmt.Errorf("failed to get study details: %s", err.Error())
	}

	// Display the result
	fmt.Fprintln(w, studyui.RenderStudy(*study))
	fmt.Fprintf(w, "\nStudy URL: %s\n", studyui.GetStudyURL(study.ID))

	return nil
}
