package project

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

const uniqueViolationCode = "23505"

type fixedSubject struct {
	role string
}

func (f fixedSubject) Subject(context.Context, uuid.UUID) (access.Subject, error) {
	return access.Subject{Role: f.role, Active: true}, nil
}

type projectCall struct {
	method   string
	path     string
	body     string
	role     string
	viewerID uuid.UUID
}

func serveProjects(t *testing.T, call projectCall, results ...databasetest.Result) (*httptest.ResponseRecorder, *databasetest.Recorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, recorder := databasetest.Open(t, results...)
	tokens := token.NewManager(config.Token{AccessSecret: "project-module-test", AccessTTL: time.Minute, Issuer: "project-module-test"})
	viewerID := call.viewerID
	if viewerID == uuid.Nil {
		viewerID = uuid.New()
	}
	issued, err := tokens.IssueAccess(viewerID)
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	router := gin.New()
	router.Use(middleware.ErrorHandler())
	api := router.Group("/api")
	role := call.role
	if role == "" {
		role = access.RoleAdmin
	}
	api.Use(middleware.Authorization(tokens, fixedSubject{role: role}))
	NewModule(db, tokens).RegisterRoutes(api)

	request := httptest.NewRequest(call.method, call.path, strings.NewReader(call.body))
	request.Header.Set("Authorization", "Bearer "+issued.Value)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response, recorder
}

func projectFound(projectID uuid.UUID) databasetest.Result {
	return databasetest.Result{Columns: []string{"id", "status"}, Rows: [][]driver.Value{{projectID.String(), statusOngoing}}}
}

func membership(isMember bool) databasetest.Result {
	result := databasetest.Result{Columns: []string{"id"}}
	if isMember {
		result.Rows = [][]driver.Value{{uuid.NewString()}}
	}
	return result
}

func projectMissing() databasetest.Result {
	return databasetest.Result{Columns: []string{"id", "status"}}
}

func memberUser(active bool, roleID uuid.UUID) databasetest.Result {
	return databasetest.Result{Columns: []string{"id", "active", "role_id"}, Rows: [][]driver.Value{{uuid.NewString(), active, roleID.String()}}}
}

func membersPath(projectID uuid.UUID, rest string) string {
	return "/api/projects/" + projectID.String() + rest
}

func requireStatus(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("got status %d, want %d: %s", response.Code, status, response.Body.String())
	}
	if code != "" && !strings.Contains(response.Body.String(), `"code":"`+code+`"`) {
		t.Fatalf("got body %s, want error code %s", response.Body.String(), code)
	}
}

func TestListMembersReturnsTheTeam(t *testing.T) {
	projectID := uuid.New()
	memberID := uuid.New()
	response, _ := serveProjects(t, projectCall{method: http.MethodGet, path: membersPath(projectID, "/members")},
		projectFound(projectID),
		databasetest.Result{Columns: memberRowColumns, Rows: [][]driver.Value{memberRowValues(memberID, projectID, nil)}},
	)

	requireStatus(t, response, http.StatusOK, "")
	var body struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("got %d members", len(body.Data))
	}
	member := body.Data[0]
	if member["id"] != memberID.String() || member["name"] != "Rina Wulandari" || member["email"] != "rina@banguninaja.id" {
		t.Fatalf("got member %v", member)
	}
	if member["position"] != nil || member["joinedAt"] == nil {
		t.Fatalf("got position %v joinedAt %v", member["position"], member["joinedAt"])
	}
	role, ok := member["role"].(map[string]any)
	if !ok || role["code"] != "manajer_proyek" || role["name"] != "Manajer Proyek" {
		t.Fatalf("got role %v", member["role"])
	}
}

func TestMemberRoutesAnswerWithTheRightStatus(t *testing.T) {
	projectID := uuid.New()
	roleID := uuid.New()
	duplicate := databasetest.Result{Err: &pgconn.PgError{Code: uniqueViolationCode}}
	validBody := `{"userId":"` + uuid.NewString() + `","position":"Estimator"}`

	cases := []struct {
		name    string
		call    projectCall
		results []databasetest.Result
		status  int
		code    string
	}{
		{"bad project id", projectCall{method: http.MethodGet, path: "/api/projects/bukan-uuid/members"}, nil, http.StatusBadRequest, "invalid_id"},
		{"list of missing project", projectCall{method: http.MethodGet, path: membersPath(projectID, "/members")}, []databasetest.Result{projectMissing()}, http.StatusNotFound, "project_not_found"},
		{"add without user", projectCall{method: http.MethodPost, path: membersPath(projectID, "/members"), body: `{"position":"Estimator"}`}, nil, http.StatusBadRequest, "invalid_payload"},
		{"add to missing project", projectCall{method: http.MethodPost, path: membersPath(projectID, "/members"), body: validBody}, []databasetest.Result{projectMissing()}, http.StatusNotFound, "project_not_found"},
		{
			"add unknown user",
			projectCall{method: http.MethodPost, path: membersPath(projectID, "/members"), body: validBody},
			[]databasetest.Result{projectFound(projectID), {Columns: []string{"id", "active", "role_id"}}},
			http.StatusUnprocessableEntity, "member_user_not_found",
		},
		{
			"add inactive user",
			projectCall{method: http.MethodPost, path: membersPath(projectID, "/members"), body: validBody},
			[]databasetest.Result{projectFound(projectID), memberUser(false, roleID)},
			http.StatusUnprocessableEntity, "member_user_inactive",
		},
		{
			"add existing member",
			projectCall{method: http.MethodPost, path: membersPath(projectID, "/members"), body: validBody},
			[]databasetest.Result{projectFound(projectID), memberUser(true, roleID), duplicate},
			http.StatusConflict, "project_member_exists",
		},
		{
			"update missing member",
			projectCall{method: http.MethodPatch, path: membersPath(projectID, "/members/"+uuid.NewString()), body: `{"position":"Site manager"}`},
			[]databasetest.Result{{Columns: []string{"id"}}},
			http.StatusNotFound, "project_member_not_found",
		},
		{
			"update with too long position",
			projectCall{method: http.MethodPatch, path: membersPath(projectID, "/members/"+uuid.NewString()), body: `{"position":"` + strings.Repeat("a", 81) + `"}`},
			nil, http.StatusBadRequest, "invalid_payload",
		},
		{
			"remove missing member",
			projectCall{method: http.MethodDelete, path: membersPath(projectID, "/members/"+uuid.NewString())},
			[]databasetest.Result{{Affected: 0}},
			http.StatusNotFound, "project_member_not_found",
		},
		{
			"remove member",
			projectCall{method: http.MethodDelete, path: membersPath(projectID, "/members/"+uuid.NewString())},
			[]databasetest.Result{{Affected: 1}},
			http.StatusNoContent, "",
		},
		{
			"candidates of missing project",
			projectCall{method: http.MethodGet, path: membersPath(projectID, "/member-candidates?search=bud")},
			[]databasetest.Result{projectMissing()},
			http.StatusNotFound, "project_not_found",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response, recorder := serveProjects(t, tc.call, tc.results...)
			requireStatus(t, response, tc.status, tc.code)
			if tc.results == nil && len(recorder.Queries()) != 0 {
				t.Fatalf("a rejected request must not reach the database, got %+v", recorder.Queries())
			}
		})
	}
}

