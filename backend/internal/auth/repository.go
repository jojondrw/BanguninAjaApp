package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var errUserNotFound = errors.New("user not found")
var errRefreshTokenNotFound = errors.New("refresh token not found")

type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	EmailExists(ctx context.Context, email string) (bool, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (User, error)
	StoreRefreshToken(ctx context.Context, refreshToken *RefreshToken) error
	FindRefreshTokenByHash(ctx context.Context, hash string) (RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) CreateUser(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *gormRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&User{}).Where("email = ?", email).Count(&total).Error
	return total > 0, err
}

func (r *gormRepository) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", email).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, errUserNotFound
	}
	return user, err
}

func (r *gormRepository) FindUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, errUserNotFound
	}
	return user, err
}

func (r *gormRepository) StoreRefreshToken(ctx context.Context, refreshToken *RefreshToken) error {
	return r.db.WithContext(ctx).Create(refreshToken).Error
}

func (r *gormRepository) FindRefreshTokenByHash(ctx context.Context, hash string) (RefreshToken, error) {
	var refreshToken RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ?", hash).Take(&refreshToken).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return RefreshToken{}, errRefreshTokenNotFound
	}
	return refreshToken, err
}

func (r *gormRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", revokedAt).Error
}

func (r *gormRepository) RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", revokedAt).Error
}
