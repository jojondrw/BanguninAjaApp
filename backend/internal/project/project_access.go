package project

import (
	"context"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/access"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
)

var rolesSeeingEveryProject = []string{access.RoleAdmin, access.RoleObserver, access.RoleFinance}

var (
	errNotProjectMember     = apperror.Forbidden("project_member_required", "Proyek ini hanya bisa dibuka anggota timnya. Minta Admin atau Manajer Proyek di proyek itu menambahkan Anda")
	errProjectViewerUnknown = apperror.Forbidden("project_viewer_unknown", "Peran Anda tidak dikenali, jadi proyek tidak bisa dibuka")
)

func seesEveryProject(role string) bool {
	return slices.Contains(rolesSeeingEveryProject, role)
}

func memberScope(viewerID uuid.UUID, role string) *uuid.UUID {
	if seesEveryProject(role) {
		return nil
	}
	return &viewerID
}

func (s *service) EnsureProjectAccess(ctx context.Context, projectID, viewerID uuid.UUID, role string) error {
	if seesEveryProject(role) {
		return nil
	}
	if role == "" {
		return errProjectViewerUnknown
	}
	member, err := s.repository.IsProjectMember(ctx, projectID, viewerID)
	if err != nil {
		return apperror.Internal(err)
	}
	if !member {
		return errNotProjectMember
	}
	return nil
}

func (c *Controller) RequireProjectAccess(ctx *gin.Context) {
	projectID, err := httprequest.PathID(ctx, projectParam)
	if err != nil {
		_ = ctx.Error(err)
		ctx.Abort()
		return
	}
	viewerID, err := httprequest.UserID(ctx)
	if err != nil {
		_ = ctx.Error(err)
		ctx.Abort()
		return
	}
	role, _ := middleware.AuthenticatedRole(ctx)
	if err := c.service.EnsureProjectAccess(ctx.Request.Context(), projectID, viewerID, role); err != nil {
		_ = ctx.Error(err)
		ctx.Abort()
		return
	}
	ctx.Next()
}

func (r *gormRepository) IsProjectMember(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).
		Table("project_member").
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Limit(1).
		Pluck("id", &ids).Error
	return len(ids) > 0, database.Translate(err)
}
