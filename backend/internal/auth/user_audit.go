package auth

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/entity"
)

const (
	auditRole      = "role"
	auditActive    = "active"
	auditListLimit = 20
)

type UserAudit struct {
	entity.Base
	ActorID  *uuid.UUID `gorm:"type:uuid;index:idx_user_audit_actor"`
	UserID   *uuid.UUID `gorm:"type:uuid;index:idx_user_audit_user"`
	Action   string     `gorm:"type:varchar(20);not null"`
	OldValue string     `gorm:"type:varchar(40);not null;default:''"`
	NewValue string     `gorm:"type:varchar(40);not null;default:''"`
}

func (UserAudit) TableName() string {
	return "user_audit"
}

type UserAuditRow struct {
	ID        uuid.UUID
	ActorName *string
	UserName  *string
	UserEmail *string
	Action    string
	OldValue  string
	NewValue  string
	CreatedAt time.Time
}

type UserAuditResponse struct {
	ID        uuid.UUID `json:"id"`
	ActorName *string   `json:"actorName"`
	UserName  *string   `json:"userName"`
	UserEmail *string   `json:"userEmail"`
	Action    string    `json:"action"`
	OldValue  string    `json:"oldValue"`
	NewValue  string    `json:"newValue"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s *service) ListUserAudit(ctx context.Context) ([]UserAuditResponse, error) {
	rows, err := s.repository.ListUserAudit(ctx, auditListLimit)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	responses := make([]UserAuditResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, UserAuditResponse(row))
	}
	return responses, nil
}

func auditEntries(actorID uuid.UUID, before, after UserRow) []UserAudit {
	var actor *uuid.UUID
	if actorID != uuid.Nil {
		actor = &actorID
	}
	target := before.ID

	entries := []UserAudit{}
	if valueOf(before.RoleCode) != valueOf(after.RoleCode) {
		entries = append(entries, UserAudit{
			ActorID: actor, UserID: &target, Action: auditRole,
			OldValue: valueOf(before.RoleCode), NewValue: valueOf(after.RoleCode),
		})
	}
	if before.Active != after.Active {
		entries = append(entries, UserAudit{
			ActorID: actor, UserID: &target, Action: auditActive,
			OldValue: strconv.FormatBool(before.Active), NewValue: strconv.FormatBool(after.Active),
		})
	}
	return entries
}

func recordAudit(ctx context.Context, repository Repository, actorID uuid.UUID, before, after UserRow) error {
	for _, entry := range auditEntries(actorID, before, after) {
		if err := repository.CreateUserAudit(ctx, &entry); err != nil {
			return err
		}
	}
	return nil
}

func (r *gormRepository) CreateUserAudit(ctx context.Context, entry *UserAudit) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *gormRepository) ListUserAudit(ctx context.Context, limit int) ([]UserAuditRow, error) {
	rows := []UserAuditRow{}
	err := r.db.WithContext(ctx).
		Table("user_audit").
		Select("user_audit.id, actor.name AS actor_name, target.name AS user_name, target.email AS user_email, " +
			"user_audit.action, user_audit.old_value, user_audit.new_value, user_audit.created_at").
		Joins("LEFT JOIN users AS actor ON actor.id = user_audit.actor_id").
		Joins("LEFT JOIN users AS target ON target.id = user_audit.user_id").
		Order("user_audit.created_at DESC, user_audit.id").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}
