package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
)

func TestRoleChangeIsRecordedWithTheActorAndBothValues(t *testing.T) {
	repository := newMemoryRepository()
	repository.addUser(access.RoleAdmin, true)
	staffID := repository.addUser(access.RoleSales, true)
	actor := uuid.New()
	service, _ := newTestService(repository)

	if _, err := service.ChangeRole(context.Background(), actor, staffID, ChangeRoleRequest{Role: access.RoleProcurement}); err != nil {
		t.Fatal(err)
	}

	if len(repository.audit) != 1 {
		t.Fatalf("recorded %d entries, want 1", len(repository.audit))
	}
	entry := repository.audit[0]
	if entry.Action != auditRole || entry.OldValue != access.RoleSales || entry.NewValue != access.RoleProcurement {
		t.Fatalf("entry = %+v", entry)
	}
	if entry.ActorID == nil || *entry.ActorID != actor || entry.UserID == nil || *entry.UserID != staffID {
		t.Fatalf("actor or target missing: %+v", entry)
	}
}

func TestDeactivationIsRecordedAndUnchangedValuesAreNot(t *testing.T) {
	repository := newMemoryRepository()
	repository.addUser(access.RoleAdmin, true)
	staffID := repository.addUser(access.RoleHR, true)
	service, _ := newTestService(repository)

	if _, err := service.ChangeActive(context.Background(), uuid.New(), staffID, ChangeActiveRequest{Active: boolPointer(false)}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ChangeRole(context.Background(), uuid.New(), staffID, ChangeRoleRequest{Role: access.RoleHR}); err != nil {
		t.Fatal(err)
	}

	if len(repository.audit) != 1 {
		t.Fatalf("recorded %d entries, want only the deactivation", len(repository.audit))
	}
	if entry := repository.audit[0]; entry.Action != auditActive || entry.OldValue != "true" || entry.NewValue != "false" {
		t.Fatalf("entry = %+v", entry)
	}
}

func TestSystemChangesHaveNoActor(t *testing.T) {
	entries := auditEntries(uuid.Nil, UserRow{ID: uuid.New()}, UserRow{Active: true})

	if len(entries) != 1 || entries[0].ActorID != nil {
		t.Fatalf("entries = %+v, want one entry without an actor", entries)
	}
}

func TestRefusedChangesLeaveNoAuditTrail(t *testing.T) {
	repository := newMemoryRepository()
	adminID := repository.addUser(access.RoleAdmin, true)
	service, _ := newTestService(repository)

	if _, err := service.ChangeActive(context.Background(), uuid.New(), adminID, ChangeActiveRequest{Active: boolPointer(false)}); err == nil {
		t.Fatal("the last admin was deactivated")
	}
	if len(repository.audit) != 0 {
		t.Fatalf("a refused change was recorded: %+v", repository.audit)
	}
}
