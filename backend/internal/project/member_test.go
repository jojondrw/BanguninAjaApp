package project

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

type memberRepository struct {
	Repository
	projectErr error
	user       MemberUser
	userErr    error
	roles      map[string]uuid.UUID
	member     ProjectMember
	memberErr  error
	createErr  error
	created    *ProjectMember
	updated    *ProjectMember
	filter     *ProjectFilter
}

func (f *memberRepository) FindProject(context.Context, uuid.UUID) (Project, error) {
	return Project{Status: statusOngoing}, f.projectErr
}

func (f *memberRepository) FindMemberUser(context.Context, uuid.UUID) (MemberUser, error) {
	return f.user, f.userErr
}

func (f *memberRepository) FindRoleID(_ context.Context, code string) (uuid.UUID, error) {
	roleID, ok := f.roles[code]
	if !ok {
		return uuid.Nil, database.ErrNotFound
	}
	return roleID, nil
}

func (f *memberRepository) CreateMember(_ context.Context, member *ProjectMember) error {
	if f.createErr != nil {
		return f.createErr
	}
	member.ID = uuid.New()
	stored := *member
	f.created = &stored
	return nil
}

func (f *memberRepository) FindMember(context.Context, uuid.UUID, uuid.UUID) (ProjectMember, error) {
	return f.member, f.memberErr
}

func (f *memberRepository) UpdateMember(_ context.Context, member *ProjectMember) error {
	stored := *member
	f.updated = &stored
	return nil
}

func (f *memberRepository) FindMemberRow(_ context.Context, projectID, id uuid.UUID) (MemberRow, error) {
	return MemberRow{ID: id, ProjectID: projectID, Name: "Rina", RoleCode: "manajer_proyek", RoleName: "Manajer Proyek"}, nil
}

func (f *memberRepository) ListProjects(_ context.Context, filter ProjectFilter) ([]Project, int64, error) {
	f.filter = &filter
	return nil, 0, nil
}

func textOf(value *string) string {
	if value == nil {
		return "<nil>"
	}
	return *value
}

func TestAddMemberFollowsTheAccountRoleWhenNoneIsChosen(t *testing.T) {
	accountRole := uuid.New()
	repository := &memberRepository{user: MemberUser{Active: true, RoleID: &accountRole}}
	userID := uuid.New()

	response, err := NewService(repository).AddMember(context.Background(), uuid.New(), MemberCreateRequest{UserID: userID, Position: "  Site manager  "})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	created := repository.created
	if created == nil || created.RoleID != accountRole || created.UserID != userID {
		t.Fatalf("got created member %+v, want account role %s", created, accountRole)
	}
	if textOf(created.Position) != "Site manager" {
		t.Fatalf("got position %q, want trimmed text", textOf(created.Position))
	}
	if response.ID != created.ID || response.Role.Code != "manajer_proyek" {
		t.Fatalf("got response %+v", response)
	}
}

