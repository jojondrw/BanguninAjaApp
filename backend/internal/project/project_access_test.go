package project

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
)

type accessRepository struct {
	Repository
	isMember     bool
	memberChecks int
	listed       ProjectFilter
	summaryScope *uuid.UUID
	roleID       uuid.UUID
	joined       *ProjectMember
}

func (f *accessRepository) Transaction(_ context.Context, work func(Repository) error) error {
	return work(f)
}

func (f *accessRepository) IsProjectMember(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	f.memberChecks++
	return f.isMember, nil
}

func (f *accessRepository) ListProjects(_ context.Context, filter ProjectFilter) ([]Project, int64, error) {
	f.listed = filter
	return nil, 0, nil
}

func (f *accessRepository) SummarizeProjects(_ context.Context, memberID *uuid.UUID) (ProjectTotals, error) {
	f.summaryScope = memberID
	return ProjectTotals{}, nil
}

func (f *accessRepository) CreateProject(_ context.Context, project *Project) error {
	project.ID = uuid.New()
	return nil
}

func (f *accessRepository) FindRoleID(context.Context, string) (uuid.UUID, error) {
	return f.roleID, nil
}

func (f *accessRepository) CreateMember(_ context.Context, member *ProjectMember) error {
	f.joined = member
	return nil
}

func TestAdminObserverAndFinanceOpenEveryProjectWithoutAMembershipCheck(t *testing.T) {
	for _, role := range []string{access.RoleAdmin, access.RoleObserver, access.RoleFinance} {
		repository := &accessRepository{}
		if err := NewService(repository).EnsureProjectAccess(context.Background(), uuid.New(), uuid.New(), role); err != nil {
			t.Fatalf("%s: unexpected error %v", role, err)
		}
		if repository.memberChecks != 0 {
			t.Fatalf("%s: membership was checked", role)
		}
	}
}

func TestOtherRolesOpenOnlyProjectsWhoseTeamTheyAreIn(t *testing.T) {
	for _, role := range []string{access.RoleProjectManager, access.RoleProcurement, access.RoleSales, access.RoleHR} {
		member := &accessRepository{isMember: true}
		if err := NewService(member).EnsureProjectAccess(context.Background(), uuid.New(), uuid.New(), role); err != nil {
			t.Fatalf("%s member: unexpected error %v", role, err)
		}
		outsider := &accessRepository{}
		if err := NewService(outsider).EnsureProjectAccess(context.Background(), uuid.New(), uuid.New(), role); !errors.Is(err, errNotProjectMember) {
			t.Fatalf("%s outsider: err = %v, want errNotProjectMember", role, err)
		}
	}
}

func TestUnknownRoleIsRefusedInsteadOfLetThrough(t *testing.T) {
	repository := &accessRepository{isMember: true}
	if err := NewService(repository).EnsureProjectAccess(context.Background(), uuid.New(), uuid.New(), ""); !errors.Is(err, errProjectViewerUnknown) {
		t.Fatalf("err = %v, want errProjectViewerUnknown", err)
	}
}

func TestListAndSummaryAreScopedToTheTeamForRestrictedRoles(t *testing.T) {
	viewerID := uuid.New()
	restricted := &accessRepository{}
	service := NewService(restricted)

	if _, err := service.ListProjects(context.Background(), viewerID, access.RoleProjectManager, ProjectQuery{}); err != nil {
		t.Fatalf("list: %v", err)
	}
	if restricted.listed.MemberID == nil || *restricted.listed.MemberID != viewerID {
		t.Fatal("project manager list is not limited to their projects")
	}
	if _, err := service.SummarizeProjects(context.Background(), viewerID, access.RoleProjectManager); err != nil {
		t.Fatalf("summary: %v", err)
	}
	if restricted.summaryScope == nil || *restricted.summaryScope != viewerID {
		t.Fatal("project manager summary counts projects outside their teams")
	}

	finance := &accessRepository{}
	if _, err := NewService(finance).SummarizeProjects(context.Background(), viewerID, access.RoleFinance); err != nil {
		t.Fatalf("finance summary: %v", err)
	}
	if finance.summaryScope != nil {
		t.Fatal("finance summary was limited to a team")
	}
}

func TestCreatorJoinsTheTeamWithTheirRole(t *testing.T) {
	repository := &accessRepository{roleID: uuid.New()}
	creator := uuid.New()

	created, err := NewService(repository).CreateProject(context.Background(), creator, access.RoleProjectManager, ProjectRequest{Code: "UJI-1", Name: "Uji", Type: "Perumahan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	joined := repository.joined
	if joined == nil || joined.UserID != creator || joined.ProjectID != created.ID || joined.RoleID != repository.roleID {
		t.Fatalf("creator membership = %+v, want the creator on the new project with their role", joined)
	}
}

func TestProjectRoutesStopOutsidersBeforeReadingTheProject(t *testing.T) {
	projectID := uuid.New()

	response, recorder := serveProjects(t, projectCall{method: http.MethodGet, path: membersPath(projectID, ""), role: access.RoleProjectManager}, membership(false))
	requireStatus(t, response, http.StatusForbidden, "project_member_required")
	if len(recorder.Queries()) != 1 {
		t.Fatalf("ran %d queries, want only the membership check", len(recorder.Queries()))
	}

	response, recorder = serveProjects(t, projectCall{method: http.MethodGet, path: membersPath(projectID, "/payment-schedule"), role: access.RoleFinance}, projectMissing())
	requireStatus(t, response, http.StatusNotFound, "project_not_found")
	if len(recorder.Queries()) != 1 {
		t.Fatalf("finance ran %d queries, want no membership check", len(recorder.Queries()))
	}
}
