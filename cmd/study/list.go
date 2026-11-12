package study

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/shared"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// defaultListFields is the default fields shown when the user has not specified --fields.
const defaultListFields = "ID,Name,Status"

type ListOptions struct {
	Args        []string
	Fields      string
	Output      shared.OutputOptions
	Raw         bool
	ProjectID   string
	Status      string
	Underpaying bool
}

// NewListCommand creates a new `study list` command to give you details about
// your studies.
func NewListCommand(commandName string, client client.API, w io.Writer) *cobra.Command {
	var opts ListOptions

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "List all of your studies",
		Long: `List your studies

This command allows you to understand what is happening with your studies on the
Prolific Platform.

Use --raw to output the full API response as JSON, including fields unknown to
the CLI and the API's pagination metadata. This returns the same single response
as a normal listing; it does not fetch additional pages. --status and --project
still apply. --raw cannot be combined with --underpaying, --fields, or table/CSV
output.`,
		Example: `
You can list all your studies in an interactive manner. This is a searchable
interface. When you have found the study you want to look into in more detail,
press enter.
$ prolific study list

You can output as a table, useful for terminal, or into another application
$ prolific study list --table
$ prolific study list -t

You can output as CSV
$ prolific study list --csv
$ prolific study list -c

You can output as JSON
$ prolific study list --json
$ prolific study list -j

You can output the full API response, preserving all fields returned by the API
$ prolific study list --raw
$ prolific study list --json --raw -p 6261321e223a605c7a4f7561

You can specify the fields you want to render in table or CSV output
$ prolific study list -f ID,InternalName,TotalCost -t
$ prolific study list -f ID,InternalName,TotalCost -c

You can filter the studies by the project they are assigned to
$ prolific study list -p 6261321e223a605c7a4f7561

You can filter the studies by their status, for example your active studies
$ prolific study list -s active

The fields you can use are
- ID
- Name
- InternalName
- DateCreated
- TotalAvailablePlaces
- Reward
- CanAutoReview
- Desc
- EstimatedCompletionTime
- MaximumAllowedTime
- CompletionURL
- ExternalStudyURL
- PublishedAt
- StartedPublishingAt
- AwardPoints
- PresentmentCurrencyCode
- Status
- AverageRewardPerHour
- DeviceCompatibility
- PeripheralRequirements
- PlacesTaken
- EstimatedRewardPerHour
- Ref
- StudyType
- TotalCost
- PublishAt
- IsPilot
- IsUnderpaying
- CredentialPoolID`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args

			if opts.Raw {
				if conflicts := rawConflictingFlags(cmd); len(conflicts) > 0 {
					return fmt.Errorf("error: --raw cannot be combined with %s", strings.Join(conflicts, ", "))
				}
				response, err := client.GetStudiesRaw(opts.Status, opts.ProjectID)
				if err != nil {
					return fmt.Errorf("error: %s", err)
				}
				encoder := json.NewEncoder(w)
				encoder.SetIndent("", "  ")
				encoder.SetEscapeHTML(false)
				if err := encoder.Encode(response); err != nil {
					return fmt.Errorf("error: %s", err)
				}
				return nil
			}

			studies, err := client.GetStudies(opts.Status, opts.ProjectID)
			if err != nil {
				return err
			}

			if opts.Underpaying {
				studies = filterByUnderpaying(*studies)
			}

			format := shared.ResolveFormat(opts.Output)
			fields := opts.Fields
			if fields == "" {
				fields = defaultListFields
			}
			switch format {
			case "json":
				r := ui.JSONRenderer[model.Study]{}
				if err := r.Render(studies.Results, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			case "csv":
				r := ui.CsvRenderer[model.Study]{}
				if err := r.Render(studies.Results, fields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			case "table":
				r := ui.TableRenderer[model.Study]{}
				if err := r.Render(studies.Results, fields, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			default:
				r := &InteractiveRenderer{}
				if err := r.Render(client, *studies, w); err != nil {
					return fmt.Errorf("error: %s", err)
				}
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Status, "status", "s", model.StatusAll, fmt.Sprintf("The status you want to filter on: %s.", strings.Join(model.StudyListStatus, ", ")))
	flags.BoolVarP(&opts.Underpaying, "underpaying", "u", false, "Filter by underpaying studies.")
	flags.StringVarP(&opts.Fields, "fields", "f", "", "Comma separated list of fields you want to display in table or csv mode.")
	flags.StringVarP(&opts.ProjectID, "project", "p", "", "Get studies for a given project ID.")
	flags.BoolVar(&opts.Raw, "raw", false, "Output the full API response as JSON, including fields unknown to the CLI")
	shared.AddOutputFlags(cmd, &opts.Output)
	for _, name := range []string{"table", "non-interactive"} {
		flag := flags.Lookup(name)
		flag.Value = &explicitBoolValue{Value: flag.Value}
	}

	return cmd
}

// explicitBoolValue records this flag's last explicit value while preserving
// the shared value used by the table aliases to choose the output format.
type explicitBoolValue struct {
	pflag.Value
	enabled bool
}

func (v *explicitBoolValue) Set(value string) error {
	if err := v.Value.Set(value); err != nil {
		return err
	}
	v.enabled = v.String() == "true"
	return nil
}

func (v *explicitBoolValue) IsBoolFlag() bool { return true }

// rawConflictingFlags returns the flags the user passed that --raw cannot honour,
// using each table alias's explicit value rather than their shared output value.
func rawConflictingFlags(cmd *cobra.Command) []string {
	flags := cmd.Flags()
	var conflicts []string
	for _, name := range []string{"csv", "table", "non-interactive", "underpaying"} {
		enabled, _ := flags.GetBool(name)
		if value, ok := flags.Lookup(name).Value.(*explicitBoolValue); ok {
			enabled = value.enabled
		}
		if enabled && flags.Changed(name) {
			conflicts = append(conflicts, "--"+name)
		}
	}
	if flags.Changed("fields") {
		conflicts = append(conflicts, "--fields")
	}
	return conflicts
}

func filterByUnderpaying(studies client.ListStudiesResponse) *client.ListStudiesResponse {
	var filtered []model.Study
	for _, study := range studies.Results {
		if study.IsUnderpaying == true {
			filtered = append(filtered, study)
		}
	}

	studies.Results = filtered
	return &studies
}

// InteractiveRenderer runs the bubbles UI framework to provide a rich
// UI experience for the user.
type InteractiveRenderer struct{}

// Render will render the list in an interactive manner.
func (r *InteractiveRenderer) Render(c client.API, studies client.ListStudiesResponse, w io.Writer) error {
	var items []list.Item
	studyMap := make(map[string]model.Study)

	for _, s := range studies.Results {
		items = append(items, s)
		studyMap[s.ID] = s
	}

	lv := ListView{
		List:    list.New(items, list.NewDefaultDelegate(), 0, 0),
		Studies: studyMap,
		Client:  c,
	}
	lv.List.Title = "My studies"

	p := tea.NewProgram(lv)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("cannot render studies: %s", err)
	}

	return nil
}

// ListView is responsible for presenting a list view to the user.
type ListView struct {
	List    list.Model
	Studies map[string]model.Study
	Study   *model.Study
	Client  client.API
}

// Init will initialise the view.
func (lv ListView) Init() tea.Cmd {
	return nil
}

// Update will update the view.
func (lv ListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return lv, tea.Quit
		}

		if msg.String() == "enter" {
			i, ok := lv.List.SelectedItem().(model.Study)
			if ok {
				lv.Study = &i
			}
			return lv, tea.Quit
		}

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		lv.List.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	lv.List, cmd = lv.List.Update(msg)
	return lv, cmd
}

// View will render the view.
func (lv ListView) View() string {
	if lv.Study != nil {
		return RenderStudy(*lv.Study)
	}
	return lv.List.View()
}
