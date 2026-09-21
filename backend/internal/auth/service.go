package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/security"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
)

var (
	errEmailAlreadyUsed  = apperror.Conflict("email_already_used", "Email sudah terdaftar")
	errInvalidCredential = apperror.Unauthorized("invalid_credential", "Email atau kata sandi salah")
	errInvalidRefresh    = apperror.Unauthorized("invalid_refresh_token", "Sesi sudah berakhir, silakan masuk lagi")
	errUserUnknown       = apperror.Unauthorized("user_not_found", "Akun tidak ditemukan")
)

type Service interface {
	Register(ctx context.Context, request RegisterRequest) (UserResponse, error)
	Login(ctx context.Context, request LoginRequest) (Session, error)
	Refresh(ctx context.Context, refreshToken string) (Session, error)
	Logout(ctx context.Context, refreshToken string) error
	Profile(ctx context.Context, userID uuid.UUID) (UserResponse, error)
}

type service struct {
	repository Repository
	tokens     *token.Manager
}

func NewService(repository Repository, tokens *token.Manager) Service {
	return &service{repository: repository, tokens: tokens}
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
	}

	if err := s.repository.CreateUser(ctx, &user); err != nil {
		return UserResponse{}, apperror.Internal(err)
	}

	return newUserResponse(user), nil
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

	user, err := s.repository.FindUserByID(ctx, stored.UserID)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return Session{}, errUserUnknown
		}
		return Session{}, apperror.Internal(err)
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
	user, err := s.repository.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return UserResponse{}, errUserUnknown
		}
		return UserResponse{}, apperror.Internal(err)
	}

	return newUserResponse(user), nil
}

func (s *service) issueSession(ctx context.Context, user User) (Session, error) {
	access, err := s.tokens.IssueAccess(user.ID)
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
			AccessToken:          access.Value,
			AccessTokenExpiresAt: access.ExpiresAt,
			User:                 newUserResponse(user),
		},
		RefreshToken:     refresh.Value,
		RefreshExpiresAt: refresh.ExpiresAt,
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
