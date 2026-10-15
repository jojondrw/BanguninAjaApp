package main

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/auth"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/databasetest"
)

var userRowColumns = []string{"id", "name", "email", "active", "created_at", "role_code", "role_name"}

func TestDemoAccountRegisteredAfterOthersStillBecomesAdmin(t *testing.T) {
	t.Setenv("DEMO_USER_PASSWORD", "rahasia-uji-123")
	adminRoleID := uuid.New()
	now := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	pendingRow := func(roleCode, roleName any) databasetest.Result {
		return databasetest.Result{Columns: userRowColumns, Rows: [][]driver.Value{
			{uuid.NewString(), demoAccountName, demoAccountEmail, true, now, roleCode, roleName},
		}}
	}
	db, recorder := databasetest.Open(t,
		databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(0)}}},
		databasetest.Result{},
		databasetest.Result{Columns: []string{"count"}, Rows: [][]driver.Value{{int64(3)}}},
		databasetest.Result{Affected: 1},
		databasetest.Result{},
		pendingRow(nil, nil),
		databasetest.Result{Columns: []string{"id", "code", "name"}, Rows: [][]driver.Value{{adminRoleID.String(), access.RoleAdmin, "Admin"}}},
		databasetest.Result{Affected: 1},
		pendingRow(access.RoleAdmin, "Admin"),
	)
	accounts := auth.NewService(auth.NewRepository(db), nil, noSubjectCache{})

	owner, err := resolveOwner(context.Background(), newFakeStore(), accounts, "")

	if err != nil {
		t.Fatal(err)
	}
	if owner == uuid.Nil {
		t.Fatal("the demo account must own the data")
	}
	var promoted bool
	for _, query := range recorder.Queries() {
		if strings.Contains(query.SQL, `UPDATE "users" SET "role_id"`) && query.Args[0] == adminRoleID.String() {
			promoted = true
		}
	}
	if !promoted {
		t.Fatalf("the demo account was not made admin: %+v", recorder.Queries())
	}
}
