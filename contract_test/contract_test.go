// Package contracttest verifies that the client package matches the Prolific OpenAPI spec.
//
// Two tests:
//   - TestAPICoverage: every operationId in openapi.yaml must appear in the operations table
//     (as a concrete client call or with an explicit skip reason).
//   - TestClientMatchesAPISpec: uses kin-openapi to validate the HTTP request that each client
//     method sends against the spec schema (method, path, query params, body fields/types).
//
// Skip tags used in the operations table:
//   - OUTOFSCOPE   — no CLI command exists or is planned for this endpoint
//   - SPECMISMATCH — client request diverges from the spec; needs a fix

package contracttest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"gopkg.in/yaml.v3"
)

const specURL = "https://docs.prolific.com/openapi.yaml"
const specFile = "openapi.yaml"

func TestMain(m *testing.M) {
	if err := downloadSpec(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to download spec: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func downloadSpec() error {
	_ = os.Remove(specFile) // best-effort cleanup before re-downloading
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, specURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", specURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: unexpected status %s", specURL, resp.Status)
	}
	f, err := os.Create(specFile)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// findRoute locates the matching OpenAPI route for r, tolerating trailing-slash
// differences between the client and the spec (the API accepts both forms).
func findRoute(router routers.Router, r *http.Request) (*routers.Route, map[string]string, error) {
	route, params, err := router.FindRoute(r)
	if err == nil {
		return route, params, nil
	}
	// Retry with trailing slash toggled.
	r2 := r.Clone(context.Background())
	if strings.HasSuffix(r.URL.Path, "/") {
		r2.URL.Path = strings.TrimRight(r.URL.Path, "/")
		if r2.URL.Path == "" {
			r2.URL.Path = "/"
		}
	} else {
		r2.URL.Path = r.URL.Path + "/"
	}
	return router.FindRoute(r2)
}

type operation struct {
	operationID string
	call        func(*client.Client) // nil when skipped
	skip        string               // required reason when call is nil; prefix with OUTOFSCOPE or SPECMISMATCH
}

// operations maps every operationId in openapi.yaml to either a client method call
// or an explicit skip reason. Any operationId missing from this table will cause
// TestAPICoverage to fail.
var operations = []operation{
	// Workspaces
	{operationID: "get-workspaces", call: func(c *client.Client) { c.GetWorkspaces(10, 0) }},
	{operationID: "create-workspace", call: func(c *client.Client) { c.CreateWorkspace(model.Workspace{Title: "t"}) }},
	{operationID: "get-workspace", skip: "OUTOFSCOPE: no CLI command for retrieving a single workspace by ID"},
	{operationID: "update-workspace", skip: "OUTOFSCOPE: no CLI command for updating a workspace"},
	{operationID: "get-workspace-balance", call: func(c *client.Client) { c.GetWorkspaceBalance("ws-id") }},

	// Projects
	{operationID: "get-projects", call: func(c *client.Client) { c.GetProjects("ws-id", 10, 0) }},
	{operationID: "create-project", call: func(c *client.Client) { c.CreateProject("ws-id", model.Project{Title: "t"}) }},
	{operationID: "get-project", call: func(c *client.Client) { c.GetProject("proj-id") }},
	{operationID: "update-project", skip: "OUTOFSCOPE: no CLI command for updating a project"},

	// Filters
	{operationID: "get-filters", call: func(c *client.Client) { c.GetFilters() }},
	{operationID: "get-filter-distribution", skip: "OUTOFSCOPE: filter distribution not needed in CLI"},
	{operationID: "get-eligible-count", skip: "OUTOFSCOPE: eligibility count not needed in CLI"},

	// Filter Sets
	{operationID: "get-filter-sets", call: func(c *client.Client) { c.GetFilterSets("ws-id", 10, 0) }},
	{operationID: "create-filter-set", call: func(c *client.Client) { c.CreateFilterSet(model.CreateFilterSet{}) }},
	{operationID: "get-filter-set", call: func(c *client.Client) { c.GetFilterSet("fs-id") }},
	{operationID: "delete-filter-set", skip: "OUTOFSCOPE: no CLI command for deleting a filter set"},
	{operationID: "update-filter-set", skip: "OUTOFSCOPE: no CLI command for updating a filter set"},
	{operationID: "clone-filter-set", skip: "OUTOFSCOPE: no CLI command for cloning a filter set"},
	{operationID: "lock-filter-set", skip: "OUTOFSCOPE: no CLI command for locking a filter set"},
	{operationID: "unlock-filter-set", skip: "OUTOFSCOPE: no CLI command for unlocking a filter set"},

	// Webhooks
	{operationID: "get-event-types", call: func(c *client.Client) { c.GetHookEventTypes() }},
	{operationID: "get-secrets", call: func(c *client.Client) { c.GetHookSecrets("ws-id") }},
	{operationID: "create-secret", call: func(c *client.Client) {
		c.CreateHookSecret(client.CreateSecretPayload{WorkspaceID: "ws-id"})
	}},
	{operationID: "get-subscriptions", call: func(c *client.Client) { c.GetHooks("ws-id", true, 10, 0) }},
	{operationID: "create-subscription", call: func(c *client.Client) {
		c.CreateHookSubscription(client.CreateHookPayload{
			EventType:   "submission.completed",
			TargetURL:   "https://example.com/hook",
			WorkspaceID: "ws-id",
		})
	}},
	{operationID: "get-subscription", skip: "OUTOFSCOPE: no CLI command for retrieving a single webhook subscription"},
	{operationID: "confirm-subscription", call: func(c *client.Client) {
		c.ConfirmHookSubscription("sub-id", "secret-value")
	}},
	{operationID: "delete-subscription", call: func(c *client.Client) { c.DeleteHookSubscription("sub-id") }},
	{operationID: "update-subscription", call: func(c *client.Client) {
		c.UpdateHookSubscription("sub-id", client.UpdateHookPayload{})
	}},
	{operationID: "get-events", call: func(c *client.Client) { c.GetEvents("sub-id", 10, 0) }},

	// Surveys
	{operationID: "get-surveys", call: func(c *client.Client) { c.GetSurveys("researcher-id", 10, 0) }},
	{operationID: "create-survey", call: func(c *client.Client) {
		c.CreateSurvey(model.CreateSurvey{Title: "t", ResearcherID: "researcher-id"})
	}},
	{operationID: "get-survey", call: func(c *client.Client) { c.GetSurvey("survey-id") }},
	{operationID: "delete-survey", call: func(c *client.Client) { c.DeleteSurvey("survey-id") }},
	{operationID: "get-responses", call: func(c *client.Client) { c.GetSurveyResponses("survey-id", 10, 0) }},
	{operationID: "create-response", call: func(c *client.Client) {
		c.CreateSurveyResponse("survey-id", model.CreateSurveyResponseRequest{
			ParticipantID: "participant-id",
			SubmissionID:  "submission-id",
		})
	}},
	{operationID: "delete-responses", call: func(c *client.Client) { c.DeleteAllSurveyResponses("survey-id") }},
	{operationID: "get-summary", call: func(c *client.Client) { c.GetSurveyResponseSummary("survey-id") }},
	{operationID: "get-response", call: func(c *client.Client) { c.GetSurveyResponse("survey-id", "response-id") }},
	{operationID: "delete-response", call: func(c *client.Client) { c.DeleteSurveyResponse("survey-id", "response-id") }},

	// AI Task Builder — Batches
	{operationID: "get-task-builder-batches", call: func(c *client.Client) { c.GetAITaskBuilderBatches("ws-id") }},
	{operationID: "create-task-builder-batch", call: func(c *client.Client) {
		c.CreateAITaskBuilderBatch(client.CreateBatchParams{
			Name:        "t",
			WorkspaceID: "ws-id",
			DatasetID:   "ds-id",
		})
	}},
	{operationID: "get-task-builder-batch", call: func(c *client.Client) { c.GetAITaskBuilderBatch("batch-id") }},
	{operationID: "update-task-builder-batch", call: func(c *client.Client) {
		c.UpdateAITaskBuilderBatch(client.UpdateBatchParams{BatchID: "batch-id", Name: "t"})
	}},
	{operationID: "get-task-builder-batch-status", call: func(c *client.Client) { c.GetAITaskBuilderBatchStatus("batch-id") }},
	{operationID: "setup-task-builder-batch", call: func(c *client.Client) {
		c.SetupAITaskBuilderBatch("batch-id", "ds-id", 5)
	}},
	{operationID: "get-task-builder-batch-task-responses", call: func(c *client.Client) { c.GetAITaskBuilderResponses("batch-id") }},
	{operationID: "get-task-builder-batch-report", skip: "OUTOFSCOPE: no CLI command for batch report; GetAITaskBuilderTasks calls /tasks which is not in the spec"},
	{operationID: "duplicate-task-builder-batch", skip: "OUTOFSCOPE: no CLI command for duplicating a batch"},
	{operationID: "request-batch-export", call: func(c *client.Client) { c.InitiateBatchExport("batch-id") }},
	{operationID: "get-batch-export-status", call: func(c *client.Client) { c.GetBatchExportStatus("batch-id", "export-id") }},

	// AI Task Builder — Datasets
	{operationID: "create-task-builder-dataset", call: func(c *client.Client) {
		c.CreateAITaskBuilderDataset("ws-id", client.CreateAITaskBuilderDatasetPayload{Name: "t"})
	}},
	{operationID: "get-dataset-upload-url", call: func(c *client.Client) { c.GetAITaskBuilderDatasetUploadURL("ds-id", "data") }},
	{operationID: "get-task-builder-dataset-status", call: func(c *client.Client) { c.GetAITaskBuilderDatasetStatus("ds-id") }},

	// AI Task Builder — Instructions
	{operationID: "get-task-builder-instructions", skip: "OUTOFSCOPE: no CLI command for getting task builder instructions"},
	{operationID: "create-task-builder-instructions", call: func(c *client.Client) {
		c.CreateAITaskBuilderInstructions("batch-id", client.CreateAITaskBuilderInstructionsPayload{
			Instructions: []client.Instruction{},
		})
	}},
	{operationID: "update-task-builder-instructions", skip: "OUTOFSCOPE: no CLI command for updating task builder instructions"},

	// AI Task Builder — Collections
	{operationID: "list-collections", call: func(c *client.Client) { c.GetCollections("ws-id", 10, 0) }},
	{operationID: "create-collection", call: func(c *client.Client) {
		c.CreateAITaskBuilderCollection(model.CreateAITaskBuilderCollection{
			WorkspaceID:     "ws-id",
			Name:            "t",
			CollectionItems: []model.CollectionPage{},
			TaskDetails:     &model.TaskDetails{TaskName: "t", TaskIntroduction: "t", TaskSteps: "t"},
		})
	}},
	{operationID: "get-collection", call: func(c *client.Client) { c.GetCollection("coll-id") }},
	{operationID: "update-collection", call: func(c *client.Client) {
		c.UpdateCollection("coll-id", model.UpdateCollection{
			Name:            "t",
			CollectionItems: []model.Page{},
			TaskDetails:     &model.TaskDetails{TaskName: "t", TaskIntroduction: "t", TaskSteps: "t"},
		})
	}},
	{operationID: "get-collection-responses", skip: "OUTOFSCOPE: no CLI command for getting collection responses"},
	{operationID: "request-collection-export", call: func(c *client.Client) { c.InitiateCollectionExport("coll-id") }},
	{operationID: "get-collection-export-status", call: func(c *client.Client) {
		c.GetCollectionExportStatus("coll-id", "export-id")
	}},

	// Invitations
	{operationID: "create-invitation", call: func(c *client.Client) {
		c.CreateInvitation(model.CreateInvitation{
			Association: "ws-id",
			Emails:      []string{"user@example.com"},
			Role:        "WORKSPACE_COLLABORATOR",
		})
	}},

	// Messages
	{operationID: "get-messages", call: func(c *client.Client) {
		uid := "user-id"
		c.GetMessages(&uid, nil)
	}},
	{operationID: "send-message", call: func(c *client.Client) { c.SendMessage("hello", "recipient-id", "study-id") }},
	{operationID: "bulk-message-participants", call: func(c *client.Client) {
		c.BulkSendMessage([]string{"p1", "p2"}, "hello", "study-id")
	}},
	{operationID: "send-message-to-participant-group", call: func(c *client.Client) {
		c.SendGroupMessage("group-id", "hello", nil)
	}},
	{operationID: "get-unread-messages", call: func(c *client.Client) { c.GetUnreadMessages() }},

	// Studies
	{operationID: "get-studies", call: func(c *client.Client) { c.GetStudies("", "") }},
	{operationID: "create-study", call: func(c *client.Client) {
		c.CreateStudy(model.CreateStudy{
			Name:                    "t",
			ProlificIDOption:        "url_parameters",
			TotalAvailablePlaces:    10,
			EstimatedCompletionTime: 5,
			Reward:                  150,
			DeviceCompatibility:     []string{"desktop"},
		})
	}},
	{operationID: "get-project-studies", call: func(c *client.Client) { c.GetStudies("", "proj-id") }},
	{operationID: "get-study", call: func(c *client.Client) { c.GetStudy("study-id") }},
	{operationID: "delete-study", skip: "OUTOFSCOPE: no CLI command for deleting a study"},
	{operationID: "update-study", call: func(c *client.Client) {
		c.UpdateStudy("study-id", map[string]any{"name": "updated"})
	}},
	{operationID: "publish-study", call: func(c *client.Client) { c.TransitionStudy("study-id", "PUBLISH") }},
	{operationID: "create-test-study", call: func(c *client.Client) { c.TestStudy("study-id") }},
	{operationID: "get-study-access-details-progress", skip: "OUTOFSCOPE: no CLI command for access details progress"},
	{operationID: "get-study-cost", skip: "OUTOFSCOPE: no CLI command for getting study cost"},
	{operationID: "get-study-submissions", call: func(c *client.Client) { c.GetSubmissions("study-id", 10, 0) }},
	{operationID: "count-study-submissions-by-status", call: func(c *client.Client) { c.GetStudySubmissionCounts("study-id") }},
	{operationID: "download-study-credential-report", call: func(c *client.Client) {
		c.GetStudyCredentialsUsageReportCSV("study-id")
	}},
	{operationID: "export-study", skip: "OUTOFSCOPE: no CLI command for exporting a study as a whole"},
	{operationID: "export-demographic-data", skip: "SPECMISMATCH: spec requires a request body, client sends none"},
	{operationID: "get-demographic-export-history", skip: "OUTOFSCOPE: no CLI command for demographic export history"},
	{operationID: "duplicate-study", skip: "SPECMISMATCH: spec requires a request body, client sends none"},
	{operationID: "get-study-predicted-recruitment-time", skip: "OUTOFSCOPE: no CLI command for predicted recruitment time"},
	{operationID: "post-study-predicted-recruitment-time", skip: "OUTOFSCOPE: no CLI command for posting predicted recruitment time"},
	{operationID: "calculate-study-cost", skip: "OUTOFSCOPE: no CLI command for calculating study cost"},

	// Credentials
	{operationID: "list-credential-pools", call: func(c *client.Client) { c.ListCredentialPools("ws-id") }},
	{operationID: "create-credential-pool", call: func(c *client.Client) {
		c.CreateCredentialPool("user,pass\nuser2,pass2", "ws-id")
	}},
	{operationID: "update-credential-pool", call: func(c *client.Client) {
		c.UpdateCredentialPool("pool-id", "user,pass\nuser2,pass2")
	}},

	// Reward Recommendations
	{operationID: "calculate-reward-recommendations", skip: "OUTOFSCOPE: reward recommendations not needed in CLI"},

	// Well-known endpoints
	{operationID: "get-study-jwks", skip: "OUTOFSCOPE: JWKS endpoint not needed in CLI"},

	// Submissions
	{operationID: "get-submissions", skip: "OUTOFSCOPE: global submissions list not exposed in CLI; use study-scoped get-study-submissions"},
	{operationID: "get-submission", skip: "OUTOFSCOPE: single submission retrieval not exposed in CLI"},
	{operationID: "transition-submission", call: func(c *client.Client) {
		c.TransitionSubmission("sub-id", client.TransitionSubmissionPayload{Action: "APPROVE"})
	}},
	{operationID: "request-submission-return", call: func(c *client.Client) {
		c.RequestSubmissionReturn("sub-id", []string{"no longer needed"})
	}},
	{operationID: "get-submission-demographics", skip: "OUTOFSCOPE: submission demographics not exposed in CLI"},
	{operationID: "bulk-approve-submissions", call: func(c *client.Client) {
		c.BulkApproveSubmissions(client.BulkApproveSubmissionsPayload{
			SubmissionIDs: []string{"sub-id"},
		})
	}},

	// Bonuses
	{operationID: "create-bonus-payments", call: func(c *client.Client) {
		c.CreateBonusPayments(client.CreateBonusPaymentsPayload{
			StudyID:    "study-id",
			CSVBonuses: "participant-id,1.50",
		})
	}},
	{operationID: "pay-bonus-payments", call: func(c *client.Client) { c.PayBonusPayments("bonus-id") }},

	// Users
	{operationID: "get-user", call: func(c *client.Client) { c.GetMe() }},
	{operationID: "create-test-participant-for-researcher", call: func(c *client.Client) {
		c.CreateTestParticipant("test@example.com")
	}},

	// Participant Groups
	{operationID: "get-participant-groups", skip: "SPECMISMATCH: spec uses a filter deepObject query param, client sends it flat"},
	{operationID: "create-participant-group", call: func(c *client.Client) {
		c.CreateParticipantGroup(model.CreateParticipantGroup{Name: "t", WorkspaceID: "ws-id"})
	}},
	{operationID: "get-participant-group", skip: "OUTOFSCOPE: no CLI command for retrieving a single participant group by ID"},
	{operationID: "delete-participant-group", skip: "OUTOFSCOPE: no CLI command for deleting a participant group"},
	{operationID: "update-participant-group", skip: "OUTOFSCOPE: no CLI command for updating a participant group"},
	{operationID: "get-participant-group-participants", call: func(c *client.Client) { c.GetParticipantGroup("group-id") }},
	{operationID: "add-to-participant-group", skip: "OUTOFSCOPE: adding participants not exposed in CLI"},
	{operationID: "remove-from-participant-group", call: func(c *client.Client) {
		c.RemoveParticipantGroupMembers("group-id", []string{"participant-id"})
	}},
}

// TestAPICoverage verifies that every operationId in openapi.yaml is accounted for
// in the operations table — either as a concrete client call or with an explicit
// skip reason. Fails when a new endpoint is added to the spec without updating this table.
func TestAPICoverage(t *testing.T) {
	data, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("failed to read openapi.yaml: %v", err)
	}

	var spec struct {
		Paths map[string]map[string]struct {
			OperationID string `yaml:"operationId"`
		} `yaml:"paths"`
	}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatalf("failed to parse openapi.yaml: %v", err)
	}

	specOps := map[string]bool{}
	for _, methods := range spec.Paths {
		for _, op := range methods {
			if op.OperationID != "" {
				specOps[op.OperationID] = true
			}
		}
	}

	covered := map[string]bool{}
	for _, op := range operations {
		covered[op.operationID] = true
		if op.call == nil && op.skip == "" {
			t.Errorf("operation %q has neither a call nor a skip reason", op.operationID)
		}
	}

	for id := range specOps {
		if !covered[id] {
			t.Errorf("operationId %q is in openapi.yaml but not in the coverage table — add a call or skip reason", id)
		}
	}

	for _, op := range operations {
		if !specOps[op.operationID] {
			t.Errorf("operationId %q is in the coverage table but not in openapi.yaml — stale entry", op.operationID)
		}
	}
}

