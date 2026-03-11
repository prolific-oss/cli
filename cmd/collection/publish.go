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
	"github.com/spf13/viper"
)

// PublishOptions is the options for the publish collection command.
type PublishOptions struct {
	Args         []string
	Participants int
	Name         string
	Description  string
	TemplatePath string
	Draft        bool
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
to participants.

Use --draft (-d) to create the study without publishing it. The study will
remain in draft/unpublished status, allowing you to review it before
publishing with 'prolific study transition <study-id> -a PUBLISH'.

You can either specify the number of participants directly, or provide a study
template file. When using a template, the collection ID will be automatically
set as the data_collection_id and data_collection_method will be set to
AI_TASK_BUILDER_COLLECTION.

When using a template, CLI flags (--participants, --name, --description) will
override the corresponding template values.`,
		Example: `
Publish a collection with 100 participants:

$ prolific collection publish 67890abcdef --participants 100

Publish with a custom study name:

$ prolific collection publish 67890abcdef -p 50 --name "My Custom Study"

Create a draft study (not published):

$ prolific collection publish 67890abcdef -p 100 --draft

Create a draft study using the -d shorthand:

$ prolific collection publish 67890abcdef -p 100 -d

Publish using a study template file:

$ prolific collection publish 67890abcdef -t /path/to/study-template.json

Publish using a template but override the participant count:

$ prolific collection publish 67890abcdef -t /path/to/template.json -p 200
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if len(opts.Args) < 1 || opts.Args[0] == "" {
				return errors.New("please provide a collection ID")
			}

			if opts.TemplatePath == "" && opts.Participants <= 0 {
				return errors.New("please provide a valid number of participants using --participants or -p, or provide a template file using --template or -t")
			}

			return publishCollection(c, opts, w)
		},
	}

	flags := cmd.Flags()
	flags.IntVarP(&opts.Participants, "participants", "p", 0, "Number of participants required (required if no template)")
	flags.StringVarP(&opts.Name, "name", "n", "", "Study name (defaults to collection's task name)")
	flags.StringVar(&opts.Description, "description", "", "Study description (defaults to collection's task introduction)")
	flags.BoolVarP(&opts.Draft, "draft", "d", false, "Create the study in draft status without publishing")
	flags.StringVarP(&opts.TemplatePath, "template", "t", "", "Path to a study template file (JSON/YAML) - collection ID and method will be set automatically")

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

	var createStudy model.CreateStudy

	if opts.TemplatePath != "" {
		// Load study configuration from template file
		v := viper.New()
		v.SetConfigFile(opts.TemplatePath)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read template file: %s", err.Error())
		}

		if err := v.Unmarshal(&createStudy); err != nil {
			return fmt.Errorf("failed to parse template file: %s", err.Error())
		}

		// Override collection-specific fields
		createStudy.DataCollectionMethod = model.DataCollectionMethodAITBCollection
		createStudy.DataCollectionID = collectionID
		// Clear external_study_url as it's incompatible with data collection method
		createStudy.ExternalStudyURL = ""

		// Allow CLI flags to override template values
		if opts.Name != "" {
			createStudy.Name = opts.Name
			createStudy.InternalName = opts.Name
		}
		if opts.Description != "" {
			createStudy.Description = opts.Description
		}
		if opts.Participants > 0 {
			createStudy.TotalAvailablePlaces = opts.Participants
		}

		// Use collection's task introduction as description if not provided in template or flags
		if createStudy.Description == "" && coll.TaskDetails != nil {
			createStudy.Description = coll.TaskDetails.TaskIntroduction
		}
		// Final fallback if still no description
		if createStudy.Description == "" {
			createStudy.Description = fmt.Sprintf("Study for collection: %s", coll.Name)
		}
	} else {
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
		createStudy = model.CreateStudy{
			Name:                 studyName,
			InternalName:         studyName,
			Description:          studyDescription,
			TotalAvailablePlaces: opts.Participants,
			DataCollectionMethod: model.DataCollectionMethodAITBCollection,
			DataCollectionID:     collectionID,
		}
	}

	study, err := c.CreateStudy(createStudy)
	if err != nil {
		return fmt.Errorf("failed to create study: %s", err.Error())
	}

	if opts.Draft {
		fmt.Fprintln(w, studyui.RenderStudy(*study))
		fmt.Fprintf(w, "\nStudy created in draft status. Study ID: %s\n", study.ID)
		fmt.Fprintf(w, "Study URL: %s\n", studyui.GetStudyURL(study.ID))
		fmt.Fprintln(w, "\nTo publish this study, run:")
		fmt.Fprintf(w, "  prolific study transition %s -a PUBLISH\n", study.ID)
		return nil
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

	fmt.Fprintln(w, studyui.RenderStudy(*study))
	fmt.Fprintf(w, "\nStudy URL: %s\n", studyui.GetStudyURL(study.ID))

	return nil
}
