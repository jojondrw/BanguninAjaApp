package project

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

var memberRowColumns = []string{
	"id", "project_id", "user_id", "name", "email", "active", "position",
	"role_code", "role_name", "created_at", "updated_at",
}

func memberRowValues(id, projectID uuid.UUID, position any) []driver.Value {
	joined := time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)
	return []driver.Value{
		id.String(), projectID.String(), uuid.NewString(), "Rina Wulandari", "rina@banguninaja.id", true, position,
		"manajer_proyek", "Manajer Proyek", joined, joined,
	}
}

func requireFragments(t *testing.T, sql string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(sql, fragment) {
			t.Errorf("query is missing %q:\n%s", fragment, sql)
		}
	}
}

func TestListMembersJoinsUserAndRoleInOneQuery(t *testing.T) {
	projectID := uuid.New()
	memberID := uuid.New()
	db, recorder := databasetest.Open(t, databasetest.Result{
		Columns: memberRowColumns,
		Rows:    [][]driver.Value{memberRowValues(memberID, projectID, "Site manager")},
	})

	rows, err := NewRepository(db).ListMembers(context.Background(), projectID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != memberID || rows[0].Name != "Rina Wulandari" || rows[0].RoleName != "Manajer Proyek" {
		t.Fatalf("got rows %+v", rows)
	}
	if rows[0].Position == nil || *rows[0].Position != "Site manager" || !rows[0].Active {
		t.Fatalf("got position %v active %v", rows[0].Position, rows[0].Active)
	}
	queries := recorder.Queries()
	if len(queries) != 1 {
		t.Fatalf("got %d queries, want one joined query", len(queries))
	}
	requireFragments(t, queries[0].SQL,
		`FROM "project_member"`,
		"JOIN users ON users.id = project_member.user_id",
		"JOIN role ON role.id = project_member.role_id",
		"WHERE project_member.project_id = $1",
		"ORDER BY users.name ASC, project_member.id ASC",
	)
	if queries[0].Args[0] != projectID.String() && queries[0].Args[0] != projectID {
		t.Fatalf("got args %v", queries[0].Args)
	}
}

func TestFindMemberRowReportsMissingMember(t *testing.T) {
	db, _ := databasetest.Open(t, databasetest.Result{Columns: memberRowColumns})

	_, err := NewRepository(db).FindMemberRow(context.Background(), uuid.New(), uuid.New())

	if !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestListMemberCandidatesSkipsInactiveUsersAndCurrentMembers(t *testing.T) {
	projectID := uuid.New()
	candidateID := uuid.New()
	db, recorder := databasetest.Open(t,
		databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(1)}}},
		databasetest.Result{
			Columns: []string{"id", "name", "email"},
			Rows:    [][]driver.Value{{candidateID.String(), "Budi Santoso", "budi@banguninaja.id"}},
		},
	)

	candidates, total, err := NewRepository(db).ListMemberCandidates(context.Background(), MemberCandidateFilter{
		ProjectID: projectID,
		Search:    "bud",
		Limit:     20,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(candidates) != 1 || candidates[0].ID != candidateID || candidates[0].Email != "budi@banguninaja.id" {
		t.Fatalf("got %d %+v", total, candidates)
	}
	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries, want count and page", len(queries))
	}
	for _, query := range queries {
		requireFragments(t, query.SQL,
			`FROM "users"`,
			"users.active = $1",
			"NOT EXISTS (SELECT 1 FROM project_member WHERE project_member.user_id = users.id AND project_member.project_id = $2)",
			"(users.name ILIKE $3 OR users.email ILIKE $4)",
		)
		if query.Args[0] != true || query.Args[2] != "%bud%" {
			t.Errorf("got args %v", query.Args)
		}
	}
	requireFragments(t, queries[1].SQL, "SELECT users.id, users.name, users.email", "ORDER BY users.name ASC, users.id ASC", "LIMIT $5")
	if strings.Contains(queries[1].SQL, "password") {
		t.Fatalf("candidate query must only read id, name, and email:\n%s", queries[1].SQL)
	}
}

