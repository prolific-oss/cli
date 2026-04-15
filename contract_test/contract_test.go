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
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"gopkg.in/yaml.v3"
)

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
	{operationID: "workspaces_GetWorkspaces", call: func(c *client.Client) { c.GetWorkspaces(10, 0) }},
	{operationID: "workspaces_CreateWorkspace", call: func(c *client.Client) { c.CreateWorkspace(model.Workspace{Title: "t"}) }},
	{operationID: "workspaces_GetWorkspace", skip: "OUTOFSCOPE: no CLI command for retrieving a single workspace by ID"},
	{operationID: "workspaces_UpdateWorkspace", skip: "OUTOFSCOPE: no CLI command for updating a workspace"},
	{operationID: "workspaces_GetWorkspaceBalance", call: func(c *client.Client) { c.GetWorkspaceBalance("ws-id") }},

	// Projects
	{operationID: "projects_GetProjects", call: func(c *client.Client) { c.GetProjects("ws-id", 10, 0) }},
	{operationID: "projects_CreateProject", call: func(c *client.Client) { c.CreateProject("ws-id", model.Project{Title: "t"}) }},
	{operationID: "projects_GetProject", call: func(c *client.Client) { c.GetProject("proj-id") }},
	{operationID: "projects_UpdateProject", skip: "OUTOFSCOPE: no CLI command for updating a project"},

	// Filters
	{operationID: "filters_GetFilters", call: func(c *client.Client) { c.GetFilters() }},
	{operationID: "filters_GetFilterDistribution", skip: "OUTOFSCOPE: filter distribution not needed in CLI"},
	{operationID: "filters_GetEligibleCount", skip: "OUTOFSCOPE: eligibility count not needed in CLI"},

	// Filter Sets
	{operationID: "filterSets_GetFilterSets", call: func(c *client.Client) { c.GetFilterSets("ws-id", 10, 0) }},
	{operationID: "filterSets_CreateFilterSet", call: func(c *client.Client) { c.CreateFilterSet(model.CreateFilterSet{}) }},
	{operationID: "filterSets_GetFilterSet", call: func(c *client.Client) { c.GetFilterSet("fs-id") }},
	{operationID: "filterSets_DeleteFilterSet", skip: "OUTOFSCOPE: no CLI command for deleting a filter set"},
	{operationID: "filterSets_UpdateFilterSet", skip: "OUTOFSCOPE: no CLI command for updating a filter set"},
	{operationID: "filterSets_CloneFilterSet", skip: "OUTOFSCOPE: no CLI command for cloning a filter set"},
	{operationID: "filterSets_LockFilterSet", skip: "OUTOFSCOPE: no CLI command for locking a filter set"},
	{operationID: "filterSets_UnlockFilterSet", skip: "OUTOFSCOPE: no CLI command for unlocking a filter set"},

	// Webhooks
	{operationID: "webhooks_GetEventTypes", call: func(c *client.Client) { c.GetHookEventTypes() }},
	{operationID: "webhooks_GetSecrets", call: func(c *client.Client) { c.GetHookSecrets("ws-id") }},
	{operationID: "webhooks_CreateSecret", call: func(c *client.Client) {
		c.CreateHookSecret(client.CreateSecretPayload{WorkspaceID: "ws-id"})
	}},
	{operationID: "webhooks_GetSubscriptions", call: func(c *client.Client) { c.GetHooks("ws-id", true, 10, 0) }},
	{operationID: "webhooks_CreateSubscription", call: func(c *client.Client) {
		c.CreateHookSubscription(client.CreateHookPayload{
			EventType:   "submission.completed",
			TargetURL:   "https://example.com/hook",
			WorkspaceID: "ws-id",
		})
	}},
	{operationID: "webhooks_GetSubscription", skip: "OUTOFSCOPE: no CLI command for retrieving a single webhook subscription"},
	{operationID: "webhooks_ConfirmSubscription", call: func(c *client.Client) {
		c.ConfirmHookSubscription("sub-id", "secret-value")
	}},
	{operationID: "webhooks_DeleteSubscription", call: func(c *client.Client) { c.DeleteHookSubscription("sub-id") }},
	{operationID: "webhooks_UpdateSubscription", call: func(c *client.Client) {
		c.UpdateHookSubscription("sub-id", client.UpdateHookPayload{})
	}},
	{operationID: "webhooks_GetEvents", call: func(c *client.Client) { c.GetEvents("sub-id", 10, 0) }},

	// Surveys
	{operationID: "surveys_GetSurveys", call: func(c *client.Client) { c.GetSurveys("researcher-id", 10, 0) }},
	{operationID: "surveys_CreateSurvey", call: func(c *client.Client) {
		c.CreateSurvey(model.CreateSurvey{Title: "t", ResearcherID: "researcher-id"})
	}},
	{operationID: "surveys_GetSurvey", call: func(c *client.Client) { c.GetSurvey("survey-id") }},
	{operationID: "surveys_DeleteSurvey", call: func(c *client.Client) { c.DeleteSurvey("survey-id") }},
	{operationID: "surveys_GetResponses", call: func(c *client.Client) {
		c.GetSurveyResponses("survey-id", 10, 0)
	}},
	{operationID: "surveys_CreateResponse", call: func(c *client.Client) {
		c.CreateSurveyResponse("survey-id", model.CreateSurveyResponseRequest{
			ParticipantID: "participant-id",
			SubmissionID:  "submission-id",
		})
	}},
	{operationID: "surveys_DeleteResponses", call: func(c *client.Client) { c.DeleteAllSurveyResponses("survey-id") }},
	{operationID: "surveys_GetSummary", call: func(c *client.Client) { c.GetSurveyResponseSummary("survey-id") }},
	{operationID: "surveys_GetResponse", call: func(c *client.Client) {
		c.GetSurveyResponse("survey-id", "response-id")
	}},
	{operationID: "surveys_DeleteResponse", call: func(c *client.Client) {
		c.DeleteSurveyResponse("survey-id", "response-id")
	}},

	// AI Task Builder — Batches
	{operationID: "aiTaskBuilder_GetTaskBuilderBatches", call: func(c *client.Client) {
		c.GetAITaskBuilderBatches("ws-id")
	}},
	{operationID: "aiTaskBuilder_CreateTaskBuilderBatch", call: func(c *client.Client) {
		c.CreateAITaskBuilderBatch(client.CreateBatchParams{
			Name:        "t",
			WorkspaceID: "ws-id",
			DatasetID:   "ds-id",
		})
	}},
	{operationID: "aiTaskBuilder_GetTaskBuilderBatch", call: func(c *client.Client) {
		c.GetAITaskBuilderBatch("batch-id")
	}},
	{operationID: "aiTaskBuilder_UpdateTaskBuilderBatch", call: func(c *client.Client) {
		c.UpdateAITaskBuilderBatch(client.UpdateBatchParams{BatchID: "batch-id", Name: "t"})
	}},
	{operationID: "aiTaskBuilder_GetTaskBuilderBatchStatus", call: func(c *client.Client) {
		c.GetAITaskBuilderBatchStatus("batch-id")
	}},
	{operationID: "aiTaskBuilder_SetupTaskBuilderBatch", call: func(c *client.Client) {
		c.SetupAITaskBuilderBatch("batch-id", "ds-id", 5)
	}},
	{operationID: "aiTaskBuilder_GetTaskBuilderBatchTaskResponses", call: func(c *client.Client) {
		c.GetAITaskBuilderResponses("batch-id")
	}},
	{operationID: "aiTaskBuilder_GetTaskBuilderBatchReport", skip: "OUTOFSCOPE: no CLI command for batch report; GetAITaskBuilderTasks calls /tasks which is not in the spec"},
	{operationID: "aiTaskBuilder_DuplicateTaskBuilderBatch", skip: "OUTOFSCOPE: no CLI command for duplicating a batch"},
	{operationID: "aiTaskBuilder_RequestBatchExport", call: func(c *client.Client) { c.InitiateBatchExport("batch-id") }},
	{operationID: "aiTaskBuilder_GetBatchExportStatus", call: func(c *client.Client) {
		c.GetBatchExportStatus("batch-id", "export-id")
	}},

	// AI Task Builder — Datasets
	{operationID: "aiTaskBuilder_CreateTaskBuilderDataset", call: func(c *client.Client) {
		c.CreateAITaskBuilderDataset("ws-id", client.CreateAITaskBuilderDatasetPayload{Name: "t"})
	}},
	{operationID: "aiTaskBuilder_getDatasetUploadUrl", call: func(c *client.Client) {
		c.GetAITaskBuilderDatasetUploadURL("ds-id", "data")
	}},
	{operationID: "aiTaskBuilder_GetTaskBuilderDatasetStatus", call: func(c *client.Client) {
		c.GetAITaskBuilderDatasetStatus("ds-id")
	}},

	// AI Task Builder — Instructions
	{operationID: "aiTaskBuilder_GetTaskBuilderInstructions", skip: "OUTOFSCOPE: no CLI command for getting task builder instructions"},
	{operationID: "aiTaskBuilder_CreateTaskBuilderInstructions", call: func(c *client.Client) {
		c.CreateAITaskBuilderInstructions("batch-id", client.CreateAITaskBuilderInstructionsPayload{
			Instructions: []client.Instruction{},
		})
	}},
	{operationID: "aiTaskBuilder_UpdateTaskBuilderInstructions", skip: "OUTOFSCOPE: no CLI command for updating task builder instructions"},

	// AI Task Builder — Collections
	{operationID: "aiTaskBuilder_ListCollections", call: func(c *client.Client) {
		c.GetCollections("ws-id", 10, 0)
	}},
	{operationID: "aiTaskBuilder_CreateCollection", call: func(c *client.Client) {
		c.CreateAITaskBuilderCollection(model.CreateAITaskBuilderCollection{
			WorkspaceID:     "ws-id",
			Name:            "t",
			CollectionItems: []model.CollectionPage{},
			TaskDetails:     &model.TaskDetails{TaskName: "t", TaskIntroduction: "t", TaskSteps: "t"},
		})
	}},
	{operationID: "aiTaskBuilder_GetCollection", call: func(c *client.Client) { c.GetCollection("coll-id") }},
	{operationID: "aiTaskBuilder_UpdateCollection", call: func(c *client.Client) {
		c.UpdateCollection("coll-id", model.UpdateCollection{
			Name:            "t",
			CollectionItems: []model.Page{},
			TaskDetails:     &model.TaskDetails{TaskName: "t", TaskIntroduction: "t", TaskSteps: "t"},
		})
	}},
	{operationID: "aiTaskBuilder_GetCollectionResponses", skip: "OUTOFSCOPE: no CLI command for getting collection responses"},
	{operationID: "aiTaskBuilder_RequestCollectionExport", call: func(c *client.Client) { c.InitiateCollectionExport("coll-id") }},
	{operationID: "aiTaskBuilder_GetCollectionExportStatus", call: func(c *client.Client) {
		c.GetCollectionExportStatus("coll-id", "export-id")
	}},

	// Invitations
	{operationID: "invitations_CreateInvitation", call: func(c *client.Client) {
		c.CreateInvitation(model.CreateInvitation{
			Association: "ws-id",
			Emails:      []string{"user@example.com"},
			Role:        "WORKSPACE_COLLABORATOR",
		})
	}},

	// Messages
	{operationID: "messages_GetMessages", call: func(c *client.Client) {
		uid := "user-id"
		c.GetMessages(&uid, nil)
	}},
	{operationID: "messages_SendMessage", call: func(c *client.Client) {
		c.SendMessage("hello", "recipient-id", "study-id")
	}},
	{operationID: "messages_BulkMessageParticipants", call: func(c *client.Client) {
		c.BulkSendMessage([]string{"p1", "p2"}, "hello", "study-id")
	}},
	{operationID: "messages_SendMessageToParticipantGroup", call: func(c *client.Client) {
		c.SendGroupMessage("group-id", "hello", nil)
	}},
	{operationID: "messages_GetUnreadMessages", call: func(c *client.Client) { c.GetUnreadMessages() }},

	// Studies — global list and CRUD
	{operationID: "studies_GetStudies", call: func(c *client.Client) { c.GetStudies("", "") }},
	{operationID: "studies_CreateStudy", call: func(c *client.Client) {
		c.CreateStudy(model.CreateStudy{
			Name:                    "t",
			ProlificIDOption:        "url_parameters",
			TotalAvailablePlaces:    10,
			EstimatedCompletionTime: 5,
			Reward:                  150,
			DeviceCompatibility:     []string{"desktop"},
		})
	}},
	{operationID: "studies_GetProjectStudies", call: func(c *client.Client) { c.GetStudies("", "proj-id") }},
	{operationID: "studies_GetStudy", call: func(c *client.Client) { c.GetStudy("study-id") }},
	{operationID: "studies_DeleteStudy", skip: "OUTOFSCOPE: no CLI command for deleting a study"},
	{operationID: "studies_UpdateStudy", call: func(c *client.Client) {
		c.UpdateStudy("study-id", map[string]any{"name": "updated"})
	}},
	{operationID: "studies_PublishStudy", call: func(c *client.Client) {
		c.TransitionStudy("study-id", "PUBLISH")
	}},
	{operationID: "studies_CreateTestStudy", call: func(c *client.Client) { c.TestStudy("study-id") }},
	{operationID: "studies_GetStudyAccessDetailsProgress", skip: "OUTOFSCOPE: no CLI command for access details progress"},
	{operationID: "studies_GetStudyCost", skip: "OUTOFSCOPE: no CLI command for getting study cost"},
	{operationID: "studies_GetStudySubmissions", call: func(c *client.Client) {
		c.GetSubmissions("study-id", 10, 0)
	}},
	{operationID: "studies_CountStudySubmissionsByStatus", call: func(c *client.Client) {
		c.GetStudySubmissionCounts("study-id")
	}},
	{operationID: "studies_DownloadStudyCredentialReport", call: func(c *client.Client) {
		c.GetStudyCredentialsUsageReportCSV("study-id")
	}},
	{operationID: "studies_ExportStudy", skip: "OUTOFSCOPE: no CLI command for exporting a study as a whole"},
	{operationID: "studies_ExportDemographicData", skip: "SPECMISMATCH: spec requires a request body, client sends none"},
	{operationID: "studies_GetDemographicExportHistory", skip: "OUTOFSCOPE: no CLI command for demographic export history"},
	{operationID: "studies_DuplicateStudy", skip: "SPECMISMATCH: spec requires a request body, client sends none"},
	{operationID: "studies_GetStudyPredictedRecruitmentTime", skip: "OUTOFSCOPE: no CLI command for predicted recruitment time"},
	{operationID: "studies_PostStudyPredictedRecruitmentTime", skip: "OUTOFSCOPE: no CLI command for posting predicted recruitment time"},
	{operationID: "studies_CalculateStudyCost", skip: "OUTOFSCOPE: no CLI command for calculating study cost"},

	// Credentials
	{operationID: "credentials_ListCredentialPools", call: func(c *client.Client) { c.ListCredentialPools("ws-id") }},
	{operationID: "credentials_CreateCredentialPool", call: func(c *client.Client) {
		c.CreateCredentialPool("user,pass\nuser2,pass2", "ws-id")
	}},
	{operationID: "credentials_UpdateCredentialPool", call: func(c *client.Client) {
		c.UpdateCredentialPool("pool-id", "user,pass\nuser2,pass2")
	}},

	// Reward Recommendations
	{operationID: "rewardRecommendations_CalculateRewardRecommendations", skip: "OUTOFSCOPE: reward recommendations not needed in CLI"},

	// Well-known endpoints
	{operationID: "wellKnownEndpoints_getStudyJwks", skip: "OUTOFSCOPE: JWKS endpoint not needed in CLI"},

	// Submissions
	{operationID: "submissions_GetSubmissions", skip: "OUTOFSCOPE: global submissions list not exposed in CLI; use study-scoped GetSubmissions"},
	{operationID: "submissions_GetSubmission", skip: "OUTOFSCOPE: single submission retrieval not exposed in CLI"},
	{operationID: "submissions_TransitionSubmission", call: func(c *client.Client) {
		c.TransitionSubmission("sub-id", client.TransitionSubmissionPayload{Action: "APPROVE"})
	}},
	{operationID: "submissions_RequestSubmissionReturn", call: func(c *client.Client) {
		c.RequestSubmissionReturn("sub-id", []string{"no longer needed"})
	}},
	{operationID: "submissions_GetSubmissionDemographics", skip: "OUTOFSCOPE: submission demographics not exposed in CLI"},
	{operationID: "submissions_BulkApproveSubmissions", call: func(c *client.Client) {
		c.BulkApproveSubmissions(client.BulkApproveSubmissionsPayload{
			SubmissionIDs: []string{"sub-id"},
		})
	}},

	// Bonuses
	{operationID: "bonuses_CreateBonusPayments", call: func(c *client.Client) {
		c.CreateBonusPayments(client.CreateBonusPaymentsPayload{
			StudyID:    "study-id",
			CSVBonuses: "participant-id,1.50",
		})
	}},
	{operationID: "bonuses_PayBonusPayments", call: func(c *client.Client) { c.PayBonusPayments("bonus-id") }},

	// Users
	{operationID: "users_GetUser", call: func(c *client.Client) { c.GetMe() }},
	{operationID: "users_CreateTestParticipantForResearcher", call: func(c *client.Client) {
		c.CreateTestParticipant("test@example.com")
	}},

	// Participant Groups
	{operationID: "participantGroups_GetParticipantGroups", skip: "SPECMISMATCH: spec uses a filter deepObject query param, client sends it flat"},
	{operationID: "participantGroups_CreateParticipantGroup", call: func(c *client.Client) {
		c.CreateParticipantGroup(model.CreateParticipantGroup{Name: "t", WorkspaceID: "ws-id"})
	}},
	{operationID: "participantGroups_GetParticipantGroup", skip: "OUTOFSCOPE: no CLI command for retrieving a single participant group by ID"},
	{operationID: "participantGroups_DeleteParticipantGroup", skip: "OUTOFSCOPE: no CLI command for deleting a participant group"},
	{operationID: "participantGroups_UpdateParticipantGroup", skip: "OUTOFSCOPE: no CLI command for updating a participant group"},
	{operationID: "participantGroups_GetParticipantGroupParticipants", call: func(c *client.Client) {
		c.GetParticipantGroup("group-id")
	}},
	{operationID: "participantGroups_AddToParticipantGroup", skip: "OUTOFSCOPE: adding participants not exposed in CLI"},
	{operationID: "participantGroups_RemoveFromParticipantGroup", call: func(c *client.Client) {
		c.RemoveParticipantGroupMembers("group-id", []string{"participant-id"})
	}},
}