// TestClientMatchesAPISpec validates that each in-scope client method sends an HTTP
// request that conforms to the OpenAPI spec: correct method, path, query params, and
// request body schema. Client response parsing errors are ignored — only the request
// structure matters.
func TestClientMatchesAPISpec(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("openapi.yaml")
	if err != nil {
		t.Fatalf("failed to load openapi.yaml: %v", err)
	}

	var (
		errCh  chan error
		router routers.Router
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, pathParams, routeErr := findRoute(router, r)
		if routeErr != nil {
			errCh <- fmt.Errorf("route not found for %s %s: %v", r.Method, r.URL.Path, routeErr)
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "{}")
			return
		}

		input := &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			},
		}
		errCh <- openapi3filter.ValidateRequest(context.Background(), input)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(struct{}{}); err != nil {
			t.Logf("failed to encode response: %v", err)
		}
	}))
	defer srv.Close()

	doc.Servers = openapi3.Servers{{URL: srv.URL}}
	router, err = gorillamux.NewRouter(doc)
	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	for _, tt := range operations {
		if tt.skip != "" {
			continue
		}
		t.Run(tt.operationID, func(t *testing.T) {
			errCh = make(chan error, 1)

			c := client.Client{
				Client:  srv.Client(),
				BaseURL: srv.URL,
				Token:   "test-token",
			}
			tt.call(&c) // client errors from status-code checks are intentionally ignored

			if err := <-errCh; err != nil {
				t.Errorf("request validation failed: %v", err)
			}
		})
	}
}
