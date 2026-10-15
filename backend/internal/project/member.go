package project

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
)

var (
	errMemberNotFound      = apperror.NotFound("project_member_not_found", "Anggota tim tidak ditemukan di proyek ini")
	errMemberExists        = apperror.Conflict("project_member_exists", "Pengguna ini sudah menjadi anggota tim proyek")
	errMemberUserNotFound  = apperror.Unprocessable("member_user_not_found", "Pengguna yang dipilih tidak ditemukan")
	errMemberUserInactive  = apperror.Unprocessable("member_user_inactive", "Akun pengguna ini sudah dinonaktifkan sehingga tidak bisa ditambahkan ke tim")
	errMemberRoleNotFound  = apperror.Unprocessable("member_role_not_found", "Peran proyek yang dipilih tidak dikenal")
	errMemberRoleRequired  = apperror.Unprocessable("member_role_required", "Pengguna ini belum punya peran akun, pilih peran proyeknya")
	errMemberReferenceGone = apperror.Unprocessable("member_reference_not_found", "Proyek, pengguna, atau peran yang dipilih sudah tidak ada. Muat ulang lalu coba lagi")
)

var (
	memberReadErrors  = database.ErrorMap{NotFound: errMemberNotFound}
	memberWriteErrors = database.ErrorMap{NotFound: errMemberNotFound, Duplicate: errMemberExists, Referenced: errMemberReferenceGone}
	memberUserErrors  = database.ErrorMap{NotFound: errMemberUserNotFound}
	memberRoleErrors  = database.ErrorMap{NotFound: errMemberRoleNotFound}
)

func (s *service) ListMembers(ctx context.Context, projectID uuid.UUID) ([]MemberResponse, error) {
	if _, err := s.repository.FindProject(ctx, projectID); err != nil {
		return nil, projectReadErrors.Resolve(err)
	}

	rows, err := s.repository.ListMembers(ctx, projectID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return pagination.Map(rows, newMemberResponse), nil
}

func (s *service) ListMemberCandidates(ctx context.Context, projectID uuid.UUID, query MemberCandidateQuery) (pagination.Page[MemberCandidateResponse], error) {
	if _, err := s.repository.FindProject(ctx, projectID); err != nil {
		return pagination.Page[MemberCandidateResponse]{}, projectReadErrors.Resolve(err)
	}

	candidates, total, err := s.repository.ListMemberCandidates(ctx, MemberCandidateFilter{
		ProjectID: projectID,
		Search:    strings.TrimSpace(query.Search),
		Offset:    query.Offset(),
		Limit:     query.Size(),
	})
	if err != nil {
		return pagination.Page[MemberCandidateResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(candidates, newMemberCandidateResponse), query.Query, total), nil
}

func (s *service) AddMember(ctx context.Context, projectID uuid.UUID, request MemberCreateRequest) (MemberResponse, error) {
	if _, err := s.repository.FindProject(ctx, projectID); err != nil {
		return MemberResponse{}, projectReadErrors.Resolve(err)
	}

	roleID, err := s.memberRole(ctx, request.UserID, strings.TrimSpace(request.Role))
	if err != nil {
		return MemberResponse{}, err
	}

	member := ProjectMember{
		ProjectID: projectID,
		UserID:    request.UserID,
		RoleID:    roleID,
		Position:  optionalText(request.Position),
	}
	if err := s.repository.CreateMember(ctx, &member); err != nil {
		return MemberResponse{}, memberWriteErrors.Resolve(err)
	}
	return s.memberResponse(ctx, projectID, member.ID)
}

func (s *service) UpdateMember(ctx context.Context, projectID, memberID uuid.UUID, request MemberUpdateRequest) (MemberResponse, error) {
	member, err := s.repository.FindMember(ctx, projectID, memberID)
	if err != nil {
		return MemberResponse{}, memberReadErrors.Resolve(err)
	}

	if err := s.applyMemberUpdate(ctx, &member, request); err != nil {
		return MemberResponse{}, err
	}
	if err := s.repository.UpdateMember(ctx, &member); err != nil {
		return MemberResponse{}, memberWriteErrors.Resolve(err)
	}
	return s.memberResponse(ctx, projectID, member.ID)
}

func (s *service) RemoveMember(ctx context.Context, projectID, memberID uuid.UUID) error {
	return memberReadErrors.Resolve(s.repository.DeleteMember(ctx, projectID, memberID))
}

func (s *service) memberRole(ctx context.Context, userID uuid.UUID, roleCode string) (uuid.UUID, error) {
	user, err := s.repository.FindMemberUser(ctx, userID)
	if err != nil {
		return uuid.Nil, memberUserErrors.Resolve(err)
	}
	if !user.Active {
		return uuid.Nil, errMemberUserInactive
	}
	if roleCode != "" {
		return s.findRole(ctx, roleCode)
	}
	if user.RoleID == nil {
		return uuid.Nil, errMemberRoleRequired
	}
	return *user.RoleID, nil
}

func (s *service) applyMemberUpdate(ctx context.Context, member *ProjectMember, request MemberUpdateRequest) error {
	if request.Position != nil {
		member.Position = optionalText(*request.Position)
	}
	roleCode := ""
	if request.Role != nil {
		roleCode = strings.TrimSpace(*request.Role)
	}
	if roleCode == "" {
		return nil
	}

	roleID, err := s.findRole(ctx, roleCode)
	if err != nil {
		return err
	}
	member.RoleID = roleID
	return nil
}

func (s *service) findRole(ctx context.Context, code string) (uuid.UUID, error) {
	roleID, err := s.repository.FindRoleID(ctx, code)
	if err != nil {
		return uuid.Nil, memberRoleErrors.Resolve(err)
	}
	return roleID, nil
}

func (s *service) memberResponse(ctx context.Context, projectID, memberID uuid.UUID) (MemberResponse, error) {
	row, err := s.repository.FindMemberRow(ctx, projectID, memberID)
	if err != nil {
		return MemberResponse{}, memberReadErrors.Resolve(err)
	}
	return newMemberResponse(row), nil
}

func optionalText(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