func TestAddMemberUsesTheChosenProjectRole(t *testing.T) {
	accountRole := uuid.New()
	chosenRole := uuid.New()
	repository := &memberRepository{
		user:  MemberUser{Active: true, RoleID: &accountRole},
		roles: map[string]uuid.UUID{"pengamat": chosenRole},
	}

	_, err := NewService(repository).AddMember(context.Background(), uuid.New(), MemberCreateRequest{UserID: uuid.New(), Position: "   ", Role: "pengamat"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.created.RoleID != chosenRole {
		t.Fatalf("got role %s, want chosen role %s", repository.created.RoleID, chosenRole)
	}
	if repository.created.Position != nil {
		t.Fatalf("blank position must be stored as NULL, got %q", *repository.created.Position)
	}
}

func TestAddMemberRejectsInvalidRequests(t *testing.T) {
	accountRole := uuid.New()
	cases := []struct {
		name       string
		repository *memberRepository
		request    MemberCreateRequest
		want       error
	}{
		{"missing project", &memberRepository{projectErr: database.ErrNotFound}, MemberCreateRequest{}, errProjectNotFound},
		{"unknown user", &memberRepository{userErr: database.ErrNotFound}, MemberCreateRequest{}, errMemberUserNotFound},
		{"inactive user", &memberRepository{user: MemberUser{Active: false, RoleID: &accountRole}}, MemberCreateRequest{}, errMemberUserInactive},
		{"user without role", &memberRepository{user: MemberUser{Active: true}}, MemberCreateRequest{}, errMemberRoleRequired},
		{"unknown role", &memberRepository{user: MemberUser{Active: true, RoleID: &accountRole}}, MemberCreateRequest{Role: "mandor"}, errMemberRoleNotFound},
		{
			"already a member",
			&memberRepository{user: MemberUser{Active: true, RoleID: &accountRole}, createErr: fmt.Errorf("%w: unique", database.ErrDuplicate)},
			MemberCreateRequest{},
			errMemberExists,
		},
		{
			"user deleted meanwhile",
			&memberRepository{user: MemberUser{Active: true, RoleID: &accountRole}, createErr: fmt.Errorf("%w: fk", database.ErrReferenceViolated)},
			MemberCreateRequest{},
			errMemberReferenceGone,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.request.UserID = uuid.New()
			_, err := NewService(tc.repository).AddMember(context.Background(), uuid.New(), tc.request)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
			if tc.repository.created != nil {
				t.Fatal("rejected member must not be stored")
			}
		})
	}
}

func TestUpdateMemberKeepsFieldsThatWereNotSent(t *testing.T) {
	roleID := uuid.New()
	position := "Estimator"
	repository := &memberRepository{member: ProjectMember{RoleID: roleID, Position: &position}}

	_, err := NewService(repository).UpdateMember(context.Background(), uuid.New(), uuid.New(), MemberUpdateRequest{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.updated.RoleID != roleID || textOf(repository.updated.Position) != position {
		t.Fatalf("got %+v, want role and position unchanged", repository.updated)
	}
}

func TestUpdateMemberChangesPositionAndRole(t *testing.T) {
	newRole := uuid.New()
	position := "Estimator"
	repository := &memberRepository{
		member: ProjectMember{RoleID: uuid.New(), Position: &position},
		roles:  map[string]uuid.UUID{"keuangan": newRole},
	}
	empty := ""
	role := " keuangan "

	_, err := NewService(repository).UpdateMember(context.Background(), uuid.New(), uuid.New(), MemberUpdateRequest{Position: &empty, Role: &role})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.updated.RoleID != newRole {
		t.Fatalf("got role %s, want %s", repository.updated.RoleID, newRole)
	}
	if repository.updated.Position != nil {
		t.Fatalf("an empty position must clear it, got %q", *repository.updated.Position)
	}
}

func TestUpdateMemberRejectsUnknownMemberAndRole(t *testing.T) {
	missing := &memberRepository{memberErr: database.ErrNotFound}
	if _, err := NewService(missing).UpdateMember(context.Background(), uuid.New(), uuid.New(), MemberUpdateRequest{}); !errors.Is(err, errMemberNotFound) {
		t.Fatalf("got %v, want errMemberNotFound", err)
	}

	role := "mandor"
	unknownRole := &memberRepository{member: ProjectMember{RoleID: uuid.New()}}
	if _, err := NewService(unknownRole).UpdateMember(context.Background(), uuid.New(), uuid.New(), MemberUpdateRequest{Role: &role}); !errors.Is(err, errMemberRoleNotFound) {
		t.Fatalf("got %v, want errMemberRoleNotFound", err)
	}
	if unknownRole.updated != nil {
		t.Fatal("member must not be saved with an unknown role")
	}
}

func TestListProjectsFiltersByMembershipOnlyWhenAsked(t *testing.T) {
	viewerID := uuid.New()
	repository := &memberRepository{}
	service := NewService(repository)

	if _, err := service.ListProjects(context.Background(), viewerID, ProjectQuery{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.filter.MemberID != nil {
		t.Fatalf("plain list must not filter by membership, got %v", *repository.filter.MemberID)
	}

	if _, err := service.ListProjects(context.Background(), viewerID, ProjectQuery{Mine: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repository.filter.MemberID == nil || *repository.filter.MemberID != viewerID {
		t.Fatalf("mine=true must filter by the caller, got %v", repository.filter.MemberID)
	}
}
