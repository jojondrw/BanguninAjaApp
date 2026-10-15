package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
)

const (
	registrationLockKey = 7301
	adminChangeLockKey  = 7302
)

var errUserNotFound = errors.New("user not found")
var errRefreshTokenNotFound = errors.New("refresh token not found")
var errRoleNotFound = errors.New("role not found")

type UserRow struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Active    bool
	RoleCode  *string
	RoleName  *string
	CreatedAt time.Time
}

type UserFilter struct {
	Search string
	Role   string
	Active *bool
	Offset int
	Limit  int
}

type subjectRow struct {
	Active   bool
	RoleCode *string
}

type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	EmailExists(ctx context.Context, email string) (bool, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (User, error)
	StoreRefreshToken(ctx context.Context, refreshToken *RefreshToken) error
	FindRefreshTokenByHash(ctx context.Context, hash string) (RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
	CreateUserWithRole(ctx context.Context, user *User, chooseRole func(existingUsers int64) string) error
	FindRoleByID(ctx context.Context, id uuid.UUID) (Role, error)
	FindRoleByCode(ctx context.Context, code string) (Role, error)
	FindSubject(ctx context.Context, userID uuid.UUID) (access.Subject, error)
	ListUsers(ctx context.Context, filter UserFilter) ([]UserRow, int64, error)
	FindUserRow(ctx context.Context, id uuid.UUID) (UserRow, error)
	UpdateUserRole(ctx context.Context, id, roleID uuid.UUID) error
	UpdateUserActive(ctx context.Context, id uuid.UUID, active bool) error
	CountActiveAdmins(ctx context.Context) (int64, error)
	LockAdminChanges(ctx context.Context) error
	UpsertRoles(ctx context.Context, roles []Role) error
	AssignRoleToLegacyUsers(ctx context.Context, roleCode string) (int64, error)
	Transaction(ctx context.Context, work func(Repository) error) error
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

func (r *gormRepository) Transaction(ctx context.Context, work func(Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return work(&gormRepository{db: tx})
	})
}

func (r *gormRepository) CreateUserWithRole(ctx context.Context, user *User, chooseRole func(existingUsers int64) string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transactional := &gormRepository{db: tx}
		if err := transactional.lock(ctx, registrationLockKey); err != nil {
			return err
		}

		var existing int64
		if err := transactional.db.WithContext(ctx).Model(&User{}).Count(&existing).Error; err != nil {
			return err
		}

		code := chooseRole(existing)
		if code == "" {
			return transactional.CreateUser(ctx, user)
		}

		role, err := transactional.FindRoleByCode(ctx, code)
		if err != nil {
			return err
		}

		user.RoleID = &role.ID
		return transactional.CreateUser(ctx, user)
	})
}

func (r *gormRepository) FindRoleByID(ctx context.Context, id uuid.UUID) (Role, error) {
	return r.findRole(ctx, "id = ?", id)
}

func (r *gormRepository) FindRoleByCode(ctx context.Context, code string) (Role, error) {
	return r.findRole(ctx, "code = ?", code)
}

func (r *gormRepository) findRole(ctx context.Context, condition string, value any) (Role, error) {
	var role Role
	err := r.db.WithContext(ctx).Where(condition, value).Take(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Role{}, errRoleNotFound
	}
	return role, err
}

func (r *gormRepository) FindSubject(ctx context.Context, userID uuid.UUID) (access.Subject, error) {
	var rows []subjectRow
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.active, role.code AS role_code").
		Joins("LEFT JOIN role ON role.id = users.role_id").
		Where("users.id = ?", userID).
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return access.Subject{}, err
	}
	if len(rows) == 0 {
		return access.Subject{}, access.ErrUnknownSubject
	}
	role := ""
	if rows[0].RoleCode != nil {
		role = *rows[0].RoleCode
	}
	return access.Subject{Role: role, Active: rows[0].Active}, nil
}

func (r *gormRepository) ListUsers(ctx context.Context, filter UserFilter) ([]UserRow, int64, error) {
	var total int64
	if err := r.userRows(ctx).Scopes(filter.apply).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]UserRow, 0, filter.Limit)
	err := r.userRows(ctx).
		Scopes(filter.apply).
		Order("users.name ASC, users.id ASC").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Scan(&rows).Error
	return rows, total, err
}

func (r *gormRepository) FindUserRow(ctx context.Context, id uuid.UUID) (UserRow, error) {
	var rows []UserRow
	if err := r.userRows(ctx).Where("users.id = ?", id).Limit(1).Scan(&rows).Error; err != nil {
		return UserRow{}, err
	}
	if len(rows) == 0 {
		return UserRow{}, errUserNotFound
	}
	return rows[0], nil
}

func (r *gormRepository) userRows(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("users").
		Select("users.id, users.name, users.email, users.active, users.created_at, role.code AS role_code, role.name AS role_name").
		Joins("LEFT JOIN role ON role.id = users.role_id")
}

func (f UserFilter) apply(db *gorm.DB) *gorm.DB {
	if f.Search != "" {
		pattern := database.ContainsPattern(f.Search)
		db = db.Where("users.name ILIKE ? OR users.email ILIKE ?", pattern, pattern)
	}
	db = f.applyRole(db)
	if f.Active != nil {
		db = db.Where("users.active = ?", *f.Active)
	}
	return db
}

func (f UserFilter) applyRole(db *gorm.DB) *gorm.DB {
	if f.Role == pendingRoleFilter {
		return db.Where("users.role_id IS NULL")
	}
	if f.Role != "" {
		return db.Where("role.code = ?", f.Role)
	}
	return db
}

func (r *gormRepository) UpdateUserRole(ctx context.Context, id, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		Updates(map[string]any{"role_id": roleID, "updated_at": time.Now()}).Error
}

func (r *gormRepository) UpdateUserActive(ctx context.Context, id uuid.UUID, active bool) error {
	return r.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		Updates(map[string]any{"active": active, "updated_at": time.Now()}).Error
}

func (r *gormRepository) CountActiveAdmins(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&User{}).
		Joins("JOIN role ON role.id = users.role_id").
		Where("users.active = ? AND role.code = ?", true, access.RoleAdmin).
		Count(&total).Error
	return total, err
}

func (r *gormRepository) LockAdminChanges(ctx context.Context) error {
	return r.lock(ctx, adminChangeLockKey)
}

func (r *gormRepository) lock(ctx context.Context, key int64) error {
	return r.db.WithContext(ctx).Exec("SELECT pg_advisory_xact_lock(?)", key).Error
}

func (r *gormRepository) UpsertRoles(ctx context.Context, roles []Role) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "description", "updated_at"}),
	}).Create(&roles).Error
}

func (r *gormRepository) AssignRoleToLegacyUsers(ctx context.Context, roleCode string) (int64, error) {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE users SET role_id = (SELECT id FROM role WHERE code = ?), updated_at = now()
		  WHERE role_id IS NULL AND EXISTS (SELECT 1 FROM role WHERE code = ?)
		    AND NOT EXISTS (SELECT 1 FROM users AS assigned WHERE assigned.role_id IS NOT NULL)`,
		roleCode, roleCode,
	)
	return result.RowsAffected, result.Error
}
