package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/pagination"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/security"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

var (
	errEmailAlreadyUsed  = apperror.Conflict("email_already_used", "Email sudah terdaftar")
	errInvalidCredential = apperror.Unauthorized("invalid_credential", "Email atau kata sandi salah")
	errInvalidRefresh    = apperror.Unauthorized("invalid_refresh_token", "Sesi sudah berakhir, silakan masuk lagi")
	errUserUnknown       = apperror.Unauthorized("user_not_found", "Akun tidak ditemukan")
	errAccountInactive   = apperror.Unauthorized("account_inactive", "Akun ini sudah dinonaktifkan, hubungi admin")
	errLoginInactive     = apperror.Forbidden("account_inactive", "Akun ini sudah dinonaktifkan, hubungi admin")
	errManagedNotFound   = apperror.NotFound("user_not_found", "Pengguna tidak ditemukan")
	errUnknownRole       = apperror.Unprocessable("unknown_role", "Peran tidak dikenal")
	errLastAdmin         = apperror.Unprocessable("last_admin", "Harus tetap ada satu admin aktif. Jadikan pengguna lain admin dulu")
)

type Service interface {
	Register(ctx context.Context, request RegisterRequest) (UserResponse, error)
	Login(ctx context.Context, request LoginRequest) (Session, error)
	Refresh(ctx context.Context, refreshToken string) (Session, error)
	Logout(ctx context.Context, refreshToken string) error
	Profile(ctx context.Context, userID uuid.UUID) (UserResponse, error)
	ListUsers(ctx context.Context, query UserQuery) (pagination.Page[ManagedUserResponse], error)
	ChangeRole(ctx context.Context, userID uuid.UUID, request ChangeRoleRequest) (ManagedUserResponse, error)
	ChangeActive(ctx context.Context, userID uuid.UUID, request ChangeActiveRequest) (ManagedUserResponse, error)
}

type subjectCache interface {
	Forget(userID uuid.UUID)
}

type service struct {
	repository Repository
	tokens     *token.Manager
	subjects   subjectCache
}

func NewService(repository Repository, tokens *token.Manager, subjects subjectCache) Service {
	return &service{repository: repository, tokens: tokens, subjects: subjects}
}

func (s *service) Register(ctx context.Context, request RegisterRequest) (UserResponse, error) {
	email := normalizeEmail(request.Email)

	taken, err := s.repository.EmailExists(ctx, email)
	if err != nil {
		return UserResponse{}, apperror.Internal(err)
	}
	if taken {
		return UserResponse{}, errEmailAlreadyUsed
	}

	passwordHash, err := security.HashPassword(request.Password)
	if err != nil {
		return UserResponse{}, apperror.Internal(err)
	}

	user := User{
		Name:         strings.TrimSpace(request.Name),
		Email:        email,
		PasswordHash: passwordHash,
		Active:       true,
	}

	if err := s.repository.CreateUserWithRole(ctx, &user, roleForNewcomer); err != nil {
		return UserResponse{}, apperror.Internal(fmt.Errorf("create user with role: %w", err))
	}

	return s.describe(ctx, user)
}

func (s *service) Login(ctx context.Context, request LoginRequest) (Session, error) {
	user, err := s.repository.FindUserByEmail(ctx, normalizeEmail(request.Email))
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return Session{}, errInvalidCredential
		}
		return Session{}, apperror.Internal(err)
	}

	if !security.PasswordMatches(user.PasswordHash, request.Password) {
		return Session{}, errInvalidCredential
	}
	if !user.Active {
		return Session{}, errLoginInactive
	}

	return s.issueSession(ctx, user)
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (Session, error) {
	if refreshToken == "" {
		return Session{}, errInvalidRefresh
	}

	stored, err := s.repository.FindRefreshTokenByHash(ctx, token.HashRefresh(refreshToken))
	if err != nil {
		if errors.Is(err, errRefreshTokenNotFound) {
			return Session{}, errInvalidRefresh
		}
		return Session{}, apperror.Internal(err)
	}

	now := time.Now()
	if !stored.IsUsable(now) {
		if stored.RevokedAt != nil {
			if err := s.repository.RevokeUserRefreshTokens(ctx, stored.UserID, now); err != nil {
				return Session{}, apperror.Internal(err)
			}
		}
		return Session{}, errInvalidRefresh
	}

	if err := s.repository.RevokeRefreshToken(ctx, stored.ID, now); err != nil {
		return Session{}, apperror.Internal(err)
	}

	user, err := s.activeUser(ctx, stored.UserID)
	if err != nil {
		return Session{}, err
	}

	return s.issueSession(ctx, user)
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	stored, err := s.repository.FindRefreshTokenByHash(ctx, token.HashRefresh(refreshToken))
	if err != nil {
		if errors.Is(err, errRefreshTokenNotFound) {
			return nil
		}
		return apperror.Internal(err)
	}

	if err := s.repository.RevokeRefreshToken(ctx, stored.ID, time.Now()); err != nil {
		return apperror.Internal(err)
	}

	return nil
}

func (s *service) Profile(ctx context.Context, userID uuid.UUID) (UserResponse, error) {
	user, err := s.activeUser(ctx, userID)
	if err != nil {
		return UserResponse{}, err
	}

	return s.describe(ctx, user)
}