// TestAPICoverage verifies that every operationId in publicapi.yaml is accounted for
// in the operations table — either as a concrete client call or with an explicit
// skip reason. Fails when a new endpoint is added to the spec without updating this table.
func TestAPICoverage(t *testing.T) {
	data, err := os.ReadFile("publicapi.yaml")
	if err != nil {
		t.Fatalf("failed to read publicapi.yaml: %v", err)
	}

	var spec struct {
		Paths map[string]map[string]struct {
			OperationID string `yaml:"operationId"`
		} `yaml:"paths"`
	}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatalf("failed to parse publicapi.yaml: %v", err)
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
			t.Errorf("operationId %q is in publicapi.yaml but not in the coverage table — add a call or skip reason", id)
		}
	}

	for _, op := range operations {
		if !specOps[op.operationID] {
			t.Errorf("operationId %q is in the coverage table but not in publicapi.yaml — stale entry", op.operationID)
		}
	}
}

// TestClientMatchesAPISpec validates that each in-scope client method sends an HTTP
// request that conforms to the OpenAPI spec: correct method, path, query params, and
// request body schema. Client response parsing errors are ignored — only the request
// structure matters.
func TestClientMatchesAPISpec(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("publicapi.yaml")
	if err != nil {
		t.Fatalf("failed to load publicapi.yaml: %v", err)
	}

	var (
		mu     sync.Mutex
		valErr error
		router routers.Router
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		route, pathParams, routeErr := findRoute(router, r)
		if routeErr != nil {
			valErr = fmt.Errorf("route not found for %s %s: %v", r.Method, r.URL.Path, routeErr)
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
		if validateErr := openapi3filter.ValidateRequest(context.Background(), input); validateErr != nil {
			valErr = validateErr
		}

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

			valErr = nil

			c := client.Client{
				Client:  srv.Client(),
				BaseURL: srv.URL,
				Token:   "test-token",
			}
			tt.call(&c) // client errors from status-code checks are intentionally ignored

			capturedErr := valErr

			if capturedErr != nil {
				t.Errorf("request validation failed: %v", capturedErr)
			}
		})
	}
}