func TestAddMemberAnswersCreatedWithTheNewMember(t *testing.T) {
	projectID := uuid.New()
	roleID := uuid.New()
	response, recorder := serveProjects(t,
		projectCall{method: http.MethodPost, path: membersPath(projectID, "/members"), body: `{"userId":"` + uuid.NewString() + `","position":" Site manager "}`},
		projectFound(projectID),
		memberUser(true, roleID),
		databasetest.Result{Affected: 1},
		databasetest.Result{Columns: memberRowColumns, Rows: [][]driver.Value{memberRowValues(uuid.New(), projectID, "Site manager")}},
	)

	requireStatus(t, response, http.StatusCreated, "")
	if !strings.Contains(response.Body.String(), `"position":"Site manager"`) {
		t.Fatalf("got body %s", response.Body.String())
	}
	insert := recorder.Queries()[2]
	requireFragments(t, insert.SQL, `INSERT INTO "project_member"`)
	found := false
	for _, arg := range insert.Args {
		found = found || arg == roleID.String() || arg == roleID
	}
	if !found {
		t.Fatalf("member must take the account role %s, got args %v", roleID, insert.Args)
	}
}

func TestMemberCandidatesAnswerAPage(t *testing.T) {
	projectID := uuid.New()
	response, _ := serveProjects(t,
		projectCall{method: http.MethodGet, path: membersPath(projectID, "/member-candidates?search=bud&pageSize=20")},
		projectFound(projectID),
		databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(1)}}},
		databasetest.Result{Columns: []string{"id", "name", "email"}, Rows: [][]driver.Value{{uuid.NewString(), "Budi Santoso", "budi@banguninaja.id"}}},
	)

	requireStatus(t, response, http.StatusOK, "")
	var body struct {
		Data struct {
			Items      []map[string]any `json:"items"`
			TotalItems int              `json:"totalItems"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.TotalItems != 1 || len(body.Data.Items) != 1 || len(body.Data.Items[0]) != 3 {
		t.Fatalf("got %+v, want one candidate with id, name, and email only", body.Data)
	}
}

func TestProjectListMineUsesTheCallerFromTheToken(t *testing.T) {
	viewerID := uuid.New()
	response, recorder := serveProjects(t,
		projectCall{method: http.MethodGet, path: "/api/projects?mine=true", viewerID: viewerID},
		databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(0)}}},
		databasetest.Result{Columns: []string{"id"}},
	)

	requireStatus(t, response, http.StatusOK, "")
	query := recorder.Queries()[0]
	requireFragments(t, query.SQL, "project_member WHERE user_id = $1")
	if query.Args[0] != viewerID.String() && query.Args[0] != viewerID {
		t.Fatalf("got args %v, want caller %s", query.Args, viewerID)
	}
}

func TestMemberWritesNeedProjectWriteAccess(t *testing.T) {
	projectID := uuid.New()
	writes := []projectCall{
		{method: http.MethodPost, path: membersPath(projectID, "/members"), body: `{"userId":"` + uuid.NewString() + `"}`},
		{method: http.MethodPatch, path: membersPath(projectID, "/members/"+uuid.NewString()), body: `{"position":"Estimator"}`},
		{method: http.MethodDelete, path: membersPath(projectID, "/members/"+uuid.NewString())},
	}
	for _, call := range writes {
		call.role = access.RoleSales
		response, recorder := serveProjects(t, call)
		requireStatus(t, response, http.StatusForbidden, "forbidden")
		if len(recorder.Queries()) != 0 {
			t.Fatalf("%s %s: a forbidden write must not reach the database", call.method, call.path)
		}
	}

	reads := []projectCall{
		{method: http.MethodGet, path: membersPath(projectID, "/members"), role: access.RoleSales},
		{method: http.MethodGet, path: membersPath(projectID, "/member-candidates"), role: access.RoleSales},
	}
	for _, call := range reads {
		response, _ := serveProjects(t, call, membership(true), projectMissing())
		requireStatus(t, response, http.StatusNotFound, "project_not_found")
	}
}