func (s *service) activeUser(ctx context.Context, userID uuid.UUID) (User, error) {
	user, err := s.repository.FindUserByID(ctx, userID)
	if errors.Is(err, errUserNotFound) {
		return User{}, errUserUnknown
	}
	if err != nil {
		return User{}, apperror.Internal(err)
	}
	if !user.Active {
		return User{}, errAccountInactive
	}
	return user, nil
}

func (s *service) describe(ctx context.Context, user User) (UserResponse, error) {
	if user.RoleID == nil {
		return newUserResponse(user, nil), nil
	}

	role, err := s.repository.FindRoleByID(ctx, *user.RoleID)
	if errors.Is(err, errRoleNotFound) {
		return newUserResponse(user, nil), nil
	}
	if err != nil {
		return UserResponse{}, apperror.Internal(err)
	}
	return newUserResponse(user, &role), nil
}

func (s *service) issueSession(ctx context.Context, user User) (Session, error) {
	profile, err := s.describe(ctx, user)
	if err != nil {
		return Session{}, err
	}

	accessToken, err := s.tokens.IssueAccess(user.ID)
	if err != nil {
		return Session{}, apperror.Internal(err)
	}

	refresh, err := s.tokens.IssueRefresh()
	if err != nil {
		return Session{}, apperror.Internal(err)
	}

	stored := RefreshToken{
		UserID:    user.ID,
		TokenHash: refresh.Hash,
		ExpiresAt: refresh.ExpiresAt,
	}

	if err := s.repository.StoreRefreshToken(ctx, &stored); err != nil {
		return Session{}, apperror.Internal(err)
	}

	return Session{
		Response: SessionResponse{
			AccessToken:          accessToken.Value,
			AccessTokenExpiresAt: accessToken.ExpiresAt,
			User:                 profile,
		},
		RefreshToken:     refresh.Value,
		RefreshExpiresAt: refresh.ExpiresAt,
	}, nil
}

func (s *service) ListUsers(ctx context.Context, query UserQuery) (pagination.Page[ManagedUserResponse], error) {
	rows, total, err := s.repository.ListUsers(ctx, UserFilter{
		Search: strings.TrimSpace(query.Search),
		Role:   query.Role,
		Active: query.Active,
		Offset: query.Offset(),
		Limit:  query.Size(),
	})
	if err != nil {
		return pagination.Page[ManagedUserResponse]{}, apperror.Internal(err)
	}
	return pagination.New(pagination.Map(rows, newManagedUserResponse), query.Query, total), nil
}

func (s *service) ChangeRole(ctx context.Context, userID uuid.UUID, request ChangeRoleRequest) (ManagedUserResponse, error) {
	if _, ok := access.FindRole(request.Role); !ok {
		return ManagedUserResponse{}, errUnknownRole
	}

	return s.changeUser(ctx, userID, func(repository Repository, target UserRow) error {
		if err := guardLastAdmin(ctx, repository, target, request.Role != access.RoleAdmin); err != nil {
			return err
		}
		role, err := repository.FindRoleByCode(ctx, request.Role)
		if err != nil {
			return err
		}
		return repository.UpdateUserRole(ctx, userID, role.ID)
	})
}

func (s *service) ChangeActive(ctx context.Context, userID uuid.UUID, request ChangeActiveRequest) (ManagedUserResponse, error) {
	active := *request.Active

	return s.changeUser(ctx, userID, func(repository Repository, target UserRow) error {
		if err := guardLastAdmin(ctx, repository, target, !active); err != nil {
			return err
		}
		if err := repository.UpdateUserActive(ctx, userID, active); err != nil {
			return err
		}
		if active {
			return nil
		}
		return repository.RevokeUserRefreshTokens(ctx, userID, time.Now())
	})
}

func (s *service) changeUser(ctx context.Context, userID uuid.UUID, change func(Repository, UserRow) error) (ManagedUserResponse, error) {
	var updated UserRow
	err := s.repository.Transaction(ctx, func(repository Repository) error {
		if err := repository.LockAdminChanges(ctx); err != nil {
			return err
		}
		target, err := repository.FindUserRow(ctx, userID)
		if err != nil {
			return err
		}
		if err := change(repository, target); err != nil {
			return err
		}
		updated, err = repository.FindUserRow(ctx, userID)
		return err
	})
	s.subjects.Forget(userID)
	if err != nil {
		return ManagedUserResponse{}, translateChangeError(err)
	}
	return newManagedUserResponse(updated), nil
}

func guardLastAdmin(ctx context.Context, repository Repository, target UserRow, losesAdmin bool) error {
	if !losesAdmin || !target.Active || valueOf(target.RoleCode) != access.RoleAdmin {
		return nil
	}
	admins, err := repository.CountActiveAdmins(ctx)
	if err != nil {
		return err
	}
	if admins <= 1 {
		return errLastAdmin
	}
	return nil
}

func translateChangeError(err error) error {
	var known *apperror.Error
	if errors.As(err, &known) {
		return known
	}
	if errors.Is(err, errUserNotFound) {
		return errManagedNotFound
	}
	if errors.Is(err, errRoleNotFound) {
		return errUnknownRole
	}
	return apperror.Internal(err)
}

func roleForNewcomer(existingUsers int64) string {
	if existingUsers == 0 {
		return access.RoleAdmin
	}
	return access.RoleObserver
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
