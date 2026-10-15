package auth

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

func assertContains(t *testing.T, sql string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("expected %q in %s", fragment, sql)
		}
	}
}

func TestAssignAdminToUnassignedUsersIsIdempotent(t *testing.T) {
	db, recorder := databasetest.Open(t, databasetest.Result{Affected: 3}, databasetest.Result{Affected: 0})
	repository := NewRepository(db)

	first, err := AssignAdminToUnassignedUsers(context.Background(), repository)
	if err != nil || first != 3 {
		t.Fatalf("first run got %d, %v", first, err)
	}
	second, err := AssignAdminToUnassignedUsers(context.Background(), repository)
	if err != nil || second != 0 {
		t.Fatalf("second run got %d, %v", second, err)
	}

	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries", len(queries))
	}
	for _, query := range queries {
		assertContains(t, query.SQL, "UPDATE users SET role_id = (SELECT id FROM role WHERE code = $1)", "WHERE role_id IS NULL")
		if query.Args[0] != access.RoleAdmin {
			t.Fatalf("got args %v", query.Args)
		}
	}
}

func TestEnsureRolesUpsertsEveryRoleByCode(t *testing.T) {
	db, recorder := databasetest.Open(t, databasetest.Result{Affected: 7})

	seeded, err := EnsureRoles(context.Background(), NewRepository(db))

	if err != nil || seeded != 7 {
		t.Fatalf("got %d, %v", seeded, err)
	}
	queries := recorder.Queries()
	if len(queries) != 1 {
		t.Fatalf("got %d queries", len(queries))
	}
	assertContains(t, queries[0].SQL, `INSERT INTO "role"`, `ON CONFLICT ("code") DO UPDATE SET`, `"name"="excluded"."name"`)
	for _, code := range access.RoleCodes() {
		found := false
		for _, arg := range queries[0].Args {
			found = found || arg == code
		}
		if !found {
			t.Fatalf("role %s missing from upsert", code)
		}
	}
}

func TestCreateUserWithRoleLocksCountsAndAssignsInOneTransaction(t *testing.T) {
	roleID := uuid.New()
	db, recorder := databasetest.Open(t,
		databasetest.Result{},
		databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(0)}}},
		databasetest.Result{Columns: []string{"id", "code", "name"}, Rows: [][]driver.Value{{roleID.String(), access.RoleAdmin, "Admin"}}},
		databasetest.Result{Affected: 1},
	)
	var counted int64 = -1
	user := User{Name: "UJI", Email: "uji@uji.test", PasswordHash: "x", Active: true}

	err := NewRepository(db).CreateUserWithRole(context.Background(), &user, func(existing int64) string {
		counted = existing
		return access.RoleAdmin
	})

	if err != nil {
		t.Fatal(err)
	}
	if counted != 0 || user.RoleID == nil || *user.RoleID != roleID {
		t.Fatalf("counted %d, role %v", counted, user.RoleID)
	}
	queries := recorder.Queries()
	if len(queries) != 6 || queries[0].SQL != databasetest.Begin || queries[5].SQL != databasetest.Commit {
		t.Fatalf("got %+v", queries)
	}
	assertContains(t, queries[1].SQL, "pg_advisory_xact_lock")
	assertContains(t, queries[2].SQL, `SELECT count(*) FROM "users"`)
	assertContains(t, queries[3].SQL, `FROM "role" WHERE code = $1`)
	assertContains(t, queries[4].SQL, `INSERT INTO "users"`)
}

func TestFindSubjectReadsRoleAndActiveInOneQuery(t *testing.T) {
	userID := uuid.New()
	db, recorder := databasetest.Open(t, databasetest.Result{
		Columns: []string{"active", "role_code"},
		Rows:    [][]driver.Value{{true, access.RoleHR}},
	})

	subject, err := NewRepository(db).FindSubject(context.Background(), userID)

	if err != nil || subject.Role != access.RoleHR || !subject.Active {
		t.Fatalf("got %+v, %v", subject, err)
	}
	queries := recorder.Queries()
	if len(queries) != 1 {
		t.Fatalf("got %d queries", len(queries))
	}
	assertContains(t, queries[0].SQL, "LEFT JOIN role ON role.id = users.role_id", "WHERE users.id = $1", "LIMIT $2")
}

func TestFindSubjectReportsUnknownUser(t *testing.T) {
	db, _ := databasetest.Open(t, databasetest.Result{Columns: []string{"active", "role_code"}})

	if _, err := NewRepository(db).FindSubject(context.Background(), uuid.New()); err != access.ErrUnknownSubject {
		t.Fatalf("got %v", err)
	}
}

func TestListUsersSearchesNameAndEmailWithRole(t *testing.T) {
	db, recorder := databasetest.Open(t, databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(0)}}})
	active := true

	if _, _, err := NewRepository(db).ListUsers(context.Background(), UserFilter{Search: "uji", Role: access.RoleSales, Active: &active, Limit: 20}); err != nil {
		t.Fatal(err)
	}
	queries := recorder.Queries()
	if len(queries) != 2 {
		t.Fatalf("got %d queries", len(queries))
	}
	assertContains(t, queries[1].SQL,
		"role.code AS role_code",
		"users.name ILIKE $1 OR users.email ILIKE $2",
		"role.code = $3",
		"users.active = $4",
		"ORDER BY users.name ASC, users.id ASC",
	)
}