func TestListProjectsMineFiltersOnMembership(t *testing.T) {
	viewerID := uuid.New()
	db, recorder := databasetest.Open(t,
		databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(0)}}},
		databasetest.Result{Columns: []string{"id"}},
	)

	if _, _, err := NewRepository(db).ListProjects(context.Background(), ProjectFilter{MemberID: &viewerID, Limit: 20}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries", len(queries))
	}
	for _, query := range queries {
		requireFragments(t, query.SQL, "id IN (SELECT project_id FROM project_member WHERE user_id = $1)")
		if query.Args[0] != viewerID.String() && query.Args[0] != viewerID {
			t.Errorf("got args %v, want viewer %s", query.Args, viewerID)
		}
	}
}

func TestListProjectsWithoutMineIgnoresMembership(t *testing.T) {
	db, recorder := databasetest.Open(t,
		databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(0)}}},
		databasetest.Result{Columns: []string{"id"}},
	)

	if _, _, err := NewRepository(db).ListProjects(context.Background(), ProjectFilter{Limit: 20}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, query := range recorder.Queries() {
		if strings.Contains(query.SQL, "project_member") {
			t.Fatalf("plain project list must not touch project_member:\n%s", query.SQL)
		}
	}
}

func TestUpdateMemberWritesOnlyRoleAndPosition(t *testing.T) {
	db, recorder := databasetest.Open(t, databasetest.Result{Affected: 1})
	member := ProjectMember{ProjectID: uuid.New(), RoleID: uuid.New()}
	member.ID = uuid.New()

	if err := NewRepository(db).UpdateMember(context.Background(), &member); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	queries := recorder.Queries()
	if len(queries) != 1 {
		t.Fatalf("got %d queries", len(queries))
	}
	requireFragments(t, queries[0].SQL,
		`UPDATE "project_member" SET "position"=$1,"role_id"=$2,"updated_at"=$3`,
		"WHERE id = $4 AND project_id = $5",
	)
	if strings.Contains(queries[0].SQL, "user_id") || strings.Contains(queries[0].SQL, "created_at") {
		t.Fatalf("update must not touch the user or join date:\n%s", queries[0].SQL)
	}
	if member.UpdatedAt.IsZero() {
		t.Fatal("updated member must carry its new updated_at")
	}
}

func TestUpdateMemberReportsMissingMember(t *testing.T) {
	db, _ := databasetest.Open(t, databasetest.Result{Affected: 0})
	member := ProjectMember{ProjectID: uuid.New(), RoleID: uuid.New()}
	member.ID = uuid.New()

	if err := NewRepository(db).UpdateMember(context.Background(), &member); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestFindMemberUserReadsActiveFlagAndAccountRole(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()
	db, recorder := databasetest.Open(t,
		databasetest.Result{Columns: []string{"id", "active", "role_id"}, Rows: [][]driver.Value{{userID.String(), false, roleID.String()}}},
		databasetest.Result{Columns: []string{"id", "active", "role_id"}},
	)
	repository := NewRepository(db)

	user, err := repository.FindMemberUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != userID || user.Active || user.RoleID == nil || *user.RoleID != roleID {
		t.Fatalf("got %+v", user)
	}
	requireFragments(t, recorder.Queries()[0].SQL, "SELECT users.id, users.active, users.role_id", `FROM "users"`, "WHERE users.id = $1")

	if _, err := repository.FindMemberUser(context.Background(), uuid.New()); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound for a missing user", err)
	}
}

func TestFindRoleIDLooksUpByCode(t *testing.T) {
	roleID := uuid.New()
	db, recorder := databasetest.Open(t,
		databasetest.Result{Columns: []string{"id"}, Rows: [][]driver.Value{{roleID.String()}}},
		databasetest.Result{Columns: []string{"id"}},
	)
	repository := NewRepository(db)

	found, err := repository.FindRoleID(context.Background(), "keuangan")
	if err != nil || found != roleID {
		t.Fatalf("got %s, %v", found, err)
	}
	requireFragments(t, recorder.Queries()[0].SQL, `FROM "role"`, "role.code = $1")
	if recorder.Queries()[0].Args[0] != "keuangan" {
		t.Fatalf("got args %v", recorder.Queries()[0].Args)
	}

	if _, err := repository.FindRoleID(context.Background(), "mandor"); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound for an unknown role", err)
	}
}
